package services

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
)

// ITestCaseService is the test case service interface
type ITestCaseService interface {
	Create(*models.TestCase) error
	Get() ([]*models.TestCase, error)
	GetByID(string) (*models.TestCase, error)
	Update(string, *models.TestCase) error
	Upgrade(string, *models.TestCase) error
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
func (s *testCaseService) Create(testCase *models.TestCase) error {

	err := s.testCaseRepository.Create(testCase)

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
func (s *testCaseService) Update(id string, testCase *models.TestCase) error {
	err := s.testCaseRepository.Update(id, testCase)

	if err != nil {
		return err
	}
	return nil
}

// Upgrade upgrades a test case
func (s *testCaseService) Upgrade(id string, testCase *models.TestCase) error {

	// updating old testCase
	testCase.IsActive = false
	err := s.Update(id, testCase)
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

func updateCodeVersion(code string) string {
	// TODO implement
	return code
}
