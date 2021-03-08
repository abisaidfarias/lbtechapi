package request

// UserRequest response to client
type UserRequest struct {
	Email    string `json:"email" bson:"email" binding:"required"`
	Password string `json:"password" bson:"password" binding:"required"`
	Name     string `json:"name" bson:"name" binding:"required"`
	LastName string `json:"lastName" bson:"lastName" binding:"required"`
	Phone    string `json:"phone" bson:"phone"`
	Dni      string `json:"dni" bson:"dni" binding:"required"`
}
