package mapping

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPlanRequestToTestPlan(testPlan *request.TestPlan) (*models.TestPlan, error) {

	var testCategories []primitive.ObjectID
	for _, value := range testPlan.TestCategories {
		oid, err := primitive.ObjectIDFromHex(value)
		if err != nil {
			return nil, err
		}
		testCategories = append(testCategories, oid)
	}
	return &models.TestPlan{
		Name:           testPlan.Name,
		TestCategories: testCategories,
		IsActive:       true,
		Description:    testPlan.Description,
		User:           testPlan.UserID,
	}, nil
}
