package request

// Company model
type Location struct {
	Name string `json:"name"  binding:"required"`
}
