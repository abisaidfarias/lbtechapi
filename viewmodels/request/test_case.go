package request

// UserRequest response to client
type TestCase struct {
	Code        string `json:"code"  binding:"required,testCaseCode"`
	Name        string `json:"name"  binding:"required"`
	CategoryID  string `json:"categoryId"  binding:"required"`
	Description string `json:"description"`
	Device      string `json:"device"`
	Expected    string `json:"expected"`
	IsActive    bool   `json:"isActive"`
}
