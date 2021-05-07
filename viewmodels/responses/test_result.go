package responses

type TestResult struct {
	Code             string       `bson:"code" json:"code"`
	Name             string       `bson:"name" json:"name"`
	TestCategory     TestCategory `bson:"test_category" json:"test_category"`
	IsActive         bool         `bson:"is_active" json:"is_active"`
	Description      string       `bson:"description" json:"description"`
	Expected         string       `bson:"expected" json:"expected"`
	Result           int          `bson:"result" json:"result"`
	IssueDescription string       `bson:"issue_description" json:"issue_description"`
	Images           []string     `bson:"images" json:"images"`
	Hyperlinks       []Hyperlink  `bson:"hyperlinks" json:"hyperlinks"`
	Value            string       `bson:"value" json:"value"`
	IssueFrequency   int          `bson:"issue_frequency" json:"issue_frequency"`
	IssueSeverity    int          `bson:"issue_severity" json:"issue_severity"`
}
