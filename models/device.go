package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Profile model
type Device struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty"`
	Type                 int                `bson:"type,omitempty"`
	Brand                string             `bson:"brand,omitempty"`
	CommercialModel      bool               `bson:"commercial_model,omitempty"`
	TechnicalModel       bool               `bson:"technical_model,omitempty"`
	DisplayType          string             `bson:"display_type"`
	DisplaySize          string             `bson:"display_size"`
	DisplayResolution    string             `bson:"display_resolution"`
	PlatformOs           string             `bson:"platform_os"`
	PlatformVersion      string             `bson:"platform_version"`
	PlatformChipsetBand  string             `bson:"platform_chipset_band"`
	PlatformChipsetModel string             `bson:"platform_chipset_model"`
	PlatformCpu          string             `bson:"platform_cpu"`
	MemoryRom            string             `bson:"memory_rom"`
	MemoryRam            string             `bson:"memory_ram"`
	MemoryExtended       string             `bson:"memory_extended"`
	MemoryCpu            string             `bson:"memory_cpu"`
	MemoryType           string             `bson:"memory_type"`
	CameraFront          string             `bson:"camera_front"`
	CameraBack           string             `bson:"camera_back"`
	CommunicationWlan    string             `bson:"communication_wlan"`
	CommunicationGps     string             `bson:"communication_gps"`
	CommunicationNfc     string             `bson:"communication_nfc"`
	CommunicationRadio   string             `bson:"communication_radio"`
	CommunicationUsb     string             `bson:"communication_usb"`
	BatteryType          string             `bson:"battery_type"`
	BatteryCapacity      string             `bson:"battery_capacity"`
	BatteryState         string             `bson:"battery_state"`
	SensorFingerprint    string             `bson:"sensor_fingerprint"`
	SensorProximity      string             `bson:"sensor_proximity"`
	SensorAmbientLight   string             `bson:"sensor_ambient_light"`
	SensorAccelerometer  string             `bson:"sensor_accelerometer"`
	SensorGyroscope      string             `bson:"sensor_gyroscope"`
	SensorHall           string             `bson:"sensor_hall"`
}
