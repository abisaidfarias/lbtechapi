package responses

type TestResultResume struct {
	Code             string      `bson:"code" json:"code"`
	Name             string      `bson:"name" json:"name"`
	Description      string      `bson:"description" json:"description"`
	Expected         string      `bson:"expected" json:"expected"`
	Result           int         `bson:"result" json:"result"`
	OverviewIssue    string      `bson:"overview_issue" json:"overview_issue"`
	StepsToReproduce string      `bson:"steps_to_reproduce" json:"steps_to_reproduce"`
	ActualResult     string      `bson:"actual_result" json:"actual_result"`
	ExpectedResult   string      `bson:"expected_result" json:"expected_result"`
	Images           []string    `bson:"images" json:"images"`
	Hyperlinks       []Hyperlink `bson:"hyperlinks" json:"hyperlinks"`
	Value            string      `bson:"value" json:"value"`
	IssueFrequency   int         `bson:"issue_frequency" json:"issue_frequency"`
	IssueSeverity    int         `bson:"issue_severity" json:"issue_severity"`
}
