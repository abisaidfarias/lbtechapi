package responses

type TimeChart struct {
	CountryTestingChart map[string][3]float64 `json:"contry_testing_chart" bson:"contry_testing_chart"`
	CountryUnderChart   map[string][3]float64 `json:"contry_under_chart" bson:"contry_under_chart"`
	BrandTestingChart   map[string][3]float64 `json:"brand_testing_chart" bson:"brand_testing_chart"`
	BrandUnderChart     map[string][3]float64 `json:"brand_under_chart" bson:"brand_under_chart"`
}
