package services

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ITestCategoryService is the testCategory service
type ITestCategoryService interface {
	Create(*request.TestCategory) (string, error)
	Get() ([]*bson.M, error)
}

type testCategoryService struct {
	testCategoryRepository repositories.ITestCategoryRepository
}

// NewTestCategoryService is a constructor
func NewTestCategoryService(testCategoryRepository repositories.ITestCategoryRepository) ITestCategoryService {
	return &testCategoryService{
		testCategoryRepository: testCategoryRepository,
	}
}

// Create creates a new cateogry
func (s *testCategoryService) Create(category *request.TestCategory) (string, error) {

	categoryModel := buildCategory(category)

	id, err := s.testCategoryRepository.Create(categoryModel)

	if err != nil {
		return "", err
	}

	return id.Hex(), nil
}

// Get gets a list of all categories
func (s *testCategoryService) Get() ([]*bson.M, error) {
	result, err := s.testCategoryRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func buildCategory(reqCategory *request.TestCategory) *models.TestCategory {

	return &models.TestCategory{
		Name:        reqCategory.Name,
		Description: reqCategory.Description,
		TestCases:   []primitive.ObjectID{},
	}
}
