package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// ICountryService is the country service
type ICountryService interface {
	Create(*request.Country) error
	Get() ([]*responses.Country, error)
}

type countryService struct {
	countryRepository repositories.ICountryRepository
}

// NewCountryService is a constructor
func NewCountryService(countryRepository repositories.ICountryRepository) ICountryService {
	return &countryService{
		countryRepository: countryRepository,
	}
}

// Create creates a new cateogry
func (s *countryService) Create(countryRequest *request.Country) error {

	country, err := mapping.CountryRequestToCountry(countryRequest)
	if err != nil {
		return err
	}
	err = s.countryRepository.Create(country)

	if err != nil {
		return err
	}

	return nil
}

// Get gets a list of all categories
func (s *countryService) Get() ([]*responses.Country, error) {
	result, err := s.countryRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}
