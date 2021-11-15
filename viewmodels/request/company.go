package request

// Company model
type Company struct {
	Email   string `json:"email"  binding:"required"`
	Name    string `json:"name"  binding:"required"`
	Address string `json:"address"  binding:"required"`
	LogoUrl string `json:"logo_url" bson:"logo_url"`
}
