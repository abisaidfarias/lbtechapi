package services

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ITestCaseService is the test case service interface
type ITestCaseService interface {
	Create(*request.TestCase) error
	Get() ([]*models.TestCase, error)
	GetByID(string) (*models.TestCase, error)
	Update(string, *request.TestCase) error
	Upgrade(string, *request.TestCase) error
	Delete(string) error
}

type testCaseService struct {
	testCaseRepository repositories.ITestCaseRepository
}

// NewTestCaseService is a constructor
func NewTestCaseService(testCaseRepository repositories.ITestCaseRepository) ITestCaseService {
	return &testCaseService{
		testCaseRepository: testCaseRepository,
	}
}

// Create creates a new test case
func (s *testCaseService) Create(testCase *request.TestCase) error {
	testCase.IsActive = true

	testCaseModel, err := generateTestCase(testCase)

	if err != nil {
		return err
	}

	err = s.testCaseRepository.Create(testCaseModel)

	if err != nil {
		return err
	}

	return nil
}

// Get gets a list of test cases
func (s *testCaseService) Get() ([]*models.TestCase, error) {
	result, err := s.testCaseRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetById gets a case by id
func (s *testCaseService) GetByID(id string) (*models.TestCase, error) {

	testCase, err := s.testCaseRepository.GetByID(id)

	if err != nil {
		return nil, err
	}
	return testCase, nil
}

// Update updates a test case
func (s *testCaseService) Update(id string, testCase *request.TestCase) error {

	testCaseModel, err := generateTestCase(testCase)

	if err != nil {
		return err
	}

	err = s.update(id, testCaseModel)

	if err != nil {
		return err
	}
	return nil
}

// TODO check this behavior
func (s *testCaseService) update(id string, testCase *models.TestCase) error {

	err := s.testCaseRepository.Update(id, testCase)

	if err != nil {
		return err
	}
	return nil
}

// Upgrade upgrades a test case
func (s *testCaseService) Upgrade(id string, testCase *request.TestCase) error {

	// updating old testCase
	// TODO fix for updating without reading again from database
	updated, err := s.GetByID(id)

	if err != nil {
		return err
	}

	updated.IsActive = false

	err = s.update(id, updated)
	if err != nil {
		return err
	}
	// updating code
	testCase.Code = updateCodeVersion(testCase.Code)
	// insert new
	testCase.IsActive = true
	err = s.Create(testCase)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a test case
func (s *testCaseService) Delete(id string) error {
	err := s.testCaseRepository.Delete(id)

	if err != nil {
		return err
	}
	return nil
}

func generateTestCase(testCase *request.TestCase) (*models.TestCase, error) {

	objID, err := primitive.ObjectIDFromHex(testCase.CategoryID)

	if err != nil {
		return nil, err
	}

	return &models.TestCase{
		Code:         testCase.Code,
		Name:         testCase.Name,
		TestCategory: objID,
		IsActive:     testCase.IsActive,
		Description:  testCase.Description,
		Device:       testCase.Device,
		Expected:     testCase.Expected,
	}, nil
}

func updateCodeVersion(code string) string {
	size := len(code)
	char := []rune(code)[size-1]
	nextVersion := nextRune(char)
	res := code[:size-1] + string(nextVersion)
	return res
}

func nextRune(r rune) rune {
	return r + 1
}
