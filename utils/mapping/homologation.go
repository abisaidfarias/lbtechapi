package mapping

import (
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HomologationRequestToHomologation(homologation *request.Homologation,
	testCategories []*responses.TestCategoryExpanded, companyID primitive.ObjectID,
	deviceID primitive.ObjectID, countryID primitive.ObjectID,
	testPlanID primitive.ObjectID, brandID primitive.ObjectID) *models.Homologation {

	var testResults []models.TestResult
	for _, testCategory := range testCategories {

		for _, testCase := range testCategory.TestCases {
			if testCase.IsActive {
				var category models.TestCategory
				category.ID = testCategory.ID
				category.Description = testCategory.Description
				category.Name = testCategory.Name
				var testResult models.TestResult
				testResult.Code = testCase.Code
				testResult.Description = testCase.Description
				testResult.Expected = testCase.Expected
				testResult.IsActive = testCase.IsActive
				testResult.Name = testCase.Name
				testResult.TestCategory = category
				testResults = append(testResults, testResult)
			}
		}
	}

	return &models.Homologation{
		Company:            companyID,
		Device:             deviceID,
		Country:            countryID,
		SoftwareVersion:    homologation.SoftwareVersion,
		HardwareVersion:    homologation.HardwareVersion,
		Type:               homologation.Type,
		CurrentPhase:       homologation.CurrentPhase,
		PlanningDate:       homologation.PlanningDate,
		SampleStartDate:    homologation.SampleStartDate,
		SampleEndDate:      homologation.SampleEndDate,
		TestStartDate:      homologation.TestStartDate,
		TestEndDate:        homologation.TestEndDate,
		UnderStartDate:     homologation.UnderStartDate,
		UnderEndDate:       homologation.UnderEndDate,
		CompletedDate:      homologation.CompletedDate,
		IsCustomTestPlan:   homologation.IsCustomTestPlan,
		Status:             enums.HomologationStatus_value["IN_PROGRESS"],
		CreatedDate:        time.Now(),
		TestResults:        testResults,
		TestPlan:           testPlanID,
		Brand:              brandID,
		IsInternalProject:  homologation.IsInternalProject,
		OsVersion:          homologation.OsVersion,
		DocumentUrl:        homologation.DocumentUrl,
		ResultUrl:          homologation.ResultUrl,
		Carrier:            homologation.Carrier,
		TestingType:        homologation.TestingType,
		Comment:            homologation.Comment,
		ApprovalTypeOption: homologation.ApprovalTypeOption,
		DashBoardPhase:     homologation.CurrentPhase,
	}
}

func HomologationRequestToHomologationResume(homologation *request.HomologationResume) *models.Homologation {

	return &models.Homologation{
		SoftwareVersion:    homologation.SoftwareVersion,
		HardwareVersion:    homologation.HardwareVersion,
		CurrentPhase:       homologation.CurrentPhase,
		PlanningDate:       homologation.PlanningDate,
		SampleStartDate:    homologation.SampleStartDate,
		SampleEndDate:      homologation.SampleEndDate,
		TestStartDate:      homologation.TestStartDate,
		TestEndDate:        homologation.TestEndDate,
		UnderStartDate:     homologation.UnderStartDate,
		UnderEndDate:       homologation.UnderEndDate,
		CompletedDate:      homologation.CompletedDate,
		Status:             homologation.Status,
		IsInternalProject:  homologation.IsInternalProject,
		OsVersion:          homologation.OsVersion,
		DocumentUrl:        homologation.DocumentUrl,
		ResultUrl:          homologation.ResultUrl,
		Carrier:            homologation.Carrier,
		TestingType:        homologation.TestingType,
		Comment:            homologation.Comment,
		ApprovalTypeOption: homologation.ApprovalTypeOption,
	}
}
func HomologationRequestToHomologationUpdate(homologation *request.Homologation) *models.Homologation {

	companyID, _ := primitive.ObjectIDFromHex(homologation.Company)
	deviceID, _ := primitive.ObjectIDFromHex(homologation.Device)
	brandID, _ := primitive.ObjectIDFromHex(homologation.Brand)

	return &models.Homologation{
		Company:           companyID,
		Device:            deviceID,
		SoftwareVersion:   homologation.SoftwareVersion,
		HardwareVersion:   homologation.HardwareVersion,
		PlanningDate:      homologation.PlanningDate,
		SampleStartDate:   homologation.SampleStartDate,
		SampleEndDate:     homologation.SampleEndDate,
		TestStartDate:     homologation.TestStartDate,
		TestEndDate:       homologation.TestStartDate,
		UnderStartDate:    homologation.UnderStartDate,
		UnderEndDate:      homologation.UnderEndDate,
		CompletedDate:     homologation.CompletedDate,
		Brand:             brandID,
		IsInternalProject: homologation.IsInternalProject,
		OsVersion:         homologation.OsVersion,
		Carrier:           homologation.Carrier,
		TestingType:       homologation.TestingType,
		CurrentPhase:      homologation.CurrentPhase,
	}
}
func TestResultsRequestToTestResults(testResultsRequest []request.TestResult) []models.TestResult {

	var testResults []models.TestResult
	for _, testResulRequest := range testResultsRequest {

		var testResult models.TestResult
		testResult.Code = testResulRequest.Code
		testResult.Description = testResulRequest.Description
		testResult.Expected = testResulRequest.Expected
		testResult.IsActive = testResulRequest.IsActive
		testResult.Name = testResulRequest.Name
		testResult.ActualResult = testResulRequest.ActualResult
		testResult.ExpectedResult = testResulRequest.ExpectedResult
		testResult.Images = testResulRequest.Images
		testResult.IssueFrequency = testResulRequest.IssueFrequency
		testResult.IssueSeverity = testResulRequest.IssueSeverity
		testResult.OverviewIssue = testResulRequest.OverviewIssue
		testResult.Result = testResulRequest.Result
		testResult.StepsToReproduce = testResulRequest.StepsToReproduce
		testResult.Value = testResulRequest.Value

		var hyperlinks []models.Hyperlink = []models.Hyperlink{}
		for _, hyperlinkRequest := range testResulRequest.Hyperlinks {
			var hyperlink models.Hyperlink
			hyperlink.Description = hyperlinkRequest.Description
			hyperlink.Link = hyperlinkRequest.Link
			hyperlinks = append(hyperlinks, hyperlink)
		}
		testResult.Hyperlinks = hyperlinks
		var testCategory models.TestCategory
		categoryID, _ := primitive.ObjectIDFromHex(testResulRequest.TestCategory.ID)
		testCategory.ID = categoryID
		testCategory.Description = testResulRequest.TestCategory.Description
		testCategory.Name = testResulRequest.TestCategory.Name
		testResult.TestCategory = testCategory
		testResults = append(testResults, testResult)
	}
	return testResults
}
func HomologationResponseToHomologationRequest(homologation responses.HomologationExpanded) request.Homologation {

	return request.Homologation{
		Company:            homologation.Company.ID.Hex(),
		Device:             homologation.Device.ID.Hex(),
		Country:            homologation.Country.ID.Hex(),
		SoftwareVersion:    homologation.SoftwareVersion,
		HardwareVersion:    homologation.HardwareVersion,
		Type:               homologation.Type,
		CurrentPhase:       homologation.CurrentPhase,
		PlanningDate:       *homologation.PlanningDate,
		SampleStartDate:    *homologation.SampleStartDate,
		SampleEndDate:      *homologation.SampleEndDate,
		TestStartDate:      *homologation.TestStartDate,
		TestEndDate:        *homologation.TestEndDate,
		UnderStartDate:     *homologation.UnderStartDate,
		UnderEndDate:       *homologation.UnderEndDate,
		CompletedDate:      *homologation.CompletedDate,
		IsCustomTestPlan:   homologation.IsCustomTestPlan,
		Status:             enums.HomologationStatus_value["IN_PROGRESS"],
		Brand:              homologation.Device.Brand,
		IsInternalProject:  homologation.IsInternalProject,
		OsVersion:          homologation.OsVersion,
		DocumentUrl:        homologation.DocumentUrl,
		ResultUrl:          homologation.ResultUrl,
		Carrier:            homologation.Carrier,
		TestingType:        homologation.TestingType,
		Comment:            homologation.Comment,
		ApprovalTypeOption: homologation.ApprovalTypeOption,
	}
}
