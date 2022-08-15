package responses

// Company model

type CategoryExpanded struct {
	Pass             int                `json:"pass" bson:"pass"`
	Fail             int                `json:"fail" bson:"fail"`
	NoRun            int                `json:"no_run" bson:"no_run"`
	NA               int                `json:"na" bson:"na"`
	TestResultResume []TestResultResume `json:"test_result_resume" bson:"test_result_resume"`
}
