package repositories

import (
	"context"
	"fmt"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ITestCateogryRepository is the category repository for test cases
type ITestCategoryRepository interface {
	Create(category *models.TestCategory) (*primitive.ObjectID, error)
	Get() ([]*responses.TestCategory, error)
	GetSimple() ([]*responses.TestCategory, error)
	GetByIds([]primitive.ObjectID) ([]*responses.TestCategoryExpanded, error)
	GetOtherCategory() (*models.TestCategory, error)
	CreateOtherCategory() (*models.TestCategory, error)
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

	cursor, err := testCategoryCollection.Aggregate(context.TODO(), queries.GetCategories())

	if err != nil {
		return nil, err
	}
	var testCategories []*responses.TestCategory = []*responses.TestCategory{}
	for cursor.Next(context.TODO()) {
		var categoryExpanded *responses.TestCategoryExpanded
		var category responses.TestCategory
		err := cursor.Decode(&categoryExpanded)
		if err != nil {
			return nil, err
		}
		var testCount int = 0
		for _, testCase := range categoryExpanded.TestCases {
			if testCase.IsActive {
				testCount++
			}
		}
		category.Description = categoryExpanded.Description
		category.ID = categoryExpanded.ID
		category.Name = fmt.Sprintf("%s(%d)", categoryExpanded.Name, testCount)
		category.TestCases = nil
		testCategories = append(testCategories, &category)
	}

	cursor.Close(context.TODO())

	return testCategories, nil
}

// Get returns a list of all test categories with simple name
func (r *testCategoryRepository) GetSimple() ([]*responses.TestCategory, error) {

	cursor, err := testCategoryCollection.Find(context.TODO(), bson.M{})

	if err != nil {
		return nil, err
	}
	var testCategories []*responses.TestCategory = []*responses.TestCategory{}
	for cursor.Next(context.TODO()) {
		var result responses.TestCategory
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		testCategories = append(testCategories, &result)
	}

	cursor.Close(context.TODO())

	return testCategories, nil
}

func (r *testCategoryRepository) GetByIds(categoriesId []primitive.ObjectID) ([]*responses.TestCategoryExpanded, error) {

	cursor, err := testCategoryCollection.Aggregate(context.TODO(), queries.GetCategoriesByIds(categoriesId))

	if err != nil {
		return nil, err
	}
	var testCategories []*responses.TestCategoryExpanded = []*responses.TestCategoryExpanded{}
	for cursor.Next(context.TODO()) {
		var category responses.TestCategoryExpanded
		err := cursor.Decode(&category)
		if err != nil {
			return nil, err
		}
		var testCases []responses.TestCase
		for _, testCase := range category.TestCases {
			if testCase.IsActive {
				testCases = append(testCases, testCase)
			}
		}
		category.TestCases = testCases
		testCategories = append(testCategories, &category)
	}

	return testCategories, nil

}
func (r *testCategoryRepository) GetOtherCategory() (*models.TestCategory, error) {

	var testCategory *models.TestCategory
	err := testCategoryCollection.FindOne(context.TODO(),
		queries.GetOtherCategory()).Decode(&testCategory)
	switch err {
	case mongo.ErrNoDocuments:
		return testCategory, err
	default:
		return testCategory, nil
	}
}
func (r *testCategoryRepository) CreateOtherCategory() (*models.TestCategory, error) {

	var otherCategory *models.TestCategory
	otherCategory.Name = utils.OtherCategory
	res, err := testCategoryCollection.InsertOne(context.TODO(), otherCategory)

	if err != nil {
		return otherCategory, err
	}

	otherCategory.ID = res.InsertedID.(primitive.ObjectID)

	return otherCategory, nil
}
