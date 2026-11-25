package services

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IMultibandaService interface {
	Create(*request.Multibanda, string) (string, error)
	Get(string) ([]*responses.MultibandaExpanded, error)
	Update(string, *request.Multibanda, string) error
	PhaseChange(string, *request.MultibandaResume, string) error
	Delete(string, string) (*responses.DeleteProcessResult, error)
	PatchRequestDelete(string, *request.RequestDeletePatch, string) (*responses.DeleteProcessResult, error)
	RejectRequestDelete(string, string) (*responses.DeleteProcessResult, error)
	ExportMultibanda(string) (bytes.Buffer, error)
	NotifyReportGenerated(primitive.ObjectID, string, functions.MultibandaReportEmailDetails)
}

type multibandaService struct {
	multibandaRepository repositories.IMultibandaRepository
	userRepository       repositories.IUserRepository
	companyRepository    repositories.ICompanyRepository
	deviceRepository     repositories.IDeviceRepository
	brandRepository      repositories.IBrandRepository
}

func NewMultibandaService(
	multibandaRepository repositories.IMultibandaRepository,
	userRepository repositories.IUserRepository,
	companyRepository repositories.ICompanyRepository,
	deviceRepository repositories.IDeviceRepository,
	brandRepository repositories.IBrandRepository,
) IMultibandaService {
	return &multibandaService{
		multibandaRepository: multibandaRepository,
		userRepository:       userRepository,
		companyRepository:    companyRepository,
		deviceRepository:     deviceRepository,
		brandRepository:      brandRepository,
	}
}

func (s *multibandaService) Create(multibandaRequest *request.Multibanda, userID string) (string, error) {
	if err := s.requireProfileClaim(userID, enums.CanWriteMultibanda); err != nil {
		return "", err
	}

	if err := utils.ValidateMultibandaCreateRequest(multibandaRequest); err != nil {
		return "", err
	}

	companyID, err := parseReferenceID("company", multibandaRequest.Company)
	if err != nil {
		return "", err
	}

	deviceID, err := parseReferenceID("device", multibandaRequest.Device)
	if err != nil {
		return "", err
	}

	brandID, err := parseReferenceID("brand", multibandaRequest.Brand)
	if err != nil {
		return "", err
	}

	company, err := s.companyRepository.GetById(multibandaRequest.Company)
	if err != nil || company == nil {
		return "", utils.NewValidationError("company not found")
	}

	if _, err := s.deviceRepository.GetById(multibandaRequest.Device); err != nil {
		return "", utils.NewValidationError("device not found")
	}

	if _, err := s.brandRepository.GetById(multibandaRequest.Brand); err != nil {
		return "", utils.NewValidationError("brand not found")
	}

	requiresTesting, err := functions.NormalizeRequiresTesting(
		functions.RequiresTestingApplies(multibandaRequest.Type, multibandaRequest.EvaluationType, company.AllowSmrWithoutTesting),
		multibandaRequest.RequiresTesting,
	)
	if err != nil {
		return "", err
	}

	exists, err := s.multibandaRepository.ExistsByCompanyDeviceSoftwareOsVersion(
		companyID,
		deviceID,
		strings.TrimSpace(multibandaRequest.SoftwareVersion),
		strings.TrimSpace(multibandaRequest.OsVersion),
	)
	if err != nil {
		return "", err
	}
	if exists {
		return "", utils.NewValidationError("a multibanda record already exists for this company, device, software_version and os_version")
	}

	multibanda := mapping.MultibandaRequestToMultibanda(
		multibandaRequest,
		companyID,
		deviceID,
		brandID,
	)
	if creatorID, err := primitive.ObjectIDFromHex(userID); err == nil {
		multibanda.CreatedBy = creatorID
	}
	multibanda.RequiresTesting = requiresTesting
	multibanda.PhaseStates = functions.ComputePhaseStates(multibanda.CurrentPhase, nil)
	withoutTesting := functions.IsSmrWithoutTesting(requiresTesting)
	if withoutTesting {
		multibanda.PhaseStates = functions.SmrWithoutTestingInitialStates()
		multibanda.SmrValidation = &models.SmrValidation{Status: models.SmrValidationPending}
	}

	id, err := s.multibandaRepository.Create(multibanda)
	if err != nil {
		return "", err
	}

	multibandaID, _ := primitive.ObjectIDFromHex(id)
	multibandaResponse, _ := s.multibandaRepository.GetByIdExpanded(multibandaID)
	if multibandaResponse != nil {
		multibandaNotify := mapping.MultibandaResponseToMultibandaNotify(*multibandaResponse)
		go s.MultibandaNotification(&multibandaNotify, multibandaResponse.Company.ID, utils.CREATE, userID)
		if withoutTesting {
			go s.NotifySmrValidationPending(multibandaResponse)
		}
	}

	return id, nil
}

