package request

// Claim model
type Claim struct {
	Name  string `json:"name" bson:"name" binding:"required"`
	Allow bool   `json:"allow" bson:"allow"`
}
