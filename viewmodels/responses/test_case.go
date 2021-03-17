package responses

type TestCase struct {
	Code         string `json:"code,omitempty" binding:"required,testCaseCode"`
	Name         string `json:"name" binding:"required"`
	IsActive     bool   `json:"is_active"`
	Description  string `json:"description" binding:"required"`
	Device       string `json:"device" binding:"required"`
	Expected     string `json:"expected" binding:"required"`
	TestCategory struct {
		Name        string `bson:"name"`
		Description string `bson:"description"`
	} `json:"test_category"`
}