func (s *multibandaService) Get(userID string) ([]*responses.MultibandaExpanded, error) {
	if err := s.requireProfileClaim(userID, enums.CanReadMultibanda); err != nil {
		return nil, err
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user.IsInternal {
		return s.multibandaRepository.GetByInternal(user.Clients, user.Brands)
	}

	return s.multibandaRepository.GetByExternal(user.Company, user.Brands)
}

func (s *multibandaService) Update(id string, multibandaRequest *request.Multibanda, userID string) error {
	if err := s.requireProfileClaim(userID, enums.CanWriteMultibanda); err != nil {
		return err
	}

	if err := utils.ValidateMultibandaUpdateRequest(multibandaRequest); err != nil {
		return err
	}

	multibandaID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.NewValidationError("invalid multibanda id")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return err
	}

	if !user.IsInternal {
		return fmt.Errorf("%w", utils.ErrorForbidden)
	}

	existing, err := s.multibandaRepository.GetByIdExpanded(multibandaID)
	if err != nil {
		return err
	}
	if existing == nil {
		return utils.NewValidationError("multibanda not found")
	}
	if existing.RequestDelete {
		return utils.NewValidationError("multibanda has a pending delete request")
	}

	companyID, err := parseReferenceID("company", multibandaRequest.Company)
	if err != nil {
		return err
	}

	deviceID, err := parseReferenceID("device", multibandaRequest.Device)
	if err != nil {
		return err
	}

	brandID, err := parseReferenceID("brand", multibandaRequest.Brand)
	if err != nil {
		return err
	}

	company, err := s.companyRepository.GetById(multibandaRequest.Company)
	if err != nil || company == nil {
		return utils.NewValidationError("company not found")
	}

	if _, err := s.deviceRepository.GetById(multibandaRequest.Device); err != nil {
		return utils.NewValidationError("device not found")
	}

	if _, err := s.brandRepository.GetById(multibandaRequest.Brand); err != nil {
		return utils.NewValidationError("brand not found")
	}

	if err := authorizeMultibandaRecordAccess(user, companyID, brandID); err != nil {
		return err
	}

	// SMR without testing: guards and requires_testing.
	validationStatus := smrValidationStatus(existing)
	if err := functions.GuardSmrChange(validationStatus, existing.CurrentPhase, existing.Status,
		multibandaRequest.CurrentPhase, multibandaRequest.Status); err != nil {
		return err
	}
	// On edit, null keeps what is stored (legacy SMRs created before the flag
	// stay editable); only a sent value is validated.
	requiresTesting := existing.RequiresTesting
	if multibandaRequest.RequiresTesting != nil {
		requiresTesting, err = functions.NormalizeRequiresTesting(
			functions.RequiresTestingApplies(multibandaRequest.Type, multibandaRequest.EvaluationType, company.AllowSmrWithoutTesting),
			multibandaRequest.RequiresTesting,
		)
		if err != nil {
			return err
		}
	}
	decided := validationStatus == models.SmrValidationApproved || validationStatus == models.SmrValidationRejected
	if decided && !sameBoolPtr(requiresTesting, existing.RequiresTesting) {
		return utils.NewValidationError(functions.SmrErrRequiresTestingLocked)
	}
	if functions.IsSmrWithoutTesting(requiresTesting) && !functions.IsSmrWithoutTesting(existing.RequiresTesting) &&
		existing.CurrentPhase != enums.HomologationPhase_value["PLANNING"] {
		return utils.NewValidationError(functions.SmrErrOnlyInPlanning)
	}

	exists, err := s.multibandaRepository.ExistsByCompanyDeviceSoftwareOsVersionExcludingID(
		multibandaID,
		companyID,
		deviceID,
		strings.TrimSpace(multibandaRequest.SoftwareVersion),
		strings.TrimSpace(multibandaRequest.OsVersion),
	)
	if err != nil {
		return err
	}
	if exists {
		return utils.NewValidationError("a multibanda record already exists for this company, device, software_version and os_version")
	}

	multibanda := mapping.MultibandaRequestToMultibandaUpdate(
		multibandaRequest,
		companyID,
		deviceID,
		brandID,
	)
	mapping.PreserveMultibandaTestReportURL(multibanda, existing.TestReportUrl)
	setMultibandaDashboardPhase(multibanda)

	multibanda.RequiresTesting = requiresTesting
	wasWithoutTesting := functions.IsSmrWithoutTesting(existing.RequiresTesting)
	nowWithoutTesting := functions.IsSmrWithoutTesting(requiresTesting)
	switch {
	case nowWithoutTesting && !wasWithoutTesting:
		multibanda.PhaseStates = functions.SmrWithoutTestingInitialStates()
	case !nowWithoutTesting && wasWithoutTesting:
		multibanda.PhaseStates = functions.ComputePhaseStates(multibanda.CurrentPhase, nil)
	default:
		multibanda.PhaseStates = functions.ComputePhaseStates(multibanda.CurrentPhase, existing.PhaseStates)
	}

	if err := s.multibandaRepository.Update(id, multibanda); err != nil {
		return err
	}

	// Switching requires_testing before a decision starts or drops the
	// pending validation.
	switch {
	case nowWithoutTesting && !wasWithoutTesting:
		if err := s.multibandaRepository.SetSmrValidation(multibandaID,
			&models.SmrValidation{Status: models.SmrValidationPending}); err != nil {
			return err
		}
		if updated, _ := s.multibandaRepository.GetByIdExpanded(multibandaID); updated != nil {
			go s.NotifySmrValidationPending(updated)
		}
	case !nowWithoutTesting && wasWithoutTesting:
		if err := s.multibandaRepository.SetSmrValidation(multibandaID, nil); err != nil {
			return err
		}
	}
	return nil
}

