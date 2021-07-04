package request

// TODO REMOVE BSON, should only be handle on model
// UserRequest response to client
type UserResume struct {
	Email      string `json:"email" bson:"email" binding:"required,email"`
	Name       string `json:"name" bson:"name" binding:"required"`
	LastName   string `json:"lastName" bson:"lastName" binding:"required"`
	IsInternal bool   `json:"is_internal"`
	UserID     string `json:"user_id"`
}
