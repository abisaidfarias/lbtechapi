package models

// AuthCredentials credentials used in sign in
type AuthCredentials struct {
	Email    string `json:"email" bson:"email" binding:"required" validate:"email"`
	Password string `json:"password" bson:"password" binding:"required"`
}
