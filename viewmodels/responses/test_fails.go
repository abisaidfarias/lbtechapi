package responses

type TestFails struct {
	TestResults []TestResult `bson:"test_results,omitempty" json:"test_results"`
	TotalTest   int          `bson:"total_test" json:"total_test"`
	TotalHigh   int          `bson:"total_high" json:"total_high"`
	TotalMedium int          `bson:"total_medium" json:"total_medium"`
	TotalLow    int          `bson:"total_low" json:"total_low"`
}
