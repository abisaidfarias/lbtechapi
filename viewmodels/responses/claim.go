package responses

// Claim model
type Claim struct {
	Name  string `bson:"name"`
	Allow bool   `bson:"allow"`
}
