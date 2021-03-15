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

// ITestCateogryRepository is the category repository for test cases
type ITestCategoryRepository interface {
	Create(category *models.TestCategory) (*primitive.ObjectID, error)
	Get() ([]*models.TestCategory, error)
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
func (r *testCategoryRepository) Get() ([]*models.TestCategory, error) {

	var categories []*models.TestCategory

	filter := bson.M{}

	cursor, err := testCategoryCollection.Find(context.TODO(), filter)

	if err != nil {
		return nil, fmt.Errorf("%w", utils.ErrorInQuery)
	}

	for cursor.Next(context.TODO()) {
		var result models.TestCategory
		err := cursor.Decode(&result)
		if err != nil {
			return nil, fmt.Errorf("%w", utils.ErrorInQuery)
		}
		categories = append(categories, &result)
	}

	cursor.Close(context.TODO())

	return categories, nil
}
