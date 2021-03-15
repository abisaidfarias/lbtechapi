package request

// TestCase model
type TestCase struct {
	Code        string `json:"code,omitempty"  binding:"testCaseCode" `
	Name        string `json:"name"  `
	CategoryID  string `json:"categoryId"  `
	IsActive    bool   `json:"isActive"`
	Description string `json:"description"`
	Device      string `json:"device"`
	Expected    string `json:"expected"`
}
