package repositories

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type ITestPlanRepository interface {
	Create(*models.TestPlan, primitive.ObjectID) error
	Get() ([]*responses.TestPlanExpanded, error)
	GetById(string) (*responses.TestPlanExpanded, error)
	Update(string, *models.TestPlan) error
	Delete(primitive.ObjectID) error
}

type testPlanRepository struct {
}

func NewTestPlanRepository() ITestPlanRepository {
	return &testPlanRepository{}
}

var testPlanCollection = database.GetInstance().Collection("test_plans")

// Create a new tet case
func (r *testPlanRepository) Create(testPlan *models.TestPlan, userId primitive.ObjectID) error {

	var user responses.User

	err := userCollection.FindOne(context.TODO(), queries.GetUserById(userId)).Decode(&user)
	if err != nil {
		return err
	}
	testPlan.UserName = fmt.Sprintf("%s %s", user.Name, user.LastName)

	_, err = testPlanCollection.InsertOne(context.TODO(), testPlan)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *testPlanRepository) Get() ([]*responses.TestPlanExpanded, error) {

	cursor, err := testPlanCollection.Aggregate(context.TODO(), queries.GetTestPlans())
	if err != nil {
		return nil, err
	}
	var testPlans []*responses.TestPlanExpanded = []*responses.TestPlanExpanded{}
	for cursor.Next(context.TODO()) {
		var result responses.TestPlanExpanded
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		result.TotalCategory = len(result.TestCategories)
		for _, element := range result.TestCategories {
			result.TotalTest += len(element.TestCases)
		}
		testPlans = append(testPlans, &result)
	}
	cursor.Close(context.TODO())
	return testPlans, nil
}

func (r *testPlanRepository) GetById(id string) (*responses.TestPlanExpanded, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result responses.TestPlanExpanded

	err = testCaseCollection.FindOne(context.TODO(), queries.GeTestPlanById(oid)).Decode(&result)

	if err != nil {
		return nil, err
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
func (r *testPlanRepository) Delete(id primitive.ObjectID) error {

	_, err := testPlanCollection.DeleteOne(context.TODO(), queries.DeleteTestPlan(id))

	if err != nil {
		return err
	}

	return nil
}
