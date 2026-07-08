package services



import (

	"bytes"

	"fmt"
	"log"
	"mime/multipart"
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



type IShipmentControlService interface {

	Create(*request.ShipmentControl, string) (string, error)

	Get(string) ([]*responses.ShipmentControlExpanded, error)

	Update(string, *request.ShipmentControl, string) error

	GetAvailableMultibandas(string, string) (*responses.ShipmentControlAvailableResponse, error)

	PhaseChange(string, *request.ShipmentControlResume, string) error

	Delete(string, string) (*responses.DeleteProcessResult, error)

	PatchRequestDelete(string, *request.RequestDeletePatch, string) (*responses.DeleteProcessResult, error)

	RejectRequestDelete(string, string) (*responses.DeleteProcessResult, error)

	ExportShipmentControl(string) (bytes.Buffer, error)

	BulkValidate(*multipart.FileHeader, string, string) (*responses.ShipmentControlBulkValidateResponse, error)

	BulkConfirm(*request.ShipmentControlBulkConfirm, string, string, string) (*responses.ShipmentControlBulkConfirmResponse, error)

}



type shipmentControlService struct {

	shipmentControlRepository repositories.IShipmentControlRepository

	multibandaRepository      repositories.IMultibandaRepository

	deviceRepository          repositories.IDeviceRepository

	userRepository            repositories.IUserRepository

	companyRepository         repositories.ICompanyRepository

	countryRepository         repositories.ICountryRepository

}



func NewShipmentControlService(

	shipmentControlRepository repositories.IShipmentControlRepository,

	multibandaRepository repositories.IMultibandaRepository,

	deviceRepository repositories.IDeviceRepository,

	userRepository repositories.IUserRepository,

	companyRepository repositories.ICompanyRepository,

	countryRepository repositories.ICountryRepository,

) IShipmentControlService {

	return &shipmentControlService{

		shipmentControlRepository: shipmentControlRepository,

		multibandaRepository:      multibandaRepository,

		deviceRepository:          deviceRepository,

		userRepository:            userRepository,

		companyRepository:         companyRepository,

		countryRepository:         countryRepository,

	}

}



func (s *shipmentControlService) Create(req *request.ShipmentControl, userID string) (string, error) {

	if err := s.requireProfileClaim(userID, enums.CanWriteShipmentControl); err != nil {

		return "", err

	}

	if err := utils.ValidateShipmentControlCreateRequest(req); err != nil {

		return "", err

	}

	user, err := s.userRepository.GetByID(userID)

	if err != nil {

		return "", err

	}

	multibandaID, err := primitive.ObjectIDFromHex(req.MultibandaID)

	if err != nil {

		return "", utils.NewValidationError("invalid multibanda_id")

	}

	countryID, err := s.resolveShipmentControlCountry(req, user)

	if err != nil {

		return "", err

	}

	country, err := s.countryRepository.GetById(countryID.Hex())

	if err != nil || country == nil {

		return "", utils.NewValidationError("country not found")

	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(multibandaID)

	if err != nil {

		return "", err

	}

	if multibanda == nil {

		return "", utils.NewValidationError("multibanda not found")

	}

	if multibanda.Status != enums.HomologationStatus_value["APPROVED"] {

		return "", utils.NewValidationError("multibanda must be approved to create shipment control")

	}

	if multibanda.RequestDelete {

		return "", utils.NewValidationError("multibanda has a pending delete request")

	}



	if !functions.UserHasClientAccess(user, multibanda.Company.ID) {

		return "", fmt.Errorf("%w", utils.ErrorForbidden)

	}

	if !user.IsInternal && multibanda.Company.ID != user.Company {

		return "", fmt.Errorf("%w", utils.ErrorForbidden)

	}

	if !userHasBrandAccess(user, multibanda.Brand.ID) {

		return "", fmt.Errorf("%w", utils.ErrorForbidden)

	}



	shipmentControl := mapping.ShipmentControlRequestToShipmentControl(req, multibandaID, multibanda.Company.ID, countryID)

	id, err := s.shipmentControlRepository.Create(shipmentControl)
	if err != nil {
		return "", err
	}

	shipmentControlID, _ := primitive.ObjectIDFromHex(id)
	go s.ShipmentControlNotification(shipmentControlID, nil, utils.CREATE, userID)

	return id, nil

}



func (s *shipmentControlService) Get(userID string) ([]*responses.ShipmentControlExpanded, error) {
	if err := s.requireProfileClaim(userID, enums.CanReadShipmentControl); err != nil {
		return nil, err
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user.IsInternal {
		return s.shipmentControlRepository.GetByInternal(user.Clients, user.Brands)
	}

	return s.shipmentControlRepository.GetByExternal(user.Company, user.Brands)
}



func (s *shipmentControlService) Update(id string, req *request.ShipmentControl, userID string) error {

	if err := s.requireProfileClaim(userID, enums.CanWriteShipmentControl); err != nil {

		return err

	}

	if err := utils.ValidateShipmentControlUpdateRequest(req); err != nil {

		return err

	}

	shipmentControlID, err := primitive.ObjectIDFromHex(id)

	if err != nil {

		return utils.NewValidationError("invalid shipment control id")

	}

	user, err := s.userRepository.GetByID(userID)

	if err != nil {

		return err

	}

	if !user.IsInternal {

		return fmt.Errorf("%w", utils.ErrorForbidden)

	}

	existing, err := s.shipmentControlRepository.GetById(shipmentControlID)

	if err != nil {

		return err

	}

	if existing == nil {

		return utils.NewValidationError("shipment control not found")

	}

	if existing.RequestDelete {

		return utils.NewValidationError("shipment control has a pending delete request")

	}

	multibandaID, err := primitive.ObjectIDFromHex(req.MultibandaID)

	if err != nil {

		return utils.NewValidationError("invalid multibanda_id")

	}

	countryID, err := parseReferenceID("country", req.Country)

	if err != nil {

		return err

	}

	country, err := s.countryRepository.GetById(countryID.Hex())

	if err != nil || country == nil {

		return utils.NewValidationError("country not found")

	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(multibandaID)

	if err != nil {

		return err

	}

	if multibanda == nil {

		return utils.NewValidationError("multibanda not found")

	}

	if multibanda.Status != enums.HomologationStatus_value["APPROVED"] {

		return utils.NewValidationError("multibanda must be approved")

	}

	if multibanda.RequestDelete {

		return utils.NewValidationError("multibanda has a pending delete request")

	}

	if !functions.UserHasClientAccess(user, multibanda.Company.ID) {

		return fmt.Errorf("%w", utils.ErrorForbidden)

	}

	if !userHasBrandAccess(user, multibanda.Brand.ID) {

		return fmt.Errorf("%w", utils.ErrorForbidden)

	}

	exists, err := s.shipmentControlRepository.ExistsByMultibandaExcludingID(shipmentControlID, multibandaID)

	if err != nil {

		return err

	}

	if exists {

		return utils.NewValidationError("a shipment control record already exists for this multibanda")

	}

	shipmentControl := mapping.ShipmentControlRequestToShipmentControlUpdate(

		req,

		multibandaID,

		multibanda.Company.ID,

		countryID,

	)

	return s.shipmentControlRepository.Update(id, shipmentControl)

}



func (s *shipmentControlService) GetAvailableMultibandas(userID, companyParam string) (*responses.ShipmentControlAvailableResponse, error) {

	if err := s.requireProfileClaim(userID, enums.CanReadShipmentControl); err != nil {

		return nil, err

	}



	user, err := s.userRepository.GetByID(userID)

	if err != nil {

		return nil, err

	}



	companyID, err := s.resolveAvailableMultibandasCompany(user, companyParam)

	if err != nil {

		return nil, err

	}



	company, err := s.companyRepository.GetById(companyID.Hex())

	if err != nil {

		return nil, utils.NewValidationError("company not found")

	}



	multibandas, err := s.shipmentControlRepository.GetAvailableMultibandas(companyID, user.Brands)

	if err != nil {

		return nil, err

	}



	responseCompany := responses.Company{

		ID:      company.ID,

		Email:   company.Email,

		Name:    company.Name,

		Address: company.Address,

		LogoUrl: company.LogoUrl,

	}



	return functions.GroupAvailableMultibandas(responseCompany, multibandas), nil

}



func (s *shipmentControlService) PhaseChange(id string, req *request.ShipmentControlResume, userID string) error {

	if err := s.requireProfileClaim(userID, enums.CanWriteShipmentControl); err != nil {

		return err

	}

	if err := utils.ValidateShipmentControlPhaseRequest(req); err != nil {

		return err

	}



	shipmentControlID, err := primitive.ObjectIDFromHex(id)

	if err != nil {

		return utils.NewValidationError("invalid shipment control id")

	}



	user, err := s.userRepository.GetByID(userID)

	if err != nil {

		return err

	}



	existing, err := s.shipmentControlRepository.GetById(shipmentControlID)

	if err != nil {

		return err

	}

	if existing == nil {

		return utils.NewValidationError("shipment control not found")

	}

	if existing.RequestDelete {

		return utils.NewValidationError("shipment control has a pending delete request")

	}



	if !functions.UserHasClientAccess(user, existing.Company) {

		return fmt.Errorf("%w", utils.ErrorForbidden)

	}



	countryID := existing.Country

	if req.Country != "" {

		countryID, err = parseReferenceID("country", req.Country)

		if err != nil {

			return err

		}

		country, err := s.countryRepository.GetById(countryID.Hex())

		if err != nil || country == nil {

			return utils.NewValidationError("country not found")

		}

	}



	shipmentControl := mapping.ShipmentControlRequestToShipmentControlResume(req, countryID)

	functions.ApplyShipmentControlStatusRules(shipmentControl, existing)
	functions.ApplyShipmentControlPhaseDateRules(shipmentControl, existing)

	if err := s.shipmentControlRepository.PhaseChange(id, shipmentControl); err != nil {
		return err
	}

	go s.ShipmentControlNotification(shipmentControlID, existing, utils.PHASE, userID)

	return nil

}

func (s *shipmentControlService) Delete(id string, userID string) (*responses.DeleteProcessResult, error) {
	if err := s.requireProfileClaim(userID, enums.CanDeleteShipmentControl); err != nil {
		return nil, err
	}

	shipmentControlID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, utils.NewValidationError("invalid shipment control id")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	existing, err := s.shipmentControlRepository.GetById(shipmentControlID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, utils.NewValidationError("shipment control not found")
	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(existing.Multibanda)
	if err != nil {
		return nil, err
	}
	if multibanda == nil {
		return nil, utils.NewValidationError("multibanda not found")
	}

	if err := authorizeMultibandaRecordAccess(user, existing.Company, multibanda.Brand.ID); err != nil {
		return nil, err
	}

	if !user.IsInternal {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	if err := s.shipmentControlRepository.Delete(shipmentControlID); err != nil {
		return nil, err
	}

	company, err := s.companyRepository.GetById(existing.Company.Hex())
	if err != nil || company == nil {
		return &responses.DeleteProcessResult{Deleted: true}, nil
	}

	countryName := ""
	country, err := s.countryRepository.GetById(existing.Country.Hex())
	if err == nil && country != nil {
		countryName = country.Name
	}

	notify := mapping.ShipmentControlToNotify(existing, multibanda, company.Name, countryName)
	go s.shipmentControlDeletedNotification(&notify, multibanda, existing.Company, userID)

	return &responses.DeleteProcessResult{Deleted: true}, nil
}

func (s *shipmentControlService) PatchRequestDelete(id string, body *request.RequestDeletePatch, userID string) (*responses.DeleteProcessResult, error) {
	if err := s.requireProfileClaim(userID, enums.CanDeleteShipmentControl); err != nil {
		return nil, err
	}

	shipmentControlID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, utils.NewValidationError("invalid shipment control id")
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

	existing, err := s.shipmentControlRepository.GetById(shipmentControlID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, utils.NewValidationError("shipment control not found")
	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(existing.Multibanda)
	if err != nil {
		return nil, err
	}
	if multibanda == nil {
		return nil, utils.NewValidationError("multibanda not found")
	}

	if err := authorizeMultibandaRecordAccess(user, existing.Company, multibanda.Brand.ID); err != nil {
		return nil, err
	}

	if existing.RequestDelete {
		return nil, utils.NewValidationError("delete already requested")
	}

	if err := s.shipmentControlRepository.SetRequestDelete(shipmentControlID, true); err != nil {
		return nil, err
	}

	go s.shipmentControlRequestDeleteNotification(shipmentControlID, userID)

	return &responses.DeleteProcessResult{RequestDelete: true}, nil
}

func (s *shipmentControlService) RejectRequestDelete(id string, userID string) (*responses.DeleteProcessResult, error) {
	if err := s.requireProfileClaim(userID, enums.CanDeleteShipmentControl); err != nil {
		return nil, err
	}

	shipmentControlID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, utils.NewValidationError("invalid shipment control id")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if !user.IsInternal {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	existing, err := s.shipmentControlRepository.GetById(shipmentControlID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, utils.NewValidationError("shipment control not found")
	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(existing.Multibanda)
	if err != nil {
		return nil, err
	}
	if multibanda == nil {
		return nil, utils.NewValidationError("multibanda not found")
	}

	if err := authorizeMultibandaRecordAccess(user, existing.Company, multibanda.Brand.ID); err != nil {
		return nil, err
	}

	if !existing.RequestDelete {
		return nil, utils.NewValidationError("no pending delete request")
	}

	if err := s.shipmentControlRepository.SetRequestDelete(shipmentControlID, false); err != nil {
		return nil, err
	}
	return &responses.DeleteProcessResult{RequestDelete: false}, nil
}

func (s *shipmentControlService) ShipmentControlNotification(
	shipmentControlID primitive.ObjectID,
	existing *models.ShipmentControl,
	notifyKey string,
	userID string,
) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return
	}

	shipment, err := s.shipmentControlRepository.GetById(shipmentControlID)
	if err != nil || shipment == nil {
		return
	}

	toList, isEmpty := functions.GetEmails(false, shipment.Company)
	if isEmpty {
		return
	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(shipment.Multibanda)
	if err != nil || multibanda == nil {
		return
	}

	company, err := s.companyRepository.GetById(shipment.Company.Hex())
	if err != nil || company == nil {
		return
	}

	countryName := ""
	country, err := s.countryRepository.GetById(shipment.Country.Hex())
	if err == nil && country != nil {
		countryName = country.Name
	}

	notify := mapping.ShipmentControlToNotify(shipment, multibanda, company.Name, countryName)

	var existingNotify *request.ShipmentControlNotify
	if existing != nil {
		en := mapping.ShipmentControlToNotify(existing, multibanda, company.Name, countryName)
		existingNotify = &en
	}

	emailKind := functions.ResolveShipmentControlPhaseEmailKind(notifyKey, existingNotify, &notify)
	if emailKind == "" {
		return
	}

	userName := fmt.Sprintf("%s %s", user.Name, user.LastName)
	mainMessage, subject := functions.GetShipmentControlNotificationMessageAndSubject(
		emailKind,
		multibanda.Brand.Name,
		multibanda.Device.CommercialModel,
		notify.SoftwareVersion,
	)
	if subject == "" {
		return
	}

	emailData := functions.BuildShipmentControlPhaseEmailData(
		&notify,
		multibanda.Brand.Name,
		multibanda.Device.TechnicalModel,
		multibanda.Device.CommercialModel,
		multibanda.Device.PlatformOs,
		userName,
		mainMessage,
		emailKind,
	)

	if err := functions.SendShipmentControlPhaseEmail(
		toList,
		subject,
		emailData,
		utils.TEMPLATE_SHIPMENT_CONTROL_PHASE_PATH,
		utils.LBOneTrackLogoPNG,
	); err != nil {
		log.Printf("shipment control notification email (%s): %v", notifyKey, err)
	}
}

func (s *shipmentControlService) shipmentControlRequestDeleteNotification(
	shipmentControlID primitive.ObjectID,
	userID string,
) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return
	}

	shipment, multibanda, companyName, countryName, ok := s.loadShipmentControlEmailContext(shipmentControlID)
	if !ok {
		return
	}

	notify := mapping.ShipmentControlToNotify(shipment, multibanda, companyName, countryName)
	userName := fmt.Sprintf("%s %s", user.Name, user.LastName)

	internalList, internalEmpty := functions.GetEmails(true, shipment.Company)
	if !internalEmpty {
		s.sendShipmentControlDeleteEmail(
			&notify,
			multibanda,
			internalList,
			functions.ShipmentControlEmailRequestDeleteInternal,
			"LB Technology Team",
			userName,
			"request delete internal",
		)
	}

	clientList, clientEmpty := functions.GetEmails(false, shipment.Company)
	if !clientEmpty {
		dearName := strings.TrimSpace(companyName)
		if dearName == "" {
			dearName = "Client"
		}
		s.sendShipmentControlDeleteEmail(
			&notify,
			multibanda,
			clientList,
			functions.ShipmentControlEmailRequestDeleteClient,
			dearName,
			userName,
			"request delete client",
		)
	}
}

func (s *shipmentControlService) shipmentControlDeletedNotification(
	notify *request.ShipmentControlNotify,
	multibanda *responses.MultibandaExpanded,
	companyID primitive.ObjectID,
	userID string,
) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return
	}

	clientList, clientEmpty := functions.GetEmails(false, companyID)
	if clientEmpty {
		return
	}

	userName := fmt.Sprintf("%s %s", user.Name, user.LastName)
	dearName := strings.TrimSpace(notify.CompanyName)
	if dearName == "" {
		dearName = "Client"
	}

	s.sendShipmentControlDeleteEmail(
		notify,
		multibanda,
		clientList,
		functions.ShipmentControlEmailDeleted,
		dearName,
		userName,
		"deleted",
	)
}

func (s *shipmentControlService) loadShipmentControlEmailContext(
	shipmentControlID primitive.ObjectID,
) (*models.ShipmentControl, *responses.MultibandaExpanded, string, string, bool) {
	shipment, err := s.shipmentControlRepository.GetById(shipmentControlID)
	if err != nil || shipment == nil {
		return nil, nil, "", "", false
	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(shipment.Multibanda)
	if err != nil || multibanda == nil {
		return nil, nil, "", "", false
	}

	company, err := s.companyRepository.GetById(shipment.Company.Hex())
	if err != nil || company == nil {
		return nil, nil, "", "", false
	}

	countryName := ""
	country, err := s.countryRepository.GetById(shipment.Country.Hex())
	if err == nil && country != nil {
		countryName = country.Name
	}

	return shipment, multibanda, company.Name, countryName, true
}

func (s *shipmentControlService) sendShipmentControlDeleteEmail(
	notify *request.ShipmentControlNotify,
	multibanda *responses.MultibandaExpanded,
	toList []string,
	emailKind string,
	dearName string,
	userName string,
	logKey string,
) {
	mainMessage, subject := functions.GetShipmentControlDeleteNotificationMessageAndSubject(
		emailKind,
		multibanda.Brand.Name,
		multibanda.Device.CommercialModel,
		notify.SoftwareVersion,
		notify.CompanyName,
	)
	if subject == "" {
		return
	}

	emailData := functions.BuildShipmentControlPhaseEmailData(
		notify,
		multibanda.Brand.Name,
		multibanda.Device.TechnicalModel,
		multibanda.Device.CommercialModel,
		multibanda.Device.PlatformOs,
		userName,
		mainMessage,
		emailKind,
	)
	emailData.ClientName = dearName

	if err := functions.SendShipmentControlPhaseEmail(
		toList,
		subject,
		emailData,
		utils.TEMPLATE_SHIPMENT_CONTROL_PHASE_PATH,
		utils.LBOneTrackLogoPNG,
	); err != nil {
		log.Printf("shipment control notification email (%s): %v", logKey, err)
	}
}



func (s *shipmentControlService) resolveShipmentControlCountry(
	req *request.ShipmentControl,
	user *responses.User,
) (primitive.ObjectID, error) {
	if user.IsInternal {
		if strings.TrimSpace(req.Country) == "" {
			return primitive.NilObjectID, utils.NewValidationError("country is required")
		}
		return parseReferenceID("country", req.Country)
	}

	if user.Company == primitive.NilObjectID {
		return primitive.NilObjectID, utils.NewValidationError("user company is not configured")
	}

	chile, err := s.countryRepository.GetByName(enums.ShipmentControlExternalCountryName)
	if err != nil || chile == nil {
		return primitive.NilObjectID, utils.NewValidationError("Chile country not found")
	}

	return chile.ID, nil
}

func (s *shipmentControlService) resolveAvailableMultibandasCompany(user *responses.User, companyParam string) (primitive.ObjectID, error) {
	if user.IsInternal {
		if companyParam == "" {
			return primitive.NilObjectID, utils.NewValidationError("company query parameter is required")
		}
		return parseReferenceID("company", companyParam)
	}

	if user.Company == primitive.NilObjectID {
		return primitive.NilObjectID, utils.NewValidationError("user company is not configured")
	}
	return user.Company, nil
}

func (s *shipmentControlService) resolveCompanyScope(user *responses.User, companyParam string) (primitive.ObjectID, error) {

	if user.IsInternal {

		if companyParam == "" {

			return primitive.NilObjectID, utils.NewValidationError("company query parameter is required")

		}

		companyID, err := parseReferenceID("company", companyParam)

		if err != nil {

			return primitive.NilObjectID, err

		}

		if !functions.UserHasClientAccess(user, companyID) {

			return primitive.NilObjectID, utils.NewValidationError("company is not assigned to user")

		}

		return companyID, nil

	}



	if user.Company == primitive.NilObjectID {

		return primitive.NilObjectID, utils.NewValidationError("user company is not configured")

	}

	return user.Company, nil

}



func (s *shipmentControlService) requireProfileClaim(userID, claimName string) error {

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

		return utils.NewValidationError(fmt.Sprintf("missing profile claim %s", claimName))

	}



	return nil

}


