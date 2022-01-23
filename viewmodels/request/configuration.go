package request

// Profile model
type Configuration struct {
	DeviceTypes        []string `bson:"device_types" json:"device_types"`
	Roms               []string `bson:"roms" json:"roms"`
	Rams               []string `bson:"rams" json:"rams"`
	MemoryCapabilities []string `bson:"memory_capabilities" json:"memory_capabilities"`
	UsbCommunications  []string `bson:"usb_communications" json:"usb_communications"`
	GsmBands           []string `bson:"gsm_bands" json:"gsm_bands"`
	WcdmaBands         []string `bson:"wcdma_bands" json:"wcdma_bands"`
	LteBands           []string `bson:"lte_bands" json:"lte_bands"`
	CaCombos           []string `bson:"ca_combos" json:"ca_combos"`
}
