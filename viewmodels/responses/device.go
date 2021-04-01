package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

// Profile model
type Device struct {
	ID                     primitive.ObjectID `bson:"_id" json:"_id"`
	Type                   string             `json:"type,omitempty"`
	Brand                  string             `json:"brand,omitempty"`
	CommercialModel        string             `json:"commercial_model"`
	TechnicalModel         string             `json:"technical_model"`
	DisplayType            string             `json:"display_type"`
	DisplaySize            string             `json:"display_size"`
	DisplayResolution      string             `json:"display_resolution"`
	PlatformOs             string             `json:"platform_os"`
	PlatformVersion        string             `json:"platform_version"`
	PlatformChipsetBrand   string             `json:"platform_chipset_brand"`
	PlatformChipsetModel   string             `json:"platform_chipset_model"`
	PlatformCpu            string             `json:"platform_cpu"`
	MemoryRom              string             `json:"memory_rom"`
	MemoryRam              string             `json:"memory_ram"`
	MemoryExtended         string             `json:"memory_extended"`
	MemoryCpu              string             `json:"memory_cpu"`
	MemoryType             string             `json:"memory_type"`
	CameraFront            string             `json:"camera_front"`
	CameraBack             string             `json:"camera_back"`
	CommunicationWlan      string             `json:"communication_wlan"`
	CommunicationGps       string             `json:"communication_gps"`
	CommunicationNfc       string             `json:"communication_nfc"`
	CommunicationRadio     string             `json:"communication_radio"`
	CommunicationUsb       string             `json:"communication_usb"`
	CommunicationBluetooth string             `json:"communication_blutooth"`
	BatteryType            string             `json:"battery_type"`
	BatteryCapacity        string             `json:"battery_capacity"`
	BatteryState           string             `json:"battery_state"`
	SensorFingerprint      string             `json:"sensor_fingerprint"`
	SensorProximity        string             `json:"sensor_proximity"`
	SensorAmbientLight     string             `json:"sensor_ambient_light"`
	SensorAccelerometer    string             `json:"sensor_accelerometer"`
	SensorGyroscope        string             `json:"sensor_gyroscope"`
	SensorHall             string             `json:"sensor_hall"`
	BandGsm                string             `json:"band_gsm"`
	BandWcdma              string             `json:"band_wcdma"`
	BandLte                string             `json:"band_lte"`
	Band5g                 string             `json:"band_5g"`
	NetworkGsm             bool               `json:"network_gsm"`
	NetworkWcdma           bool               `json:"network_wcdma"`
	NetworkLte             bool               `json:"network_lte"`
	NetworkVolte           bool               `json:"network_volte"`
	NetworkVowifi          bool               `json:"network_vowifi"`
	NetworkVilte           bool               `json:"network_vilte"`
	Network5g              bool               `json:"network_5g"`
}
