package request

// Claim model
type Claim struct {
	Name  string `bson:"name" binding:"required"`
	Allow bool   `bson:"allow" binding:"required"`
}
