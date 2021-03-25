package request

// Company model
type Brand struct {
	Name string `json:"name"  binding:"required"`
}
