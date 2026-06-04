package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Profile model
type Device struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty"`
	Type                   string             `bson:"type,omitempty"`
	Brand                  primitive.ObjectID `bson:"brand,omitempty"`
	CommercialModel        string             `bson:"commercial_model,omitempty"`
	TechnicalModel         string             `bson:"technical_model,omitempty"`
	DisplayType            string             `bson:"display_type"`
	DisplaySize            string             `bson:"display_size"`
	DisplayResolution      string             `bson:"display_resolution"`
	PlatformOs             string             `bson:"platform_os"`
	PlatformVersion        string             `bson:"platform_version"`
	PlatformChipsetBrand   string             `bson:"platform_chipset_brand"`
	PlatformChipsetModel   string             `bson:"platform_chipset_model"`
	PlatformCpu            string             `bson:"platform_cpu"`
	MemoryRom              string             `bson:"memory_rom"`
	MemoryRam              string             `bson:"memory_ram"`
	MemoryExtended         bool               `bson:"memory_extended"`
	MemoryCpu              string             `bson:"memory_cpu"`
	MemoryType             string             `bson:"memory_type"`
	CameraFront            string             `bson:"camera_front"`
	CameraBack             string             `bson:"camera_back"`
	CommunicationWlan      bool               `bson:"communication_wlan"`
	CommunicationGps       bool               `bson:"communication_gps"`
	CommunicationNfc       bool               `bson:"communication_nfc"`
	CommunicationRadio     bool               `bson:"communication_radio"`
	CommunicationUsb       string             `bson:"communication_usb"`
	CommunicationBluetooth bool               `bson:"communication_blutooth"`
	BatteryType            string             `bson:"battery_type"`
	BatteryCapacity        string             `bson:"battery_capacity"`
	BatteryState           string             `bson:"battery_state"`
	BatteryInductedCharger bool               `bson:"battery_inducted_charger"`
	SensorFingerprint      bool               `bson:"sensor_fingerprint"`
	SensorProximity        bool               `bson:"sensor_proximity"`
	SensorAmbientLight     bool               `bson:"sensor_ambient_light"`
	SensorAccelerometer    bool               `bson:"sensor_accelerometer"`
	SensorGyroscope        bool               `bson:"sensor_gyroscope"`
	SensorHall             bool               `bson:"sensor_hall"`
	BandGsm                []string           `bson:"band_gsm"`
	BandWcdma              []string           `bson:"band_wcdma"`
	BandLte                []string           `bson:"band_lte"`
	Band5g                 []string           `bson:"band_5g"`
	CarrierAgg             []string           `bson:"carrier_agg"`
	NetworkGsm             bool               `bson:"network_gsm"`
	NetworkWcdma           bool               `bson:"network_wcdma"`
	NetworkLte             bool               `bson:"network_lte"`
	NetworkVolte           bool               `bson:"network_volte"`
	NetworkVowifi          bool               `bson:"network_vowifi"`
	NetworkVilte           bool               `bson:"network_vilte"`
	Network5g              bool               `bson:"network_5g"`
	NetworkCarrierAgg      bool               `bson:"network_carrier_agg"`
	ImageUrl               string             `bson:"image_url"`
	SoftwareCode           string             `bson:"software_code"`
	HardwareCode           string             `bson:"hardware_code"`
	IngCode                string             `bson:"ing_code"`
	LoggingCode            string             `bson:"logging_code"`
	SimSupported           string             `bson:"sim_supported"`
	DualSim                string             `bson:"dual_sim"`
	SimType                string             `bson:"sim_type"`
	Esim                   bool               `bson:"e_sim"`
	SarValue               float64            `bson:"sar_value,omitempty"`
}
