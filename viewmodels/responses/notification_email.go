package responses

// Company model
type NotificationEmail struct {
	Type     int    `bson:"type" json:"type"`
	Email    string `bson:"email" json:"email"`
	Name     string `bson:"name" json:"name"`
	LastName string `bson:"lastName" json:"lastName"`
}
