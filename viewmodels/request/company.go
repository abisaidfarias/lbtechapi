package request

// Company model
type Company struct {
	Email   string `json:"email"  binding:"required"`
	Name    string `json:"name"  binding:"required"`
	Address string `json:"address"  binding:"required"`
}