func smrValidationStatus(m *responses.MultibandaExpanded) string {
	if m == nil || m.SmrValidation == nil {
		return ""
	}
	return m.SmrValidation.Status
}

func sameBoolPtr(a, b *bool) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func (s *multibandaService) PhaseChange(id string, multibandaRequest *request.MultibandaResume, userID string) error {
	if err := s.requireProfileClaim(userID, enums.CanWriteMultibanda); err != nil {
		return err
	}

	multibandaID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.NewValidationError("invalid multibanda id")
	}

	existing, _ := s.multibandaRepository.GetByIdExpanded(multibandaID)
	if existing == nil {
		return utils.NewValidationError("multibanda not found")
	}
	if existing.RequestDelete {
		return utils.NewValidationError("multibanda has a pending delete request")
	}

	if err := functions.GuardSmrChange(smrValidationStatus(existing), existing.CurrentPhase, existing.Status,
		multibandaRequest.CurrentPhase, multibandaRequest.Status); err != nil {
		return err
	}

	multibanda := mapping.MultibandaRequestToMultibandaResume(multibandaRequest)
	functions.ApplyMultibandaPhaseDateRules(multibanda, existing)
	mapping.PreserveMultibandaTestReportURL(multibanda, existing.TestReportUrl)
	setMultibandaDashboardPhase(multibanda)
	multibanda.PhaseStates = functions.ComputePhaseStates(multibanda.CurrentPhase, existing.PhaseStates)

	if err := s.multibandaRepository.PhaseChange(id, multibanda); err != nil {
		return err
	}

	multibandaResponse, _ := s.multibandaRepository.GetByIdExpanded(multibandaID)
	if multibandaResponse == nil {
		return nil
	}

	multibandaNotify := mapping.MultibandaResponseToMultibandaNotify(*multibandaResponse)
	multibandaNotify.Status = multibanda.Status
	go s.MultibandaNotification(&multibandaNotify, multibandaResponse.Company.ID, utils.PHASE, userID)

	return nil
}

// NotifyReportGenerated tells the internal contacts that the test report can be
// downloaded. The process is reloaded so the email carries the report URL that
// Generate has just linked. details are the stored report values (FM Radio,
// stamp) so the email matches the PDF.
func (s *multibandaService) NotifyReportGenerated(
	multibandaID primitive.ObjectID,
	userID string,
	details functions.MultibandaReportEmailDetails,
) {
	go func() {
		multibandaResponse, err := s.multibandaRepository.GetByIdExpanded(multibandaID)
		if multibandaResponse == nil {
			log.Printf("multibanda notification (%s) skipped: multibanda %s not loaded: %v",
				utils.REPORT, multibandaID.Hex(), err)
			return
		}

		multibandaNotify := mapping.MultibandaResponseToMultibandaNotify(*multibandaResponse)
		s.notify(&multibandaNotify, multibandaResponse.Company.ID, utils.REPORT, userID, &details)
	}()
}

