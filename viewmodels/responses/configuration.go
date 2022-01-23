package responses

// Company model
type Configuration struct {
	DeviceTypes        []string `bson:"device_types"`
	Roms               []string `bson:"roms"`
	Rams               []string `bson:"rams"`
	MemoryCapabilities []string `bson:"memory_capabilities"`
	UsbCommunications  []string `bson:"usb_communications"`
	GsmBands           []string `bson:"gsm_bands"`
	WcdmaBands         []string `bson:"wcdma_bands"`
	LteBands           []string `bson:"lte_bands"`
	CaCombos           []string `bson:"ca_combos"`
}
