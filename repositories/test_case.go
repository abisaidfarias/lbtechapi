package repositories

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	mgoBson "go.mongodb.org/mongo-driver/bson"
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

var testCaseCollection = database.GetInstance().Collection("test-cases")

// Create a new tet case
func (r *testCaseRepository) Create(test *models.TestCase) error {

	_, err := testCaseCollection.InsertOne(context.TODO(), test)

	if err != nil {
		return fmt.Errorf("%w", utils.ErrorInQuery)
	}

	return nil
}

// Get returns a list of all test cases
func (r *testCaseRepository) Get() ([]*bson.M, error) {

	// mongo tutorial
	lookupStage := mgoBson.D{
		{"$lookup", mgoBson.D{
			{"from", "test-categories"},
			{"localField", "test_category"},
			{"foreignField", "_id"},
			{"as", "test_category"},
		}}}
	unwindStage := mgoBson.D{{"$unwind", mgoBson.D{{"path", "$test_category"}, {"preserveNullAndEmptyArrays", false}}}}
	matchStage := mgoBson.D{{"$match", mgoBson.D{{"is_active", true}}}}
	cursor, err := testCaseCollection.Aggregate(context.TODO(), mongo.Pipeline{lookupStage, unwindStage, matchStage})

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

	err = testCaseCollection.FindOne(context.TODO(), mgoBson.M{"_id": oid}).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	return &result, nil
}
func (r *testCaseRepository) Update(id string, testCase *models.TestCase) error {
	
	oid, err := primitive.ObjectIDFromHex(id)

	update := mgoBson.M{
		"$set": testCase,
	}
	log.Println("update test", testCase)
	_, err = testCaseCollection.UpdateOne(context.TODO(), mgoBson.M{"_id": oid}, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *testCaseRepository) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)

	_, err = testCaseCollection.DeleteOne(context.TODO(), mgoBson.M{"_id": oid})

	if err != nil {
		return err
	}

	return nil
}
