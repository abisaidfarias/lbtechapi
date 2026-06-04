package responses

type SearchOptionRelationByBrand struct {
	CommercialModels []string `bson:"commercial_models" json:"commercial_models"`
	Countries        []string `bson:"countries" json:"countries"`
	Locations        []string `bson:"locations" json:"locations"`
}

type SearchOptionRelationByModel struct {
	Brands    []string `bson:"brands" json:"brands"`
	Countries []string `bson:"countries" json:"countries"`
	Locations []string `bson:"locations" json:"locations"`
}

type SearchOptionRelationByCountry struct {
	Brands           []string `bson:"brands" json:"brands"`
	CommercialModels []string `bson:"commercial_models" json:"commercial_models"`
	Locations        []string `bson:"locations" json:"locations"`
}

type SearchOptionRelationByLocation struct {
	Brands           []string `bson:"brands" json:"brands"`
	CommercialModels []string `bson:"commercial_models" json:"commercial_models"`
	Countries        []string `bson:"countries" json:"countries"`
}

type SearchOptionRelations struct {
	ByBrand    map[string]SearchOptionRelationByBrand    `bson:"by_brand" json:"byBrand"`
	ByModel    map[string]SearchOptionRelationByModel    `bson:"by_model" json:"byModel"`
	ByCountry  map[string]SearchOptionRelationByCountry  `bson:"by_country" json:"byCountry"`
	ByLocation map[string]SearchOptionRelationByLocation `bson:"by_location" json:"byLocation"`
}

// Company model
type SearchOption struct {
	Brands           []string              `bson:"brands" json:"brands"`
	CommercialModels []string              `bson:"commercial_models" json:"commercial_models"`
	Countries        []string              `bson:"countries" json:"countries"`
	Locations        []string              `bson:"locations" json:"locations"`
	Relations        SearchOptionRelations `bson:"relations" json:"relations"`
}
