package request

// UserRequest response to client
type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"passwordHash"`
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	Phone    string `json:"phone"`
	Dni      string `json:"dni"`
}
