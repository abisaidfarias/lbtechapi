package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ICountryService is the country service
type ICountryService interface {
	Create(*request.Country) error
	Get() ([]*responses.Country, error)
	Delete(string) (bool, error)
	Update(string, *request.Country) error
}

type countryService struct {
	countryRepository      repositories.ICountryRepository
	homologationRepository repositories.IHomologationRepository
}

// NewCountryService is a constructor
func NewCountryService(countryRepository repositories.ICountryRepository,
	homologationRepository repositories.IHomologationRepository) ICountryService {
	return &countryService{
		countryRepository:      countryRepository,
		homologationRepository: homologationRepository,
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
func (s *countryService) Update(id string, countryRequest *request.Country) error {
	countryId, _ := primitive.ObjectIDFromHex(id)
	country, err := mapping.CountryRequestToCountry(countryRequest)

	if err != nil {
		return err
	}
	err = s.countryRepository.Update(countryId, country)
	if err != nil {
		return err
	}
	return nil
}
func (s *countryService) Delete(id string) (bool, error) {
	countryId, _ := primitive.ObjectIDFromHex(id)

	homologation, err := s.homologationRepository.GetByCountry(countryId)
	if err != nil {
		return false, err
	}
	if homologation != nil {
		return true, err
	}
	err = s.countryRepository.Delete(countryId)
	if err != nil {
		return false, err
	}
	return false, nil
}
