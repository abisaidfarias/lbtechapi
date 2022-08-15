package request

// TestCategory request to create category
type TestCategory struct {
	ID          string `json:"_id" bson:"_id"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
