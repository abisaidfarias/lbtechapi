package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCategoryRequestToTestCategory(testCategory *request.TestCategory) (*models.TestCategory, error) {

	return &models.TestCategory{
		Name:        testCategory.Name,
		Description: testCategory.Description,
		TestCases:   []primitive.ObjectID{},
	}, nil
}
func TestCategoryResponseToTestCategory(testCategory *responses.TestCategory) models.TestCategory {

	return models.TestCategory{
		Name:        testCategory.Name,
		Description: testCategory.Description,
		TestCases:   []primitive.ObjectID{},
	}
}
