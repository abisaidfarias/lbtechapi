package repositories

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type ITestPlanRepository interface {
	Create(*models.TestPlan) error
	Get() ([]*responses.TestPlan, error)
	GetById(string) (*responses.TestPlan, error)
	Update(string, *models.TestPlan) error
	Delete(string) error
}

type testPlanRepository struct {
}

func NewTestPlanRepository() ITestPlanRepository {
	return &testPlanRepository{}
}

var testPlanCollection = database.GetInstance().Collection("test_plans")

// Create a new tet case
func (r *testPlanRepository) Create(testPlan *models.TestPlan) error {

	_, err := testPlanCollection.InsertOne(context.TODO(), testPlan)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *testPlanRepository) Get() ([]*responses.TestPlan, error) {

	cursor, err := testPlanCollection.Aggregate(context.TODO(), queries.GetTestPlans())
	if err != nil {
		panic(err)
	}
	var testPlans []*responses.TestPlan
	for cursor.Next(context.TODO()) {
		var result responses.TestPlan
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		result.UserName = fmt.Sprintf("%s %s", result.Users[0].Name, result.Users[0].LastName)
		result.Users = []responses.User{}
		result.TotalCategory = len(result.TestCategories)
		for _, element := range result.TestCategories {
			result.TotalTest += len(element.TestCases)
		}
		testPlans = append(testPlans, &result)
	}
	cursor.Close(context.TODO())
	return testPlans, nil
}

func (r *testPlanRepository) GetById(id string) (*responses.TestPlan, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result responses.TestPlan

	err = testCaseCollection.FindOne(context.TODO(), queries.GeTestPlanById(oid)).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	return &result, nil
}
func (r *testPlanRepository) Update(id string, testPlan *models.TestPlan) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter, update := queries.UpdateTestPlan(testPlan, oid)

	_, err = testPlanCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *testPlanRepository) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)

	_, err = testPlanCollection.DeleteOne(context.TODO(), queries.DeleteTestPlan(oid))

	if err != nil {
		return err
	}

	return nil
}
