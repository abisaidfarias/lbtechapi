package responses

type ChartVolumeCountry struct {
	Keys  VolumeCountryBrand `json:"_id" bson:"_id"`
	Count int                `json:"count" bson:"count"  `
}
