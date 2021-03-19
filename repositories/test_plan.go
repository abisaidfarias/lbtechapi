package repositories

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
)

type ITestPlanRepository interface {
	Create(*models.TestPlan) error
	Get() ([]*bson.M, error)
	GetById(string) (*bson.M, error)
	Update(string, *models.TestPlan) error
	Delete(string) error
}

type testPlanRepository struct {
}

func NewTestPlanRepository() ITestPlanRepository {
	return &testPlanRepository{}
}

var testPlanCollection = database.GetInstance().Collection("test_plan")

// Create a new tet case
func (r *testPlanRepository) Create(testPlan *models.TestPlan) error {

	_, err := testPlanCollection.InsertOne(context.TODO(), testPlan)
	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *testPlanRepository) Get() ([]*bson.M, error) {

	cursor, err := testPlanCollection.Aggregate(context.TODO(), queries.GetTestPlans())

	if err != nil {

		panic(err)
	}
	var testCases []*bson.M
	if err = cursor.All(context.TODO(), &testCases); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return testCases, nil
}

func (r *testPlanRepository) GetById(id string) (*bson.M, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result bson.M

	err = testCaseCollection.FindOne(context.TODO(), queries.GeTestPlanById(oid)).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	return &result, nil
}
func (r *testPlanRepository) Update(id string, testPlan *models.TestPlan) error {

	oid, err := primitive.ObjectIDFromHex(id)

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
