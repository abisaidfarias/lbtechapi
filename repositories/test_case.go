package repositories

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
)

// ITestCaseRepository is the test cases repository
type ITestCaseRepository interface {
	Create(*models.TestCase) error
	Get() ([]*bson.M, error)
	GetByID(string) (*bson.M, error)
	Update(string, *models.TestCase) error
	Delete(string) error
}

type testCaseRepository struct {
}

// NewTestCaseRepository is a constructor for the case repository
func NewTestCaseRepository() ITestCaseRepository {
	return &testCaseRepository{}
}

var testCaseCollection = database.GetInstance().Collection("test_cases")

// Create a new tet case
func (r *testCaseRepository) Create(test *models.TestCase) error {

	_, err := testCaseCollection.InsertOne(context.TODO(), test)
	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *testCaseRepository) Get() ([]*bson.M, error) {

	cursor, err := testCaseCollection.Aggregate(context.TODO(), queries.GetAll())

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

func (r *testCaseRepository) GetByID(id string) (*bson.M, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result bson.M

	err = testCaseCollection.FindOne(context.TODO(), queries.GetById(oid)).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	return &result, nil
}
func (r *testCaseRepository) Update(id string, testCase *models.TestCase) error {

	log.Println("testCase", testCase)
	oid, err := primitive.ObjectIDFromHex(id)

	filter, update := queries.Update(testCase, oid)

	_, err = testCaseCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *testCaseRepository) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)

	_, err = testCaseCollection.DeleteOne(context.TODO(), queries.Delete(oid))

	if err != nil {
		return err
	}

	return nil
}
