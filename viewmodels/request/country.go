package request

// Country model
type Country struct {
	Name       string   `json:"name"  binding:"required"`
	BandGsm    []string `bson:"band_gsm" json:"band_gsm"`
	BandWcdma  []string `bson:"band_wcdma" json:"band_wcdma"`
	BandLte    []string `bson:"band_lte" json:"band_lte"`
	Band5g     []string `bson:"band_5g" json:"band_5g"`
	CarrierAgg []string `bson:"carrier_agg" json:"carrier_agg"`
}
