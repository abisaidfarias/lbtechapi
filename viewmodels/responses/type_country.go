package responses

type TypeCountry struct {
	Type    int    `json:"type" bson:"type"`
	Country string `json:"country" bson:"country"`
	Month   int    `json:"month" bson:"month"`
	Year    int    `json:"year" bson:"year"`
}
