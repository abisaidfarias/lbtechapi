package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRequestToTestCase(testCase *request.TestCase) (*models.TestCase, error) {

	objID, err := primitive.ObjectIDFromHex(testCase.CategoryID)

	if err != nil {
		return nil, err
	}

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
