package services



import (

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



type IShipmentControlService interface {

	Create(*request.ShipmentControl, string) (string, error)

	Get(string) ([]*responses.ShipmentControlExpanded, error)

	GetAvailableMultibandas(string, string) (*responses.ShipmentControlAvailableResponse, error)

	PhaseChange(string, *request.ShipmentControlResume, string) error

}



type shipmentControlService struct {

	shipmentControlRepository repositories.IShipmentControlRepository

	multibandaRepository      repositories.IMultibandaRepository

	userRepository            repositories.IUserRepository

	companyRepository         repositories.ICompanyRepository

	countryRepository         repositories.ICountryRepository

}



func NewShipmentControlService(

	shipmentControlRepository repositories.IShipmentControlRepository,

	multibandaRepository repositories.IMultibandaRepository,

	userRepository repositories.IUserRepository,

	companyRepository repositories.ICompanyRepository,

	countryRepository repositories.ICountryRepository,

) IShipmentControlService {

	return &shipmentControlService{

		shipmentControlRepository: shipmentControlRepository,

		multibandaRepository:      multibandaRepository,

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
		shipment.ReworkNumber,
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


