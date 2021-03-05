package models

// User model
type User struct {
	ID           string `bson:"_id,omitempty"`
	Email        string `bson:"email,omitempty"`
	PasswordHash string `bson:"passwordHash,omitempty"`
}
