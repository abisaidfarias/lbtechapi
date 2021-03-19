package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCategoryRequestToTestCategory(testCategory *request.TestCategory) (*models.TestCategory, error) {

	return &models.TestCategory{
		Name:        testCategory.Name,
		Description: testCategory.Description,
		TestCases:   []primitive.ObjectID{},
	}, nil
}
