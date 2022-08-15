package responses

// Company model
type CategoryResult struct {
	Pass           int              `json:"pass" bson:"pass"`
	Fail           int              `json:"fail" bson:"fail"`
	NoRun          int              `json:"no_run" bson:"no_run"`
	NA             int              `json:"na" bson:"na"`
	TestCaseResume []TestCaseResume `json:"test_cases_resume" bson:"test_cases_resume"`
}
