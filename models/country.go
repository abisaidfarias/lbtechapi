package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Country model
type Country struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	BandGsm    []string           `bson:"band_gsm"`
	BandWcdma  []string           `bson:"band_wcdma"`
	BandLte    []string           `bson:"band_lte"`
	Band5gNsa  []string           `bson:"band_5g_nsa"`
	Band5gSa   []string           `bson:"band_5g_sa"`
	CarrierAgg []string           `bson:"carrier_agg"`
	// Active is nil on countries created before the field existed; they
	// count as active. omitempty keeps tracking-log snapshots unchanged.
	Active *bool `bson:"active,omitempty"`
}

// IsCountryActive applies the "missing means active" rule.
func IsCountryActive(active *bool) bool {
	return active == nil || *active
}
