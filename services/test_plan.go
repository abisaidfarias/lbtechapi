package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// ITestPlanService is the test Plan service interface
type ITestPlanService interface {
	Create(*request.TestPlan) error
	Get() ([]*responses.TestPlanExpanded, error)
	GetById(string) (*responses.TestPlanExpanded, error)
	Update(string, *request.TestPlan) error
	Delete(string) error
}

type testPlanService struct {
	testPlanRepository repositories.ITestPlanRepository
}

// NewTestPlanService is a constructor
func NewTestPlanService(testPlanRepository repositories.ITestPlanRepository) ITestPlanService {
	return &testPlanService{
		testPlanRepository: testPlanRepository,
	}
}

// Create creates a new test Plan
func (s *testPlanService) Create(testPlanRequest *request.TestPlan) error {

	testPlan, err := mapping.TestPlanRequestToTestPlan(testPlanRequest)
	testPlan.IsActive = true

	if err != nil {
		return err
	}

	err = s.testPlanRepository.Create(testPlan, testPlanRequest.UserID)

	if err != nil {
		return err
	}

	return nil
}

// Get gets a list of test Plans
func (s *testPlanService) Get() ([]*responses.TestPlanExpanded, error) {
	result, err := s.testPlanRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetById gets a Plan by id
func (s *testPlanService) GetById(id string) (*responses.TestPlanExpanded, error) {

	testPlan, err := s.testPlanRepository.GetById(id)

	if err != nil {
		return nil, err
	}
	return testPlan, nil
}

// Update updates a test Plan
func (s *testPlanService) Update(id string, testPlanRequest *request.TestPlan) error {

	testPlan, err := mapping.TestPlanRequestToTestPlan(testPlanRequest)

	if err != nil {
		return err
	}
	err = s.testPlanRepository.Update(id, testPlan)
	if err != nil {
		return err
	}
	return nil
}
func (s *testPlanService) Delete(id string) error {
	err := s.testPlanRepository.Delete(id)
	if err != nil {
		return err
	}
	return nil
}
