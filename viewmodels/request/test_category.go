package request

// TestCategory request to create category
type TestCategory struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
