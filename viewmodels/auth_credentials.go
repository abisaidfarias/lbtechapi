package viewmodels

// AuthCredentials credentials used in sign in
type AuthCredentials struct {
	Email    string `json:"email" bson:"email" binding:"required,email"`
	Password string `json:"password" bson:"password" binding:"required"`
}
