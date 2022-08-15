package request

// Company model
type Person struct {
	Name string `json:"name"  binding:"required"`
}
