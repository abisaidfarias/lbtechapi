package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// ITestCategoryService is the testCategory service
type ITestCategoryService interface {
	Create(*request.TestCategory) (string, error)
	Get() ([]*responses.TestCategory, error)
	GetSimple() ([]*responses.TestCategory, error)
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
func (s *testCategoryService) Create(testCategoryRequest *request.TestCategory) (string, error) {

	testCategory, err := mapping.TestCategoryRequestToTestCategory(testCategoryRequest)
	if err != nil {
		return "", err
	}
	id, err := s.testCategoryRepository.Create(testCategory)

	if err != nil {
		return "", err
	}

	return id.Hex(), nil
}

// Get gets a list of all categories
func (s *testCategoryService) Get() ([]*responses.TestCategory, error) {
	result, err := s.testCategoryRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Get gets a list of all categories
func (s *testCategoryService) GetSimple() ([]*responses.TestCategory, error) {
	result, err := s.testCategoryRepository.GetSimple()

	if err != nil {
		return nil, err
	}

	return result, nil
}
