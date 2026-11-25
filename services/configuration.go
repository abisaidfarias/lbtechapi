package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
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
	deviceCatalogs          IDeviceCatalogService
}

// NewConfigurationService is a constructor. The five device lists
// (DeviceTypes, Roms, Rams, MemoryCapabilities, UsbCommunications) now come
// from the device catalogs.
func NewConfigurationService(
	configurationRepository repositories.IConfigurationRepository,
	deviceCatalogs IDeviceCatalogService,
) IConfigurationService {
	return &configurationService{
		configurationRepository: configurationRepository,
		deviceCatalogs:          deviceCatalogs,
	}
}

// Create replaces the configuration. The five catalog-backed lists in the
// body are ignored: the stored ones are kept, since the catalog is the source
// of truth for them.
func (s *configurationService) Create(configurationRequest *request.Configuration) (string, error) {
	configuration, err := mapping.ConfigurationRequestToConfiguration(configurationRequest)
	if err != nil {
		return "", err
	}
	current, err := s.configurationRepository.Get()
	if err != nil {
		return "", err
	}
	configuration.DeviceTypes, configuration.Roms, configuration.Rams = nil, nil, nil
	configuration.MemoryCapabilities, configuration.UsbCommunications = nil, nil
	if len(current) > 0 && current[0] != nil {
		stored := current[0]
		configuration.DeviceTypes = stored.DeviceTypes
		configuration.Roms = stored.Roms
		configuration.Rams = stored.Rams
		configuration.MemoryCapabilities = stored.MemoryCapabilities
		configuration.UsbCommunications = stored.UsbCommunications
	}
	id, err := s.configurationRepository.Create(configuration)
	if err != nil {
		return "", err
	}
	return id.Hex(), nil
}

// Get returns the configuration with the five device lists computed from
// the catalogs (active values, in order). A category without catalog values
// (catalog not migrated yet) keeps the stored list.
func (s *configurationService) Get() ([]*responses.Configuration, error) {
	result, err := s.configurationRepository.Get()
	if err != nil {
		return nil, err
	}
	if s.deviceCatalogs == nil {
		return result, nil
	}
	active, err := s.deviceCatalogs.ActiveValues()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 && len(active) > 0 {
		result = []*responses.Configuration{{}}
	}
	for _, cfg := range result {
		ApplyDeviceCatalogsToConfiguration(cfg, active)
	}
	return result, nil
}

// ApplyDeviceCatalogsToConfiguration overrides the catalog-backed lists.
func ApplyDeviceCatalogsToConfiguration(cfg *responses.Configuration, active map[string][]string) {
	if cfg == nil {
		return
	}
	targets := map[string]*[]string{
		enums.DeviceCatalogDeviceType:       &cfg.DeviceTypes,
		enums.DeviceCatalogRom:              &cfg.Roms,
		enums.DeviceCatalogRam:              &cfg.Rams,
		enums.DeviceCatalogMemoryCapability: &cfg.MemoryCapabilities,
		enums.DeviceCatalogUsb:              &cfg.UsbCommunications,
	}
	for category, target := range targets {
		if values, ok := active[category]; ok {
			*target = values
		}
	}
}
