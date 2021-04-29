package models

// Profile model
type Hyperlink struct {
	Link        string `bson:"link" json:"link"`
	Description string `bson:"description" json:"description"`
}
