package request

type TestResultResume struct {
	Code             string      `bson:"code" json:"code" binding:"required"`
	Result           int         `bson:"result" json:"result"`
	IssueDescription string      `bson:"issue_description" json:"issue_description"`
	Images           []string    `bson:"images" json:"images"`
	Hyperlinks       []Hyperlink `bson:"hyperlinks" json:"hyperlinks"`
	Value            string      `bson:"value" json:"value"`
	IssueFrequency   int         `bson:"issue_frequency" json:"issue_frequency"`
	IssueSeverity    int         `bson:"issue_severity" json:"issue_severity"`
}
