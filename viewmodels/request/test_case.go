package request

// TestCase model
type TestCase struct {
	Code        string `json:"code,omitempty" binding:"required,testCaseCode"`
	Name        string `json:"name" binding:"required"`
	CategoryID  string `json:"categoryId" binding:"required"`
	IsActive    bool   `json:"isActive"`
	Description string `json:"description"`
	Device      string `json:"device"`
	Expected    string `json:"expected"`
}
