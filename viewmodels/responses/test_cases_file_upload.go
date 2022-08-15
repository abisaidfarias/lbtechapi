package responses

type TestCaseFileUpload struct {
	InvalidRows []int `json:"invalid_rows"`
	TotalRows   int   `json:"total_rows"`
}
