package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// IConfigurationService is the configuration service
type IConfigurationService interface {
	Create(*request.Configuration) (string, error)
	Get() ([]*responses.Configuration, error)
}

type configurationService struct {
	configurationRepository repositories.IConfigurationRepository
}

// NewConfigurationService is a constructor
func NewConfigurationService(configurationRepository repositories.IConfigurationRepository) IConfigurationService {
	return &configurationService{
		configurationRepository: configurationRepository,
	}
}

// Create creates a new cateogry
func (s *configurationService) Create(configurationRequest *request.Configuration) (string, error) {

	configuration, err := mapping.ConfigurationRequestToConfiguration(configurationRequest)
	if err != nil {
		return "", err
	}

	id, err := s.configurationRepository.Create(configuration)

	if err != nil {
		return "", err
	}

	return id.Hex(), nil
}

// Get gets a list of all categories
func (s *configurationService) Get() ([]*responses.Configuration, error) {
	result, err := s.configurationRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}
