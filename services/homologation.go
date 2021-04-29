package services

import (
	"fmt"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IHomologationService is the homologation service
type IHomologationService interface {
	Create(*request.Homologation) (*models.CustomError, error)
	Get(string) ([]*responses.HomologationExpanded, error)
	GetReport(string) (*responses.HomologationReport, error)
	GetCategoriesWithTest(string) (map[string]responses.CategoryExpanded, error)
}

type homologationService struct {
	homologationRepository repositories.IHomologationRepository
	testCategoryRepository repositories.ITestCategoryRepository
	userRepository         repositories.IUserRepository
}

// NewHomologationService is a constructor
func NewHomologationService(homologationRepository repositories.IHomologationRepository,
	testCategoryRepository repositories.ITestCategoryRepository,
	userRepository repositories.IUserRepository) IHomologationService {
	return &homologationService{
		homologationRepository: homologationRepository,
		testCategoryRepository: testCategoryRepository,
		userRepository:         userRepository,
	}
}

// Create creates a new cateogry
func (s *homologationService) Create(homologationRequest *request.Homologation) (*models.CustomError, error) {

	companyID, _ := primitive.ObjectIDFromHex(homologationRequest.Company)
	deviceID, _ := primitive.ObjectIDFromHex(homologationRequest.Device)
	countryID, _ := primitive.ObjectIDFromHex(homologationRequest.Country)
	testPlanID, _ := primitive.ObjectIDFromHex(homologationRequest.TestPlan)
	brandID, _ := primitive.ObjectIDFromHex(homologationRequest.Brand)

	existHomologation, err := s.homologationRepository.GetPrevious(deviceID,
		countryID, companyID)
	if existHomologation != nil || err != nil {
		var customeErr models.CustomError
		customeErr.Code = utils.HomologationExistCode
		customeErr.Err = utils.HomologationExist
		return &customeErr, nil
	}

	if response, err := utils.ValidateHomologationRequest(existHomologation,
		homologationRequest); response == false {
		return err, nil
	}
	ids := functions.StringsToObjectIds(homologationRequest.TestCategories)
	categories, err := s.testCategoryRepository.GetByIds(ids)
	if err != nil {
		return nil, err
	}
	homologation := mapping.HomologationRequestToHomologation(homologationRequest,
		categories, companyID, deviceID, countryID, testPlanID, brandID)

	err = s.homologationRepository.Create(homologation)

	if err != nil {
		return nil, err
	}
	return nil, nil
}
func (s *homologationService) Get(userID string) ([]*responses.HomologationExpanded, error) {
	user, err := s.userRepository.GetByID(userID)

	if err != nil {
		return nil, err
	}
	if user.IsInternal {
		homologations, err := s.homologationRepository.GetByInternal(user.Clients,
			user.Brands, user.Countries)
		if err != nil {
			return nil, err
		}
		return homologations, nil
	} else {
		homologations, err := s.homologationRepository.GetByExternal(user.Company,
			user.Brands, user.Countries)
		if err != nil {
			return nil, err
		}
		return homologations, nil
	}

}
func (s *homologationService) GetReport(id string) (*responses.HomologationReport, error) {

	homologationID, _ := primitive.ObjectIDFromHex(id)
	homologation, err := s.homologationRepository.GetByID(homologationID)

	if err != nil || homologation == nil {
		return nil, err
	}
	categoriesGrouped := make(map[string]responses.CategoryResult)

	for _, t := range homologation.TestResults {

		value, _ := categoriesGrouped[t.TestCategory.Name]

		if t.Result == enums.TestResult_value["FAIL"] {
			value.Fail++
		} else if t.Result == enums.TestResult_value["PASS"] {
			value.Pass++
		} else if t.Result == enums.TestResult_value["NA"] {
			value.NA++
		} else {
			value.NoRun++
		}
		var testCaseResume responses.TestCaseResume
		testCaseResume.Code = t.Code
		testCaseResume.Name = t.Name
		testCaseResume.Result = t.Result
		value.TestCaseResume = append(value.TestCaseResume, testCaseResume)
		categoriesGrouped[t.TestCategory.Name] = value
	}
	var homologationReport responses.HomologationReport
	homologationReport.Categories = categoriesGrouped
	return &homologationReport, nil

}
func (s *homologationService) GetCategoriesWithTest(id string) (map[string]responses.CategoryExpanded, error) {

	homologationID, _ := primitive.ObjectIDFromHex(id)
	homologation, err := s.homologationRepository.GetByID(homologationID)

	if err != nil || homologation == nil {
		return nil, err
	}
	categoriesGrouped := make(map[string]responses.CategoryExpanded)

	for _, t := range homologation.TestResults {

		value, _ := categoriesGrouped[t.TestCategory.Name]

		if t.Result == enums.TestResult_value["FAIL"] {
			value.Fail++
		} else if t.Result == enums.TestResult_value["PASS"] {
			value.Pass++
		} else if t.Result == enums.TestResult_value["NA"] {
			value.NA++
		} else {
			value.NoRun++
		}
		var testResultResume responses.TestResultResume
		testResultResume.Code = t.Code
		testResultResume.Name = fmt.Sprintf("%s %s", t.Code, t.Name)
		testResultResume.Result = t.Result
		testResultResume.Description = t.Description
		testResultResume.Expected = t.Expected
		testResultResume.Hyperlinks = t.Hyperlinks
		testResultResume.Images = t.Images
		testResultResume.IssueDescription = t.IssueDescription
		testResultResume.Value = t.Value

		value.TestResultResume = append(value.TestResultResume, testResultResume)
		categoriesGrouped[t.TestCategory.Name] = value
	}
	return categoriesGrouped, nil

}
