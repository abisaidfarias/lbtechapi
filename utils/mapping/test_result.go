package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func TestResultRequestToTestResult(testResult *request.TestResultResume) *models.TestResult {

	var hyperlinks []models.Hyperlink = []models.Hyperlink{}
	for _, h := range testResult.Hyperlinks {
		var hyperlink models.Hyperlink
		hyperlink.Link = h.Link
		hyperlink.Description = h.Description
		hyperlinks = append(hyperlinks, hyperlink)
	}
	return &models.TestResult{
		Name:             testResult.Name,
		Code:             testResult.Code,
		IsActive:         true,
		Result:           testResult.Result,
		OverviewIssue:    testResult.OverviewIssue,
		StepsToReproduce: testResult.StepsToReproduce,
		ActualResult:     testResult.ActualResult,
		ExpectedResult:   testResult.ExpectedResult,
		IssueFrequency:   testResult.IssueFrequency,
		IssueSeverity:    testResult.IssueSeverity,
		Images:           testResult.Images,
		Hyperlinks:       hyperlinks,
		Value:            testResult.Value,
	}
}
