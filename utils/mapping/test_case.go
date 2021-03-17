package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRequestToTestCase(testCase *request.TestCase, isCreate bool) (*models.TestCase, error) {

	objID, err := primitive.ObjectIDFromHex(testCase.CategoryID)

	if err != nil {
		return nil, err
	}
	if isCreate {
		return &models.TestCase{
			Code:         testCase.Code,
			Name:         testCase.Name,
			TestCategory: objID,
			IsActive:     testCase.IsActive,
			Description:  testCase.Description,
			Device:       testCase.Device,
			Expected:     testCase.Expected,
		}, nil
	}
	return &models.TestCase{
		Name:         testCase.Name,
		TestCategory: objID,
		IsActive:     testCase.IsActive,
		Description:  testCase.Description,
		Device:       testCase.Device,
		Expected:     testCase.Expected,
	}, nil
}
