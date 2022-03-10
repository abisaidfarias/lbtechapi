package responses

// Company model
type SearchOption struct {
	Brands           []string `bson:"brands" json:"brands"`
	CommercialModels []string `bson:"commercial_models" json:"commercial_models"`
	Countries        []string `bson:"countries" json:"countries"`
	Locations        []string `bson:"locations" json:"locations"`
}