func (s *multibandaService) Delete(id string, userID string) (*responses.DeleteProcessResult, error) {
	if err := s.requireProfileClaim(userID, enums.CanDeleteMultibanda); err != nil {
		return nil, err
	}

	multibandaID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, utils.NewValidationError("invalid multibanda id")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	existing, err := s.multibandaRepository.GetByIdExpanded(multibandaID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, utils.NewValidationError("multibanda not found")
	}

	if err := authorizeMultibandaRecordAccess(user, existing.Company.ID, existing.Brand.ID); err != nil {
		return nil, err
	}

	if !user.IsInternal {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	if err := s.multibandaRepository.Delete(multibandaID); err != nil {
		return nil, err
	}
	return &responses.DeleteProcessResult{Deleted: true}, nil
}

func (s *multibandaService) PatchRequestDelete(id string, body *request.RequestDeletePatch, userID string) (*responses.DeleteProcessResult, error) {
	if err := s.requireProfileClaim(userID, enums.CanDeleteMultibanda); err != nil {
		return nil, err
	}

	multibandaID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, utils.NewValidationError("invalid multibanda id")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user.IsInternal {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	if err := utils.ValidateRequestDeletePatch(body); err != nil {
		return nil, err
	}

	existing, err := s.multibandaRepository.GetByIdExpanded(multibandaID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, utils.NewValidationError("multibanda not found")
	}

	if err := authorizeMultibandaRecordAccess(user, existing.Company.ID, existing.Brand.ID); err != nil {
		return nil, err
	}

	if existing.RequestDelete {
		return nil, utils.NewValidationError("delete already requested")
	}

	if err := s.multibandaRepository.SetRequestDelete(multibandaID, true); err != nil {
		return nil, err
	}
	return &responses.DeleteProcessResult{RequestDelete: true}, nil
}

func (s *multibandaService) RejectRequestDelete(id string, userID string) (*responses.DeleteProcessResult, error) {
	if err := s.requireProfileClaim(userID, enums.CanDeleteMultibanda); err != nil {
		return nil, err
	}

	multibandaID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, utils.NewValidationError("invalid multibanda id")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if !user.IsInternal {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	existing, err := s.multibandaRepository.GetByIdExpanded(multibandaID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, utils.NewValidationError("multibanda not found")
	}

	if err := authorizeMultibandaRecordAccess(user, existing.Company.ID, existing.Brand.ID); err != nil {
		return nil, err
	}

	if !existing.RequestDelete {
		return nil, utils.NewValidationError("no pending delete request")
	}

	if err := s.multibandaRepository.SetRequestDelete(multibandaID, false); err != nil {
		return nil, err
	}
	return &responses.DeleteProcessResult{RequestDelete: false}, nil
}

func authorizeMultibandaRecordAccess(user *responses.User, companyID, brandID primitive.ObjectID) error {
	if !functions.UserHasClientAccess(user, companyID) {
		return fmt.Errorf("%w", utils.ErrorForbidden)
	}
	if !user.IsInternal && user.Company != companyID {
		return fmt.Errorf("%w", utils.ErrorForbidden)
	}
	if !userHasBrandAccess(user, brandID) {
		return fmt.Errorf("%w", utils.ErrorForbidden)
	}
	return nil
}

func setMultibandaDashboardPhase(multibanda *models.Multibanda) {
	if multibanda.CurrentPhase == enums.HomologationPhase_value["PLANNING"] {
		multibanda.DashBoardPhase = multibanda.CurrentPhase
		return
	}

	if multibanda.CurrentPhase == enums.HomologationPhase_value["SAMPLE_RECEPTION"] {
		if multibanda.SampleEndDate == nil || multibanda.SampleEndDate.IsZero() {
			multibanda.DashBoardPhase = enums.DashBoardPhase_value["PLANNING"]
		} else {
			multibanda.DashBoardPhase = multibanda.CurrentPhase
		}
		return
	}

	if multibanda.CurrentPhase == enums.HomologationPhase_value["COMPLETE"] {
		multibanda.DashBoardPhase = enums.DashBoardPhase_value["COMPLETE"]
		return
	}

	multibanda.DashBoardPhase = enums.DashBoardPhase_value["ONGOING"]
}

func (s *multibandaService) MultibandaNotification(
	multibanda *request.MultibandaNotify,
	companyID primitive.ObjectID,
	key string,
	userID string,
) {
	s.notify(multibanda, companyID, key, userID, nil)
}

// notify sends a multibanda email; reportDetails is set only for the report
// email and adds its FM Radio and Stamp Type rows.
func (s *multibandaService) notify(
	multibanda *request.MultibandaNotify,
	companyID primitive.ObjectID,
	key string,
	userID string,
	reportDetails *functions.MultibandaReportEmailDetails,
) {
	// Every early return is logged: a skipped email must be visible.
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		log.Printf("multibanda notification (%s) skipped: user %s not loaded: %v", key, userID, err)
		return
	}

	userName := fmt.Sprintf("%s %s", user.Name, user.LastName)
	includeExternal := functions.MultibandaNotifiesExternalRecipients(key, multibanda.CurrentPhase)
	toList, isEmpty := functions.GetNotificationEmails(companyID, includeExternal)
	if isEmpty {
		log.Printf("multibanda notification (%s) skipped: no recipients configured for company %s "+
			"(Notifications; include_company_contacts=%v)", key, companyID.Hex(), includeExternal)
		return
	}

	device, err := s.deviceRepository.GetById(multibanda.Device)
	if err != nil {
		log.Printf("multibanda notification (%s) skipped: device %s not loaded: %v", key, multibanda.Device, err)
		return
	}

	brand, err := s.brandRepository.GetById(multibanda.Brand)
	if err != nil {
		log.Printf("multibanda notification (%s) skipped: brand %s not loaded: %v", key, multibanda.Brand, err)
		return
	}

	finished := false
	desicion := strings.ToUpper(enums.HomologationStatus_type[multibanda.Status])
	if multibanda.Status != enums.HomologationStatus_value["IN_PROGRESS"] {
		finished = true
	}

	resume := request.MultibandaResume{
		CurrentPhase:  multibanda.CurrentPhase,
		Status:        multibanda.Status,
		SampleEndDate: multibanda.SampleEndDate,
		TestStartDate: multibanda.TestStartDate,
	}

	var subject string
	var mainMessage string

	switch key {
	case utils.CREATE:
		subject = fmt.Sprintf("Subject: New Multibanda process was created %s %s",
			brand.Name, device.CommercialModel)
		mainMessage = utils.MULTIBANDA_CREATE_MAIN_MESSAGE
	case utils.PHASE:
		mainMessage, subject = functions.GetMultibandaNotificationMessageAndSubject(
			&resume,
			brand.Name,
			device.CommercialModel,
		)
	case utils.REPORT:
		subject = fmt.Sprintf("Subject: Test Report Available Multibanda %s %s",
			brand.Name, device.CommercialModel)
		mainMessage = utils.MULTIBANDA_REPORT_MAIN_MESSAGE
	default:
		return
	}

	emailData := functions.BuildMultibandaPhaseEmailData(
		multibanda,
		brand.Name,
		device.TechnicalModel,
		device.CommercialModel,
		device.PlatformOs,
		device.SarValue,
		userName,
		mainMessage,
		finished,
		desicion,
		key,
	)
	if key == utils.REPORT && reportDetails != nil {
		functions.ApplyMultibandaReportEmailDetails(&emailData, *reportDetails)
	}

	if err := functions.SendMultibandaPhaseEmail(
		toList,
		subject,
		emailData,
		utils.TEMPLATE_MULTIBANDA_PHASE_PATH,
		utils.LBOneTrackLogoPNG,
	); err != nil {
		log.Printf("multibanda notification email (%s): %v", key, err)
	}
}

func (s *multibandaService) requireProfileClaim(userID, claimName string) error {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return err
	}

	profile, err := s.userRepository.GetProfileByID(user.ID)
	if err != nil {
		return err
	}

	modelClaims := make([]models.Claim, 0, len(profile.Claims))
	for _, claim := range profile.Claims {
		modelClaims = append(modelClaims, models.Claim{
			Name:  claim.Name,
			Allow: claim.Allow,
		})
	}

	if !functions.HasProfileClaim(modelClaims, claimName) {
		return fmt.Errorf("%w", utils.ErrorForbidden)
	}

	return nil
}

func parseReferenceID(fieldName, value string) (primitive.ObjectID, error) {
	if _, err := utils.ValidateObjectIDField(fieldName, value); err != nil {
		return primitive.NilObjectID, err
	}

	oid, err := primitive.ObjectIDFromHex(value)
	if err != nil {
		return primitive.NilObjectID, utils.NewValidationError(fmt.Sprintf("invalid %s id", fieldName))
	}

	return oid, nil
}
