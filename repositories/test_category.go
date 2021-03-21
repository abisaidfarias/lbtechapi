package repositories

import (
	"context"
	"fmt"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ITestCateogryRepository is the category repository for test cases
type ITestCategoryRepository interface {
	Create(category *models.TestCategory) (*primitive.ObjectID, error)
	Get() ([]*responses.TestCategory, error)
}

type testCategoryRepository struct {
}

// NewTestCategoryRepository is a constructor for the category repository
func NewTestCategoryRepository() ITestCategoryRepository {
	return &testCategoryRepository{}
}

var testCategoryCollection = database.GetInstance().Collection("test_categories")

// Create a new category
func (r *testCategoryRepository) Create(category *models.TestCategory) (*primitive.ObjectID, error) {

	res, err := testCategoryCollection.InsertOne(context.TODO(), category)

	if err != nil {
		return nil, err
	}

	id := res.InsertedID.(primitive.ObjectID)

	return &id, nil
}

// Get returns a list of all test categories
func (r *testCategoryRepository) Get() ([]*responses.TestCategory, error) {

	cursor, err := testCategoryCollection.Find(context.TODO(), bson.M{})

	if err != nil {
		return nil, err
	}
	var testCategories []*responses.TestCategory
	for cursor.Next(context.TODO()) {
		var result responses.TestCategory
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		result.Name = fmt.Sprintf("%s(%d)", result.Name, len(result.TestCases))
		result.TestCases = nil
		testCategories = append(testCategories, &result)
	}

	cursor.Close(context.TODO())

	return testCategories, nil
}
