package responses

type VolumeCountryBrand struct {
	Brand   string `json:"brand" bson:"brand"`
	Country string `json:"country" bson:"country"`
}
