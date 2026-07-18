package request

// Company model
type Company struct {
	ID          string `bson:"_id" json:"_id"`
	Email       string `json:"email" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Address     string `json:"address" binding:"required"`
	LogoUrl     string `json:"logo_url" bson:"logo_url"`
	ClientID    string `json:"client_id"`
	Rut         string `json:"rut"`
	RazonSocial string `json:"razon_social"`
}
