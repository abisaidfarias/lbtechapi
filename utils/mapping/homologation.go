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
	testCategories []*responses.TestCategoryExpanded, companyId primitive.ObjectID,
	deviceId primitive.ObjectID, countryId primitive.ObjectID) *models.Homologation {

	var testResults []models.TestResult
	for _, testCategory := range testCategories {

		for _, testCase := range testCategory.TestCases {
			if testCase.IsActive {
				var testResult models.TestResult
				testResult.Code = testCase.Code
				testResult.Description = testCase.Description
				testResult.Expected = testCase.Expected
				testResult.IsActive = testCase.IsActive
				testResult.Name = testCase.Name
				testResult.TestCategory = testCase.TestCategory
				testResults = append(testResults, testResult)
			}
		}
	}

	return &models.Homologation{
		Company:          companyId,
		Device:           deviceId,
		Country:          countryId,
		SoftwareVersion:  homologation.SoftwareVersion,
		HardwareVersion:  homologation.HardwareVersion,
		Type:             homologation.Type,
		CurrentPhase:     homologation.CurrentPhase,
		PlanningDate:     homologation.PlanningDate,
		SampleStartDate:  homologation.SampleStartDate,
		SampleEndDate:    homologation.SampleEndDate,
		TestStartDate:    homologation.TestStartDate,
		TestEndDate:      homologation.TestEndDate,
		UnderStartDate:   homologation.UnderStartDate,
		UnderEndDate:     homologation.UnderEndDate,
		CompletedDate:    homologation.CompletedDate,
		IsCustomTestPlan: homologation.IsCustomTestPlan,
		Status:           enums.HomologationStatus_value["IN_PROGRESS"],
		CreatedDate:      time.Now(),
		TestResults:      testResults,
	}
}
