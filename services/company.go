package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ICompanyService is the company service
type ICompanyService interface {
	Create(*request.Company) error
	Get() ([]*responses.Company, error)
	Delete(string) (bool, error)
	Update(string, *request.Company) error
}

type companyService struct {
	companyRepository      repositories.ICompanyRepository
	homologationRepository repositories.IHomologationRepository
	deviceTrackingRepository repositories.IDeviceTrackingRepository
	userRepository repositories.IUserRepository
}

// NewCompanyService is a constructor
func NewCompanyService(companyRepository repositories.ICompanyRepository,
	homologationRepository repositories.IHomologationRepository,
	deviceTrackingRepository repositories.IDeviceTrackingRepository,
	userRepository repositories.IUserRepository) ICompanyService {
	return &companyService{
		companyRepository:      companyRepository,
		homologationRepository: homologationRepository,
		deviceTrackingRepository: deviceTrackingRepository,
		userRepository: userRepository,
	}
}

// Create creates a new cateogry
func (s *companyService) Create(companyRequest *request.Company) error {

	company, err := mapping.CompanyRequestToCompany(companyRequest)
	if err != nil {
		return err
	}
	err = s.companyRepository.Create(company)

	if err != nil {
		return err
	}

	return nil
}

// Get gets a list of all categories
func (s *companyService) Get() ([]*responses.Company, error) {
	result, err := s.companyRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}
func (s *companyService) Update(id string, companyRequest *request.Company) error {
	companyId, _ := primitive.ObjectIDFromHex(id)
	company, err := mapping.CompanyRequestToCompany(companyRequest)

	if err != nil {
		return err
	}
	err = s.companyRepository.Update(companyId, company)
	if err != nil {
		return err
	}
	return nil
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
