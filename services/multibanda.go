package services

import (
	"fmt"
	"strings"
	"time"

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
	PhaseChange(string, *request.MultibandaResume, string) error
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

	if _, err := s.companyRepository.GetById(multibandaRequest.Company); err != nil {
		return "", utils.NewValidationError("company not found")
	}

	if _, err := s.deviceRepository.GetById(multibandaRequest.Device); err != nil {
		return "", utils.NewValidationError("device not found")
	}

	if _, err := s.brandRepository.GetById(multibandaRequest.Brand); err != nil {
		return "", utils.NewValidationError("brand not found")
	}

	multibanda := mapping.MultibandaRequestToMultibanda(
		multibandaRequest,
		companyID,
		deviceID,
		brandID,
	)

	return s.multibandaRepository.Create(multibanda)
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

func (s *multibandaService) PhaseChange(id string, multibandaRequest *request.MultibandaResume, userID string) error {
	if err := s.requireProfileClaim(userID, enums.CanWriteMultibanda); err != nil {
		return err
	}

	multibanda := mapping.MultibandaRequestToMultibandaResume(multibandaRequest)
	setMultibandaDashboardPhase(multibanda)

	if err := s.multibandaRepository.PhaseChange(id, multibanda); err != nil {
		return err
	}

	multibandaID, _ := primitive.ObjectIDFromHex(id)
	multibandaResponse, _ := s.multibandaRepository.GetByIdExpanded(multibandaID)
	if multibandaResponse == nil {
		return nil
	}

	multibandaNotify := mapping.MultibandaResponseToMultibandaNotify(*multibandaResponse)
	multibandaNotify.Status = multibanda.Status
	go s.MultibandaNotification(&multibandaNotify, multibandaResponse.Company.ID, utils.PHASE, userID)

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
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return
	}

	userName := fmt.Sprintf("%s %s", user.Name, user.LastName)
	toList, isEmpty := functions.GetEmails(false, companyID)
	if isEmpty {
		return
	}

	device, err := s.deviceRepository.GetById(multibanda.Device)
	if err != nil {
		return
	}

	brand, err := s.brandRepository.GetById(multibanda.Brand)
	if err != nil {
		return
	}

	projectType := "External"
	if multibanda.IsInternalProject {
		projectType = "Internal"
	}

	planningDate := formatOptionalDate(multibanda.PlanningDate)
	sampleStartDate := formatOptionalDate(multibanda.SampleStartDate)
	sampleEndDate := formatOptionalDate(multibanda.SampleEndDate)
	testStartDate := formatOptionalDate(multibanda.TestStartDate)
	testEndDate := formatOptionalDate(multibanda.TestEndDate)
	underStartDate := formatOptionalDate(multibanda.UnderStartDate)
	underEndDate := formatOptionalDate(multibanda.UnderEndDate)
	resultDate := formatOptionalDate(multibanda.CompletedDate)

	finished := false
	desicion := strings.ToUpper(enums.HomologationStatus_type[multibanda.Status])
	if multibanda.Status != enums.HomologationStatus_value["IN_PROGRESS"] {
		finished = true
	}

	resume := request.MultibandaResume{
		CurrentPhase:    multibanda.CurrentPhase,
		Status:          multibanda.Status,
		SampleEndDate:   multibanda.SampleEndDate,
		TestStartDate:   multibanda.TestStartDate,
	}

	var subject string
	var mainMessage string

	switch key {
	case utils.PHASE:
		mainMessage, subject = functions.GetMultibandaNotificationMessageAndSubject(
			&resume,
			brand.Name,
			device.CommercialModel,
		)
	default:
		return
	}

	body, err := functions.GetHomologationBodyMessage(
		subject,
		mainMessage,
		projectType,
		brand.Name,
		device.TechnicalModel,
		device.CommercialModel,
		multibanda.SoftwareVersion,
		multibanda.OsVersion,
		"Multibanda",
		multibanda.Type,
		"",
		"",
		planningDate,
		sampleStartDate,
		sampleEndDate,
		testStartDate,
		testEndDate,
		underStartDate,
		underEndDate,
		resultDate,
		utils.TEMPLATE_HOMOLOGATION_PATH,
		userName,
		finished,
		desicion,
	)
	if err != nil {
		return
	}

	functions.SendNotifications(toList, body)
}

func formatOptionalDate(date time.Time) string {
	if date.IsZero() {
		return "N/A"
	}
	return fmt.Sprintf("%02d/%02d/%d", date.Day(), date.Month(), date.Year())
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
