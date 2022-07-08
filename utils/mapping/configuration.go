package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func ConfigurationRequestToConfiguration(configuration *request.Configuration) (*models.Configuration, error) {

	return &models.Configuration{
		DeviceTypes:        configuration.DeviceTypes,
		Roms:               configuration.Roms,
		Rams:               configuration.Rams,
		MemoryCapabilities: configuration.MemoryCapabilities,
		UsbCommunications:  configuration.UsbCommunications,
		GsmBands:           configuration.GsmBands,
		WcdmaBands:         configuration.WcdmaBands,
		LteBands:           configuration.LteBands,
		CaCombos:           configuration.CaCombos,
		Bands5g:            configuration.Bands5g,
	}, nil
}
