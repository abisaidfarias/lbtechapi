package services

import (
	"bufio"
	"encoding/csv"
	"io"
	"mime/multipart"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"gopkg.in/mgo.v2/bson"
)

// ITestCaseService is the test case service interface
type ITestCaseService interface {
	Create(*request.TestCase) error
	Get() ([]*bson.M, error)
	GetById(string) (*bson.M, error)
	Update(string, *request.TestCase) error
	Upgrade(string, *request.TestCase) error
	Delete(string) error
	ProcessFile(file *multipart.FileHeader) *responses.TestCaseFileUpload
}

type testCaseService struct {
	testCaseRepository  repositories.ITestCaseRepository
	testCategoryService ITestCategoryService
}

// NewTestCaseService is a constructor
func NewTestCaseService(testCaseRepository repositories.ITestCaseRepository, testCategoryService ITestCategoryService) ITestCaseService {
	return &testCaseService{
		testCaseRepository:  testCaseRepository,
		testCategoryService: testCategoryService,
	}
}

// Create creates a new test case
func (s *testCaseService) Create(testCaseRequest *request.TestCase) error {

	testCase, err := mapping.TestCaseRequestToTestCase(testCaseRequest, true)
	testCase.IsActive = true

	if err != nil {
		return err
	}

	err = s.testCaseRepository.Create(testCase)

	if err != nil {
		return err
	}

	return nil
}

// Get gets a list of test cases
func (s *testCaseService) Get() ([]*bson.M, error) {
	result, err := s.testCaseRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetById gets a case by id
func (s *testCaseService) GetById(id string) (*bson.M, error) {

	testCase, err := s.testCaseRepository.GetById(id)

	if err != nil {
		return nil, err
	}
	return testCase, nil
}

// Update updates a test case
func (s *testCaseService) Update(id string, testCaseRequest *request.TestCase) error {

	testCase, err := mapping.TestCaseRequestToTestCase(testCaseRequest, false)

	if err != nil {
		return err
	}
	err = s.testCaseRepository.Update(id, testCase)
	if err != nil {
		return err
	}
	return nil
}
func (s *testCaseService) Upgrade(id string, testCaseRequest *request.TestCase) error {

	testCaseRequest.Code = functions.UpdateCodeVersion(testCaseRequest.Code)
	err := s.Create(testCaseRequest)
	if err != nil {
		return err
	}
	testCaseRequest.IsActive = false
	err = s.Update(id, testCaseRequest)
	if err != nil {
		return err
	}

	return nil
}

func (s *testCaseService) Delete(id string) error {
	err := s.testCaseRepository.Delete(id)

	if err != nil {
		return err
	}
	return nil
}

func (s *testCaseService) ProcessFile(fileHeader *multipart.FileHeader) *responses.TestCaseFileUpload {
	file, err := fileHeader.Open()

	response := responses.TestCaseFileUpload{}
	errors := []int{}
	i := 0
	if err != nil {
		return &response
	}

	categories, err := s.testCategoryService.Get()
	if err != nil {
		// TOOD handle error properly
		panic(err)
	}

	catMap := generateCategoryMap(categories)

	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))

	for {

		line, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			errors = append(errors, i)
		} else {
			err := processLine(line, catMap)
			if err != nil {
				errors = append(errors, i)
			}
		}

		i++
	}

	response.InvalidRows = errors
	response.TotalRows = i - len(errors)

	return &response
}

func processLine(line []string, catMap *map[string]string) error {

	// generate testcase here
	// validate category
	// insert in db

	return nil
}

func generateCategoryMap(categories []*responses.TestCategory) *map[string]string {
	m := make(map[string]string)

	for _, v := range categories {
		m[v.Name] = v.ID.Hex()
	}

	return &m
}
