package repositories

import (
	"context"
	"fmt"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

// ITestCaseRepository is the test cases repository
type ITestCaseRepository interface {
	Create(*models.TestCase) error
	Get() ([]*models.TestCase, error)
	GetByID(string) (*models.TestCase, error)
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
func (r *testCaseRepository) Get() ([]*models.TestCase, error) {

	var cases []*models.TestCase

	cursor, err := testCaseCollection.Find(context.TODO(), bson.M{})

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	for cursor.Next(context.TODO()) {
		var result models.TestCase
		err := cursor.Decode(&result)
		if err != nil {
			return nil, fmt.Errorf("%w", utils.ErrorInQuery)
		}
		cases = append(cases, &result)
	}

	cursor.Close(context.TODO())

	return cases, nil
}

func (r *testCaseRepository) GetByID(id string) (*models.TestCase, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, err
	}

	var result models.TestCase

	err = testCaseCollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	return &result, nil
}

// Update updates the test case
func (r *testCaseRepository) Update(id string, testCase *models.TestCase) error {
	oid, err := primitive.ObjectIDFromHex(id)

	update := bson.M{
		"$set": testCase,
	}

	_, err = testCaseCollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, update)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes the test case
func (r *testCaseRepository) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)

	_, err = testCaseCollection.DeleteOne(context.TODO(), bson.M{"_id": oid})

	if err != nil {
		return err
	}

	return nil
}
