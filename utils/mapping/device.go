package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func DeviceRequestToDevice(device *request.Device) *models.Device {

	return &models.Device{
		Type:                 device.Type,
		Brand:                device.Brand,
		CommercialModel:      device.CommercialModel,
		TechnicalModel:       device.TechnicalModel,
		DisplayType:          device.DisplayType,
		DisplaySize:          device.DisplaySize,
		DisplayResolution:    device.DisplayResolution,
		PlatformOs:           device.PlatformOs,
		PlatformVersion:      device.PlatformVersion,
		PlatformChipsetBand:  device.PlatformChipsetBand,
		PlatformChipsetModel: device.PlatformChipsetModel,
		PlatformCpu:          device.PlatformCpu,
		MemoryRom:            device.MemoryRom,
		MemoryRam:            device.MemoryRam,
		MemoryExtended:       device.MemoryExtended,
		MemoryCpu:            device.MemoryCpu,
		MemoryType:           device.MemoryType,
		CameraFront:          device.CameraFront,
		CameraBack:           device.CameraBack,
		CommunicationWlan:    device.CommunicationWlan,
		CommunicationGps:     device.CommunicationGps,
		CommunicationNfc:     device.CommunicationNfc,
		CommunicationRadio:   device.CommunicationRadio,
		CommunicationUsb:     device.CommunicationUsb,
		BatteryType:          device.BatteryType,
		BatteryCapacity:      device.BatteryCapacity,
		BatteryState:         device.BatteryState,
		SensorFingerprint:    device.SensorFingerprint,
		SensorProximity:      device.SensorProximity,
		SensorAmbientLight:   device.SensorAmbientLight,
		SensorAccelerometer:  device.SensorAccelerometer,
		SensorGyroscope:      device.SensorGyroscope,
		SensorHall:           device.SensorHall,
	}
}
