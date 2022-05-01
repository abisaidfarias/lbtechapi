package responses

// Company model
type NotificationEmail struct {
	Type  int    `bson:"type" json:"type"`
	Email string `bson:"email" json:"email"`
}
