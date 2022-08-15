package responses

// Company model
type HomologationReport struct {
	Categories map[string]CategoryResult `json:"categories" bson:"categories"`
}
