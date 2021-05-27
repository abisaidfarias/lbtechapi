package responses

type ChartTypeCountry struct {
	Keys  TypeCountry `json:"_id" bson:"_id"`
	Count int         `json:"count" bson:"count"  `
}
