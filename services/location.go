package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// ILocationService is the location service
type ILocationService interface {
	Create(*request.Location) (string, error)
	Get() ([]*responses.Location, error)
}

type locationService struct {
	locationRepository repositories.ILocationRepository
}

// NewLocationService is a constructor
func NewLocationService(locationRepository repositories.ILocationRepository) ILocationService {
	return &locationService{
		locationRepository: locationRepository,
	}
}

// Create creates a new cateogry
func (s *locationService) Create(locationRequest *request.Location) (string, error) {

	location, err := mapping.LocationRequestToLocation(locationRequest)
	if err != nil {
		return "", err
	}

	id, err := s.locationRepository.Create(location)

	if err != nil {
		return "", err
	}

	return id.Hex(), nil
}

// Get gets a list of all categories
func (s *locationService) Get() ([]*responses.Location, error) {
	result, err := s.locationRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}
