package request

// Profile model
type Device struct {
	Type                   string   `json:"type,omitempty" binding:"required"`
	Brand                  string   `json:"brand,omitempty" binding:"required"`
	CommercialModel        string   `json:"commercial_model,omitempty" binding:"required"`
	TechnicalModel         string   `json:"technical_model,omitempty" binding:"required"`
	DisplayType            string   `json:"display_type"`
	DisplaySize            string   `json:"display_size"`
	DisplayResolution      string   `json:"display_resolution"`
	PlatformOs             string   `json:"platform_os"`
	PlatformVersion        string   `json:"platform_version"`
	PlatformChipsetBrand   string   `json:"platform_chipset_brand"`
	PlatformChipsetModel   string   `json:"platform_chipset_model"`
	PlatformCpu            string   `json:"platform_cpu"`
	MemoryRom              string   `json:"memory_rom"`
	MemoryRam              string   `json:"memory_ram"`
	MemoryExtended         bool     `json:"memory_extended"`
	MemoryCpu              string   `json:"memory_cpu"`
	MemoryType             string   `json:"memory_type"`
	CameraFront            string   `json:"camera_front"`
	CameraBack             string   `json:"camera_back"`
	CommunicationWlan      bool     `json:"communication_wlan"`
	CommunicationGps       bool     `json:"communication_gps"`
	CommunicationNfc       bool     `json:"communication_nfc"`
	CommunicationRadio     bool     `json:"communication_radio"`
	CommunicationUsb       string   `json:"communication_usb"`
	CommunicationBluetooth bool     `json:"communication_blutooth"`
	BatteryType            string   `json:"battery_type"`
	BatteryCapacity        string   `json:"battery_capacity"`
	BatteryState           string   `json:"battery_state"`
	BatteryInductedCharger bool     `json:"battery_inducted_charger"`
	SensorFingerprint      bool     `json:"sensor_fingerprint"`
	SensorProximity        bool     `json:"sensor_proximity"`
	SensorAmbientLight     bool     `json:"sensor_ambient_light"`
	SensorAccelerometer    bool     `json:"sensor_accelerometer"`
	SensorGyroscope        bool     `json:"sensor_gyroscope"`
	SensorHall             bool     `json:"sensor_hall"`
	BandGsm                []string `json:"band_gsm"`
	BandWcdma              []string `json:"band_wcdma"`
	BandLte                []string `json:"band_lte"`
	Band5g                 []string `json:"band_5g"`
	CarrierAgg             []string `json:"carrier_agg"`
	NetworkGsm             bool     `json:"network_gsm"`
	NetworkWcdma           bool     `json:"network_wcdma"`
	NetworkLte             bool     `json:"network_lte"`
	NetworkVolte           bool     `json:"network_volte"`
	NetworkVowifi          bool     `json:"network_vowifi"`
	NetworkVilte           bool     `json:"network_vilte"`
	Network5g              bool     `json:"network_5g"`
	NetworkCarrierAgg      bool     `json:"network_carrier_agg"`
	ImageUrl               string   `json:"image_url"`
	SoftwareCode           string   `json:"software_code"`
	HardwareCode           string   `json:"hardware_code"`
	IngCode                string   `json:"ing_code"`
	LoggingCode            string   `json:"logging_code"`
	SimSupported           string   `json:"sim_supported"`
	DualSim                string   `json:"dual_sim"`
	SimType                string   `json:"sim_type"`
	Esim                   bool     `json:"e_sim"`
}
