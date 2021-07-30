package services

import (
	"bufio"
	"encoding/csv"
	"io"
	"mime/multipart"
	"strings"

	"github.com/abisaidfarias/lbtechapi/repositories"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
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
	ProcessFile(file *multipart.FileHeader) (*responses.TestCaseFileUpload, error)
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

func (s *testCaseService) ProcessFile(fileHeader *multipart.FileHeader) (*responses.TestCaseFileUpload, error) {

	file, err := fileHeader.Open()

	if err != nil {
		return nil, err
	}

	defer file.Close()

	response := responses.TestCaseFileUpload{}
	errors := []int{}
	lineIndex := 1

	categories, err := s.testCategoryService.GetSimple()

	if err != nil {
		return nil, err
	}

	catMap := functions.GenerateCategoryMap(categories)

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';'
	for {

		line, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			// unable to read line
			errors = append(errors, lineIndex)
		} else {

			err := s.ProcessLine(line, catMap)

			if err != nil {
				// error with line data
				errors = append(errors, lineIndex)
			}
		}

		lineIndex++
	}

	response.InvalidRows = errors
	response.TotalRows = lineIndex - len(errors) - 1

	return &response, nil
}

func (s *testCaseService) ProcessLine(line []string, catMap map[string]string) error {
	if len(line) != 5 {
		// invalid csv columns length
		return utils.ErrorInvalidLineFormat
	}

	if line[4] == "" {
		// category is empty
		return utils.ErrorInvalidLineFormat
	}
	catName := strings.Trim(line[4], " ")

	catId, ok := catMap[catName]
	if !ok {
		// category does not exist
		return utils.ErrorInvalidLineFormat
	}

	testCase := functions.GenerateTestCaseFromLine(line, catId)

	if testCase == nil {
		// failed to build test case from line
		return utils.ErrorInvalidLineFormat
	}

	err := s.testCaseRepository.Upsert(testCase.Code, testCase)

	if err != nil {
		// failed to upsert the test case
		return err
	}

	return nil
}
