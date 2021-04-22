package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Country model
type Country struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	BandGsm   []string           `bson:"band_gsm"`
	BandWcdma []string           `bson:"band_wcdma"`
	BandLte   []string           `bson:"band_lte"`
	Band5g    []string           `bson:"band_5g"`
}
