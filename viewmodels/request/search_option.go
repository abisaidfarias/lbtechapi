package request

// Company model
type SearchOption struct {
	Brand           string `bson:"brand" json:"brand"`
	CommercialModel string `bson:"commercial_model" json:"commercial_model"`
	Country         string `bson:"country" json:"country"`
	Location        string `bson:"location" json:"location"`
}
