package responses

type TestCaseResume struct {
	Code   string `json:"code,omitempty" binding:"required,testCaseCode"`
	Name   string `json:"name" binding:"required"`
	Result int    `bson:"result" json:"result"`
}
