package request

// Country model
type Country struct {
	Name string `json:"name"  binding:"required"`
}
