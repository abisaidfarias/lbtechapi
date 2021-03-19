package request

type TestPlan struct {
	Name           string   `json:"name" binding:"required"`
	TestCategories []string `json:"test_categories" binding:"required"`
	Description    string   `json:"description"`
}
