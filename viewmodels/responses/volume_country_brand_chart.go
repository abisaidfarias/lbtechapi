package responses

type VolumeCountryBrandChart struct {
	Countries         []string            `json:"countries" bson:"countries"`
	Brands            []string            `json:"brand" bson:"brands"`
	StackedSerieChart []StackedSerieChart `json:"stacked_serie" bson:"stacked_serie"`
}
