package services

import (
	"strings"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ICompanyService is the company service
type ICompanyService interface {
	Create(*request.Company) error
	Get() ([]*responses.CompanyListItem, error)
	Delete(string) (bool, error)
	Update(string, *request.Company) error
	ClientIDAvailable(clientID, excludeID string) (*responses.CompanyClientIDAvailability, error)
	GetUsers(string) ([]responses.CompanyUser, error)
}

type companyService struct {
	companyRepository        repositories.ICompanyRepository
	homologationRepository   repositories.IHomologationRepository
	deviceTrackingRepository repositories.IDeviceTrackingRepository
	userRepository           repositories.IUserRepository
}

// NewCompanyService is a constructor
func NewCompanyService(companyRepository repositories.ICompanyRepository,
	homologationRepository repositories.IHomologationRepository,
	deviceTrackingRepository repositories.IDeviceTrackingRepository,
	userRepository repositories.IUserRepository) ICompanyService {
	return &companyService{
		companyRepository:        companyRepository,
		homologationRepository:   homologationRepository,
		deviceTrackingRepository: deviceTrackingRepository,
		userRepository:           userRepository,
	}
}

// Create validates and stores a new company.
func (s *companyService) Create(companyRequest *request.Company) error {
	if err := s.validate(companyRequest, primitive.NilObjectID); err != nil {
		return err
	}
	company, err := mapping.CompanyRequestToCompany(companyRequest)
	if err != nil {
		return err
	}
	return s.companyRepository.Create(company)
}

// Get lists the companies with the Companies screen data. Companies with
// incomplete data are listed as usual; oabi_missing says what they lack.
func (s *companyService) Get() ([]*responses.CompanyListItem, error) {
	companies, err := s.companyRepository.Get()
	if err != nil {
		return nil, err
	}
	users, err := s.companyRepository.UsersCountByCompany()
	if err != nil {
		return nil, err
	}
	certificates, err := s.companyRepository.IssuedCertificatesByCompany()
	if err != nil {
		return nil, err
	}

	items := make([]*responses.CompanyListItem, 0, len(companies))
	for _, c := range companies {
		issued := certificates[c.ID]
		items = append(items, &responses.CompanyListItem{
			Company:                *c,
			UsersCount:             users[c.ID],
			OabiCertificatesIssued: issued,
			ClientIDLocked:         issued > 0,
			OabiMissing:            utils.CompanyOabiMissing(c.ClientID, c.Rut, c.RazonSocial, c.Address),
		})
	}
	return items, nil
}

// Update validates the company, keeps client_id unique and refuses to change
// it once OABI certificates were issued with it.
func (s *companyService) Update(id string, companyRequest *request.Company) error {
	companyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.NewValidationError("invalid company id")
	}
	existing, err := s.companyRepository.GetById(id)
	if err != nil || existing == nil {
		return utils.NewValidationError("company not found")
	}
	if err := s.validate(companyRequest, companyID); err != nil {
		return err
	}

	if strings.TrimSpace(companyRequest.ClientID) != strings.TrimSpace(existing.ClientID) {
		issued, err := s.companyRepository.CountIssuedCertificates(companyID)
		if err != nil {
			return err
		}
		if issued > 0 {
			return utils.NewValidationError(utils.CompanyClientIDLockedMessage)
		}
	}

	company, err := mapping.CompanyRequestToCompany(companyRequest)
	if err != nil {
		return err
	}
	return s.companyRepository.Update(companyID, company)
}

// validate runs the field rules (normalizing the RUT) and the client_id
// uniqueness check.
func (s *companyService) validate(req *request.Company, exclude primitive.ObjectID) error {
	if err := utils.ValidateCompanyRequest(req); err != nil {
		return err
	}
	clientID := strings.TrimSpace(req.ClientID)
	other, err := s.companyRepository.GetByClientID(clientID, exclude)
	if err != nil {
		return err
	}
	if other != nil {
		return utils.NewValidationError(utils.CompanyClientIDInUseMessage(clientID, other.Name))
	}
	return nil
}

// ClientIDAvailable backs the live check of the company form.
func (s *companyService) ClientIDAvailable(clientID, excludeID string) (*responses.CompanyClientIDAvailability, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, utils.NewValidationError("client_id is required")
	}
	exclude := primitive.NilObjectID
	if strings.TrimSpace(excludeID) != "" {
		oid, err := primitive.ObjectIDFromHex(strings.TrimSpace(excludeID))
		if err != nil {
			return nil, utils.NewValidationError("invalid exclude_id")
		}
		exclude = oid
	}
	other, err := s.companyRepository.GetByClientID(clientID, exclude)
	if err != nil {
		return nil, err
	}
	if other != nil {
		return &responses.CompanyClientIDAvailability{Available: false, UsedBy: other.Name}, nil
	}
	return &responses.CompanyClientIDAvailability{Available: true, UsedBy: ""}, nil
}

// GetUsers lists the users of a company (its own users and the internal
// users assigned to it as a client).
func (s *companyService) GetUsers(id string) ([]responses.CompanyUser, error) {
	companyID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, utils.NewValidationError("invalid company id")
	}
	return s.companyRepository.GetCompanyUsers(companyID)
}

func (s *companyService) Delete(id string) (bool, error) {
	companyId, _ := primitive.ObjectIDFromHex(id)
	homologation, err := s.homologationRepository.GetByCompany(companyId)
	if err != nil {
		return false, err
	}
	if homologation != nil {
		return true, err
	}
	deviceTacking, err := s.deviceTrackingRepository.GetByCompany(companyId)
	if err != nil {
		return false, err
	}
	if deviceTacking != nil {
		return true, err
	}
	user, err := s.userRepository.GetUserByCompany(companyId)
	if err != nil {
		return false, err
	}
	if user != nil {
		return true, err
	}
	err = s.companyRepository.Delete(companyId)
	if err != nil {
		return false, err
	}
	return false, nil
}
