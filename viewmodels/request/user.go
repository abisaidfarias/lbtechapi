package request

// TODO REMOVE BSON, should only be handle on model
// UserRequest response to client
type User struct {
	Email      string   `json:"email" bson:"email" binding:"required,email"`
	Password   string   `json:"password" bson:"password" binding:"required,passwordFormat"`
	Name       string   `json:"name" bson:"name" binding:"required"`
	LastName   string   `json:"lastName" bson:"lastName" binding:"required"`
	Phone      string   `json:"phone" bson:"phone"`
	Profile    string   `json:"profile" binding:"required"`
	Company    string   `json:"company"`
	IsInternal bool     `json:"is_internal"`
	Brands     []string `json:"brands"`
	Countries  []string `json:"countries"`
	Clients    []string `json:"clients"`
	UserID     string   `json:"user_id"`
}
