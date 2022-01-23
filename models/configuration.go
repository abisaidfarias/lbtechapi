package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Configuration struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	DeviceTypes        []string           `bson:"device_types"`
	Roms               []string           `bson:"roms"`
	Rams               []string           `bson:"rams"`
	MemoryCapabilities []string           `bson:"memory_capabilities"`
	UsbCommunications  []string           `bson:"usb_communications"`
	GsmBands           []string           `bson:"gsm_bands"`
	WcdmaBands         []string           `bson:"wcdma_bands"`
	LteBands           []string           `bson:"lte_bands"`
	CaCombos           []string           `bson:"ca_combos"`
}
