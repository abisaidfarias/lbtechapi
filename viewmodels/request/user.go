package request

// TODO REMOVE BSON, should only be handle on model
// UserRequest response to client
type UserRequest struct {
	Email     string   `json:"email" bson:"email" binding:"required,email"`
	Password  string   `json:"password" bson:"password" binding:"required,passwordFormat"`
	Name      string   `json:"name" bson:"name" binding:"required"`
	LastName  string   `json:"lastName" bson:"lastName" binding:"required"`
	Phone     string   `json:"phone" bson:"phone"`
	Dni       string   `json:"dni" bson:"dni" binding:"required"`
	Profile   string   `json:"profile" binding:"required"`
	Company   string   `json:"company" binding:"required"`
	Brands    []string `json:"brands"`
	Countries []string `json:"countries"`
}
