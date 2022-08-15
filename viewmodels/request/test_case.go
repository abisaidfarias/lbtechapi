package request

type TestCase struct {
	Code        string `json:"code,omitempty" binding:"required,testCaseCode"`
	Name        string `json:"name" binding:"required"`
	CategoryID  string `json:"category_id" binding:"required"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description" binding:"required"`
	Device      string `json:"device"`
	Expected    string `json:"expected" binding:"required"`
}
