package responses

// Country model
type DashboardReport struct {
	Countries  []string `bson:"countries"`
	Name       string   `bson:"name"`
	BandGsm    []string `bson:"band_gsm" json:"band_gsm"`
	BandWcdma  []string `bson:"band_wcdma" json:"band_wcdma"`
	BandLte    []string `bson:"band_lte" json:"band_lte"`
	Band5gNsa  []string `bson:"band_5g_nsa" json:"band_5g_nsa"`
	Band5gSa   []string `bson:"band_5g_sa" json:"band_5g_sa"`
	CarrierAgg []string `bson:"carrier_agg" json:"carrier_agg"`
}
