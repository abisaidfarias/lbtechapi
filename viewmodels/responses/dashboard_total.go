package responses

type DashboardTotal struct {
	CurrentPhase int `json:"_id" bson:"_id"`
	Count  int `json:"count" bson:"count"  `
}
