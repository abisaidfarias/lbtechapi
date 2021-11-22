package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

// Profile model
type DeviceExpanded struct {
	ID                     primitive.ObjectID `bson:"_id" json:"_id"`
	Type                   string             `bson:"type" json:"type"`
	Brand                  Brand              `bson:"brand" json:"brand"`
	CommercialModel        string             `bson:"commercial_model" json:"commercial_model"`
	TechnicalModel         string             `bson:"technical_model" json:"technical_model"`
	DisplayType            string             `bson:"display_type" json:"display_type"`
	DisplaySize            string             `bson:"display_size" json:"display_size"`
	DisplayResolution      string             `bson:"display_resolution" json:"display_resolution"`
	PlatformOs             string             `bson:"platform_os" json:"platform_os"`
	PlatformVersion        string             `bson:"platform_version" json:"platform_version"`
	PlatformChipsetBrand   string             `bson:"platform_chipset_brand" json:"platform_chipset_brand"`
	PlatformChipsetModel   string             `bson:"platform_chipset_model" json:"platform_chipset_model"`
	PlatformCpu            string             `bson:"platform_cpu" json:"platform_cpu"`
	MemoryRom              string             `bson:"memory_rom" json:"memory_rom"`
	MemoryRam              string             `bson:"memory_ram" json:"memory_ram"`
	MemoryExtended         bool               `bson:"memory_extended" json:"memory_extended"`
	MemoryCpu              string             `bson:"memory_cpu" json:"memory_cpu"`
	MemoryType             string             `bson:"memory_type" json:"memory_type"`
	CameraFront            string             `bson:"camera_front" json:"camera_front"`
	CameraBack             string             `bson:"camera_back" json:"camera_back"`
	CommunicationWlan      bool               `bson:"communication_wlan" json:"communication_wlan"`
	CommunicationGps       bool               `bson:"communication_gps" json:"communication_gps"`
	CommunicationNfc       bool               `bson:"communication_nfc" json:"communication_nfc"`
	CommunicationRadio     bool               `bson:"communication_radio" json:"communication_radio"`
	CommunicationUsb       string             `bson:"communication_usb" json:"communication_usb"`
	CommunicationBluetooth bool               `bson:"communication_blutooth" json:"communication_blutooth"`
	BatteryType            string             `bson:"battery_type" json:"battery_type"`
	BatteryCapacity        string             `bson:"battery_capacity" json:"battery_capacity"`
	BatteryState           string             `bson:"battery_state" json:"battery_state"`
	BatteryInductedCharger bool               `bson:"battery_inducted_charger" json:"battery_inducted_charger"`
	SensorFingerprint      bool               `bson:"sensor_fingerprint" json:"sensor_fingerprint"`
	SensorProximity        bool               `bson:"sensor_proximity" json:"sensor_proximity"`
	SensorAmbientLight     bool               `bson:"sensor_ambient_light" json:"sensor_ambient_light"`
	SensorAccelerometer    bool               `bson:"sensor_accelerometer" json:"sensor_accelerometer"`
	SensorGyroscope        bool               `bson:"sensor_gyroscope" json:"sensor_gyroscope"`
	SensorHall             bool               `bson:"sensor_hall" json:"sensor_hall"`
	BandGsm                []string           `bson:"band_gsm" json:"band_gsm"`
	BandWcdma              []string           `bson:"band_wcdma" json:"band_wcdma"`
	BandLte                []string           `bson:"band_lte" json:"band_lte"`
	Band5g                 []string           `bson:"band_5g" json:"band_5g"`
	CarrierAgg             []string           `bson:"carrier_agg" json:"carrier_agg"`
	NetworkGsm             bool               `bson:"network_gsm" json:"network_gsm"`
	NetworkWcdma           bool               `bson:"network_wcdma" json:"network_wcdma"`
	NetworkLte             bool               `bson:"network_lte" json:"network_lte"`
	NetworkVolte           bool               `bson:"network_volte" json:"network_volte"`
	NetworkVowifi          bool               `bson:"network_vowifi" json:"network_vowifi"`
	NetworkVilte           bool               `bson:"network_vilte" json:"network_vilte"`
	Network5g              bool               `bson:"network_5g" json:"network_5g"`
	NetworkCarrierAgg      bool               `bson:"network_carrier_agg" json:"network_carrier_agg"`
	ImageUrl               string             `bson:"image_url" json:"image_url"`
	SoftwareCode           string             `bson:"software_code" json:"software_code"`
	HardwareCode           string             `bson:"hardware_code" json:"hardware_code"`
	IngCode                string             `bson:"ing_code" json:"ing_code"`
	LoggingCode            string             `bson:"logging_code" json:"logging_code"`
	SimSupported           string             `bson:"sim_supported" json:"sim_supported"`
	DualSim                string             `bson:"dual_sim" json:"dual_sim"`
	SimType                string             `bson:"sim_type" json:"sim_type"`
	Esim                   bool               `bson:"e_sim" json:"e_sim"`
}
