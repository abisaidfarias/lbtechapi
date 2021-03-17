package repositories

import (
	"context"
	"fmt"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/models"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ITestCateogryRepository is the category repository for test cases
type ITestCategoryRepository interface {
	Create(category *models.TestCategory) (*primitive.ObjectID, error)
	Get() ([]*bson.M, error)
}

type testCategoryRepository struct {
}

// NewTestCategoryRepository is a constructor for the category repository
func NewTestCategoryRepository() ITestCategoryRepository {
	return &testCategoryRepository{}
}

var testCategoryCollection = database.GetInstance().Collection("test-categories")

// Create a new category
func (r *testCategoryRepository) Create(category *models.TestCategory) (*primitive.ObjectID, error) {

	res, err := testCategoryCollection.InsertOne(context.TODO(), category)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	id := res.InsertedID.(primitive.ObjectID)

	return &id, nil
}

// Get returns a list of all test categories
func (r *testCategoryRepository) Get() ([]*bson.M, error) {

	var categories []*bson.M

	cursor, err := testCategoryCollection.Find(context.TODO(), bson.M{})

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	for cursor.Next(context.TODO()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return nil, fmt.Errorf("%w", utils.ErrorInQuery)
		}
		categories = append(categories, &result)
	}

	cursor.Close(context.TODO())

	return categories, nil
}
