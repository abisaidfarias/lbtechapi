package responses

type VolumeTypeBrandChart struct {
	Types             []string            `json:"types" bson:"types"`
	Brands            []string            `json:"brand" bson:"brands"`
	StackedSerieChart []StackedSerieChart `json:"stacked_serie" bson:"stacked_serie"`
}
