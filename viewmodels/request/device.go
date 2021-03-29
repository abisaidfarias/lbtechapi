package request

// Profile model
type Device struct {
	Type                 int    `json:"type,omitempty" binding:"required"`
	Brand                string `json:"brand,omitempty" binding:"required"`
	CommercialModel      bool   `json:"commercial_model,omitempty" binding:"required"`
	TechnicalModel       bool   `json:"technical_model,omitempty" binding:"required"`
	DisplayType          string `json:"display_type"`
	DisplaySize          string `json:"display_size"`
	DisplayResolution    string `json:"display_resolution"`
	PlatformOs           string `json:"platform_os"`
	PlatformVersion      string `json:"platform_version"`
	PlatformChipsetBand  string `json:"platform_chipset_band"`
	PlatformChipsetModel string `json:"platform_chipset_model"`
	PlatformCpu          string `json:"platform_cpu"`
	MemoryRom            string `json:"memory_rom"`
	MemoryRam            string `json:"memory_ram"`
	MemoryExtended       string `json:"memory_extended"`
	MemoryCpu            string `json:"memory_cpu"`
	MemoryType           string `json:"memory_type"`
	CameraFront          string `json:"camera_front"`
	CameraBack           string `json:"camera_back"`
	CommunicationWlan    string `json:"communication_wlan"`
	CommunicationGps     string `json:"communication_gps"`
	CommunicationNfc     string `json:"communication_nfc"`
	CommunicationRadio   string `json:"communication_radio"`
	CommunicationUsb     string `json:"communication_usb"`
	BatteryType          string `json:"battery_type"`
	BatteryCapacity      string `json:"battery_capacity"`
	BatteryState         string `json:"battery_state"`
	SensorFingerprint    string `json:"sensor_fingerprint"`
	SensorProximity      string `json:"sensor_proximity"`
	SensorAmbientLight   string `json:"sensor_ambient_light"`
	SensorAccelerometer  string `json:"sensor_accelerometer"`
	SensorGyroscope      string `json:"sensor_gyroscope"`
	SensorHall           string `json:"sensor_hall"`
}
