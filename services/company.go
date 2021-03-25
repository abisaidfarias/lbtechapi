package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// ICompanyService is the company service
type ICompanyService interface {
	Create(*request.Company) error
	Get() ([]*responses.Company, error)
}

type companyService struct {
	companyRepository repositories.ICompanyRepository
}

// NewCompanyService is a constructor
func NewCompanyService(companyRepository repositories.ICompanyRepository) ICompanyService {
	return &companyService{
		companyRepository: companyRepository,
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
