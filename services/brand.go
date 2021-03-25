package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IBrandService is the brand service
type IBrandService interface {
	Create(*request.Brand) error
	Get() ([]*responses.Brand, error)
}

type brandService struct {
	brandRepository repositories.IBrandRepository
}

// NewBrandService is a constructor
func NewBrandService(brandRepository repositories.IBrandRepository) IBrandService {
	return &brandService{
		brandRepository: brandRepository,
	}
}

// Create creates a new cateogry
func (s *brandService) Create(brandRequest *request.Brand) error {

	brand, err := mapping.BrandRequestToBrand(brandRequest)
	if err != nil {
		return err
	}
	err = s.brandRepository.Create(brand)

	if err != nil {
		return err
	}

	return nil
}

// Get gets a list of all categories
func (s *brandService) Get() ([]*responses.Brand, error) {
	result, err := s.brandRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}
