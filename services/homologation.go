package services

import (
	"bytes"
	"fmt"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IHomologationService is the homologation service
type IHomologationService interface {
	Create(*request.Homologation) (*models.CustomError, error)
	Get(string) ([]*responses.HomologationExpanded, error)
	GetReport(string) (*responses.HomologationReport, error)
	GetCategoriesWithTest(string) (map[string]responses.CategoryExpanded, error)
	UpdateTestResult(string, request.TestResultResume) error
	PhaseChange(string, *request.HomologationResume) error
	GetHomologationFails(string) (*responses.TestFails, error)
	CreateFailTestResult(string, *request.TestResultResume) error
	UpdateDocument(string, *request.Homologation) error
	Update(string, *request.Homologation) error
	Delete(string) error
	ExportHomologation(string) (bytes.Buffer, error)
	ExportFailTest(string) (bytes.Buffer, error)
	UpdateFailTest(string, []request.TestResult) error
	HomologationNotification(*request.Homologation, primitive.ObjectID, string)
}

type homologationService struct {
	homologationRepository repositories.IHomologationRepository
	testCategoryRepository repositories.ITestCategoryRepository
	userRepository         repositories.IUserRepository
	notificationRepository repositories.INotificationRepository
	brandRepository        repositories.IBrandRepository
	deviceRepository       repositories.IDeviceRepository
	countryRepository      repositories.ICountryRepository
}

// NewHomologationService is a constructor
func NewHomologationService(homologationRepository repositories.IHomologationRepository,
	testCategoryRepository repositories.ITestCategoryRepository,
	userRepository repositories.IUserRepository,
	notificationRepository repositories.INotificationRepository,
	brandRepository repositories.IBrandRepository,
	deviceRepository repositories.IDeviceRepository,
	countryRepository repositories.ICountryRepository) IHomologationService {
	return &homologationService{
		homologationRepository: homologationRepository,
		testCategoryRepository: testCategoryRepository,
		userRepository:         userRepository,
		notificationRepository: notificationRepository,
		brandRepository:        brandRepository,
		deviceRepository:       deviceRepository,
		countryRepository:      countryRepository,
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
		countryID, companyID, homologationRequest.IsInternalProject)
	if err != nil {
		return nil, err
	}

	if response, err := utils.ValidateHomologationRequest(existHomologation,
		homologationRequest); !response {
		return err, nil
	}
	ids := functions.StringsToObjectIds(homologationRequest.TestCategories)
	var categories []*responses.TestCategoryExpanded = []*responses.TestCategoryExpanded{}
	if len(ids) > 0 {
		categories, err = s.testCategoryRepository.GetByIds(ids)
		if err != nil {
			return nil, err
		}
	}
	homologation := mapping.HomologationRequestToHomologation(homologationRequest,
		categories, companyID, deviceID, countryID, testPlanID, brandID)

	err = s.homologationRepository.Create(homologation)

	if err != nil {
		return nil, err
	}

	go s.HomologationNotification(homologationRequest, companyID, utils.CREATE)

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
	homologation, err := s.homologationRepository.GetByIdExpanded(homologationID)

	if err != nil || homologation == nil {
		return nil, err
	}
	categoriesGrouped := make(map[string]responses.CategoryResult)

	for _, t := range homologation.TestResults {

		value := categoriesGrouped[t.TestCategory.Name]

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
	homologation, err := s.homologationRepository.GetByIdExpanded(homologationID)

	if err != nil || homologation == nil {
		return nil, err
	}
	categoriesGrouped := make(map[string]responses.CategoryExpanded)

	for _, t := range homologation.TestResults {

		value := categoriesGrouped[t.TestCategory.Name]

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
		testResultResume.Name = t.Name
		testResultResume.Result = t.Result
		testResultResume.Description = t.Description
		testResultResume.Expected = t.Expected
		testResultResume.Hyperlinks = t.Hyperlinks
		testResultResume.Images = t.Images
		testResultResume.OverviewIssue = t.OverviewIssue
		testResultResume.ActualResult = t.ActualResult
		testResultResume.StepsToReproduce = t.StepsToReproduce
		testResultResume.ExpectedResult = t.ExpectedResult
		testResultResume.IssueSeverity = t.IssueSeverity
		testResultResume.IssueFrequency = t.IssueFrequency
		testResultResume.Value = t.Value

		value.TestResultResume = append(value.TestResultResume, testResultResume)
		categoriesGrouped[t.TestCategory.Name] = value
	}
	return categoriesGrouped, nil

}

// Update updates a test case
func (s *homologationService) UpdateTestResult(id string, testResultRequest request.TestResultResume) error {

	err := s.homologationRepository.UpdateTestResult(id, testResultRequest)
	if err != nil {
		return err
	}
	return nil
}
func (s *homologationService) PhaseChange(id string, homologationRequest *request.HomologationResume) error {

	homologation := mapping.HomologationRequestToHomologationResume(homologationRequest)

	err := s.homologationRepository.PhaseChange(id, homologation)
	if err != nil {
		return err
	}
	homologationId, _ := primitive.ObjectIDFromHex(id)
	homologationResponse, _ := s.homologationRepository.GetByIdExpanded(homologationId)
	homologationR := mapping.HomologationResponseToHomologationRequest(*homologationResponse)
	homologationR.Status = homologation.Status
	go s.HomologationNotification(&homologationR, homologationResponse.Company.ID, utils.PHASE)

	return nil
}
func (s *homologationService) GetHomologationFails(id string) (*responses.TestFails, error) {

	homologationID, _ := primitive.ObjectIDFromHex(id)
	homologation, err := s.homologationRepository.GetByIdExpanded(homologationID)
	if err != nil {
		return nil, err
	}

	var testResults []responses.TestResult = []responses.TestResult{}
	testFails := new(responses.TestFails)
	for _, test := range homologation.TestResults {
		testFails.TotalTest++
		if test.Result == enums.TestResult_value["FAIL"] {
			if test.IssueSeverity == enums.TestFailureSeverity_value["HIGH"] {
				testFails.TotalHigh++
			} else if test.IssueSeverity == enums.TestFailureSeverity_value["MEDIUM"] {
				testFails.TotalMedium++
			} else {
				testFails.TotalLow++
			}
			testResults = append(testResults, test)
		}
	}
	testFails.TestResults = testResults
	return testFails, nil

}
func (s *homologationService) CreateFailTestResult(id string, testResultRequest *request.TestResultResume) error {

	failCategory, err := s.testCategoryRepository.GetOtherCategory()
	if err != nil {
		failCategory, err = s.testCategoryRepository.CreateOtherCategory()
		if err != nil {
			return err
		}
	}

	if failCategory.ID == primitive.NilObjectID {
		failCategory, err = s.testCategoryRepository.CreateOtherCategory()
		if err != nil {
			return err
		}
	}

	testResult := mapping.TestResultRequestToTestResult(testResultRequest)
	testResult.TestCategory = *failCategory

	err = s.homologationRepository.CreateFailTestResult(id, testResult)
	if err != nil {
		return err
	}

	homologationId, _ := primitive.ObjectIDFromHex(id)
	homologation, _ := s.homologationRepository.GetByIdExpanded(homologationId)
	go s.FailNotification(homologation, *testResultRequest, utils.CREATE)
	return nil
}
func (s *homologationService) UpdateDocument(id string, homologation *request.Homologation) error {

	homologationId, _ := primitive.ObjectIDFromHex(id)
	err := s.homologationRepository.UpdateDocument(homologation.DocumentUrl, homologationId)
	if err != nil {
		return err
	}
	return nil
}
func (s *homologationService) Update(id string, homologationRequest *request.Homologation) error {
	homologationId, _ := primitive.ObjectIDFromHex(id)
	homologation := mapping.HomologationRequestToHomologationUpdate(homologationRequest)

	err := s.homologationRepository.Update(homologationId, homologation)
	if err != nil {
		return err
	}
	//go functions.SendNotifications(homologation.Company, false, "PRUEBA")
	return nil
}
func (s *homologationService) Delete(id string) error {
	homologationId, _ := primitive.ObjectIDFromHex(id)

	homologation, err := s.homologationRepository.GetByIdExpanded(homologationId)
	if err != nil {
		return err
	}
	if homologation.Type == enums.HomologationType_value["INITIAL"] {
		err = s.homologationRepository.DeleteHierarchy(homologation.Device.ID, homologation.Country.ID,
			homologation.Company.ID, homologation.IsInternalProject)
		if err != nil {
			return err
		}
		return nil
	}
	err = s.homologationRepository.Delete(homologationId)
	if err != nil {
		return err
	}
	return nil
}
func (s *homologationService) ExportHomologation(userId string) (bytes.Buffer, error) {
	user, err := s.userRepository.GetByID(userId)
	var b bytes.Buffer
	if err != nil {
		return b, err
	}
	var homologations []*responses.HomologationExpanded = []*responses.HomologationExpanded{}
	if user.IsInternal {
		homologations, err = s.homologationRepository.GetByInternal(user.Clients,
			user.Brands, user.Countries)
		if err != nil {
			return b, err
		}
	} else {
		homologations, err = s.homologationRepository.GetByExternal(user.Company,
			user.Brands, user.Countries)
		if err != nil {
			return b, err
		}
	}
	file, err := exportHomologationFile(homologations)
	if err != nil {
		return file, err
	}
	return file, nil
}
func exportHomologationFile(homologations []*responses.HomologationExpanded) (bytes.Buffer, error) {
	file := excelize.NewFile()
	categories := enums.ExcelHomologationHeaders
	for k, v := range categories {
		file.SetCellValue(utils.PAGE, k, v)
	}
	for index, h := range homologations {
		cell, _ := excelize.CoordinatesToCellName(1, index+2)
		file.SetCellValue(utils.PAGE, cell, h.Company.Name)
		cell, _ = excelize.CoordinatesToCellName(2, index+2)
		file.SetCellValue(utils.PAGE, cell, h.Country.Name)
		cell, _ = excelize.CoordinatesToCellName(3, index+2)
		file.SetCellValue(utils.PAGE, cell, h.Brand.Name)
		cell, _ = excelize.CoordinatesToCellName(4, index+2)
		file.SetCellValue(utils.PAGE, cell, h.Device.CommercialModel)
		cell, _ = excelize.CoordinatesToCellName(5, index+2)
		file.SetCellValue(utils.PAGE, cell, h.Device.TechnicalModel)
		cell, _ = excelize.CoordinatesToCellName(6, index+2)
		file.SetCellValue(utils.PAGE, cell, h.OsVersion)
		cell, _ = excelize.CoordinatesToCellName(7, index+2)
		file.SetCellValue(utils.PAGE, cell, h.ApprovalType)
		cell, _ = excelize.CoordinatesToCellName(8, index+2)
		file.SetCellValue(utils.PAGE, cell, h.TestPlan.Name)
		if h.TestStartDate != nil {
			year, month, day := h.TestStartDate.Date()
			cell, _ = excelize.CoordinatesToCellName(9, index+2)
			file.SetCellValue(utils.PAGE, cell, fmt.Sprintf("%d/%d/%d", day, month, year))
		}
		if h.TestEndDate != nil {
			year, month, day := h.TestStartDate.Date()
			cell, _ = excelize.CoordinatesToCellName(10, index+2)
			file.SetCellValue(utils.PAGE, cell, fmt.Sprintf("%d/%d/%d", day, month, year))
		}
		cell, _ = excelize.CoordinatesToCellName(11, index+2)
		file.SetCellValue(utils.PAGE, cell, h.ProjectType)
		cell, _ = excelize.CoordinatesToCellName(12, index+2)
		file.SetCellValue(utils.PAGE, cell, h.StatusView)
	}

	var b bytes.Buffer
	if err := file.Write(&b); err != nil {
		return b, err
	}
	return b, nil
}
func (s *homologationService) ExportFailTest(id string) (bytes.Buffer, error) {

	homologationID, _ := primitive.ObjectIDFromHex(id)
	homologation, err := s.homologationRepository.GetByIdExpanded(homologationID)
	var b bytes.Buffer
	if err != nil {
		return b, err
	}
	file, err := exportFailsFile(homologation)
	if err != nil {
		return file, err
	}
	return file, nil
}
func exportFailsFile(homologation *responses.HomologationExpanded) (bytes.Buffer, error) {
	file := excelize.NewFile()
	categories := enums.ExcelFailsHeaders

	cell, _ := excelize.CoordinatesToCellName(1, 1)
	file.SetCellValue(utils.PAGE, cell, utils.COMPANY)
	cell, _ = excelize.CoordinatesToCellName(2, 1)
	file.SetCellValue(utils.PAGE, cell, homologation.Company.Name)

	cell, _ = excelize.CoordinatesToCellName(1, 2)
	file.SetCellValue(utils.PAGE, cell, utils.SOFTWARE_VERSION)
	cell, _ = excelize.CoordinatesToCellName(2, 2)
	file.SetCellValue(utils.PAGE, cell, homologation.SoftwareVersion)

	cell, _ = excelize.CoordinatesToCellName(1, 3)
	file.SetCellValue(utils.PAGE, cell, utils.TECHNICAL_MODEL)
	cell, _ = excelize.CoordinatesToCellName(2, 3)
	file.SetCellValue(utils.PAGE, cell, homologation.Device.TechnicalModel)

	cell, _ = excelize.CoordinatesToCellName(1, 4)
	file.SetCellValue(utils.PAGE, cell, utils.BRAND)
	cell, _ = excelize.CoordinatesToCellName(2, 4)
	file.SetCellValue(utils.PAGE, cell, homologation.Brand.Name)

	cell, _ = excelize.CoordinatesToCellName(1, 5)
	file.SetCellValue(utils.PAGE, cell, utils.COUNTRY)
	cell, _ = excelize.CoordinatesToCellName(2, 5)
	file.SetCellValue(utils.PAGE, cell, homologation.Country.Name)

	for k, v := range categories {
		file.SetCellValue(utils.PAGE, k, v)
	}
	index := 7
	for _, t := range homologation.TestResults {
		if t.Result == enums.TestResult_value["FAIL"] {
			cell, _ := excelize.CoordinatesToCellName(1, index)
			file.SetCellValue(utils.PAGE, cell, t.Code)
			cell, _ = excelize.CoordinatesToCellName(2, index)
			file.SetCellValue(utils.PAGE, cell, t.OverviewIssue)
			cell, _ = excelize.CoordinatesToCellName(3, index)
			file.SetCellValue(utils.PAGE, cell, t.ActualResult)
			cell, _ = excelize.CoordinatesToCellName(4, index)
			file.SetCellValue(utils.PAGE, cell, t.ExpectedResult)
			cell, _ = excelize.CoordinatesToCellName(5, index)
			file.SetCellValue(utils.PAGE, cell, t.StepsToReproduce)
			cell, _ = excelize.CoordinatesToCellName(6, index)
			file.SetCellValue(utils.PAGE, cell, enums.TestFailureFrequency_key[t.IssueFrequency])
			cell, _ = excelize.CoordinatesToCellName(7, index)
			file.SetCellValue(utils.PAGE, cell, enums.TestFailureSeverity_key[t.IssueSeverity])
			var hyperlinks string
			for _, link := range t.Hyperlinks {
				hyperlinks = hyperlinks + "," + link.Link
			}
			cell, _ = excelize.CoordinatesToCellName(8, index)
			file.SetCellValue(utils.PAGE, cell, hyperlinks)
			index++
		}
	}

	var b bytes.Buffer
	if err := file.Write(&b); err != nil {
		return b, err
	}
	return b, nil
}

func (s *homologationService) UpdateFailTest(id string, testResultRequests []request.TestResult) error {
	homologationId, _ := primitive.ObjectIDFromHex(id)
	testResults := mapping.TestResultsRequestToTestResults(testResultRequests)

	err := s.homologationRepository.UpdateFailTest(homologationId, testResults)
	if err != nil {
		return err
	}
	return nil
}

func (s *homologationService) HomologationNotification(homologation *request.Homologation,
	companyId primitive.ObjectID, key string) {

	toList, isEmpty := functions.GetEmails(false, companyId)
	if isEmpty {
		return
	}
	country, err := s.countryRepository.GetById(homologation.Country)
	if err != nil {
		return
	}
	device, err := s.deviceRepository.GetById(homologation.Device)
	if err != nil {
		return
	}
	brand, err := s.brandRepository.GetById(homologation.Brand)
	if err != nil {
		return
	}
	projectType := "External"
	if homologation.IsInternalProject {
		projectType = "Internal"
	}
	planningDate := "N/A"
	if !homologation.PlanningDate.IsZero() {

		planningDate = fmt.Sprintf("%02d/%02d/%d",
			homologation.PlanningDate.Day(), homologation.PlanningDate.Month(),
			homologation.PlanningDate.Year())
	}
	sampleStartDate := "N/A"
	if !homologation.SampleStartDate.IsZero() {
		sampleStartDate = fmt.Sprintf("%02d/%02d/%d",
			homologation.SampleStartDate.Day(), homologation.SampleStartDate.Month(),
			homologation.SampleStartDate.Year())
	}
	sampleEndDate := "N/A"
	if !homologation.SampleEndDate.IsZero() {
		sampleEndDate = fmt.Sprintf("%02d/%02d/%d",
			homologation.SampleEndDate.Day(), homologation.SampleEndDate.Month(),
			homologation.SampleEndDate.Year())
	}
	testStartDate := "N/A"
	if !homologation.TestStartDate.IsZero() {
		testStartDate = fmt.Sprintf("%02d/%02d/%d",
			homologation.TestStartDate.Day(), homologation.TestStartDate.Month(),
			homologation.TestStartDate.Year())
	}
	testEndDate := "N/A"
	if !homologation.TestEndDate.IsZero() {
		testEndDate = fmt.Sprintf("%02d/%02d/%d",
			homologation.TestEndDate.Day(), homologation.TestEndDate.Month(),
			homologation.TestEndDate.Year())
	}
	underStartDate := "N/A"
	if !homologation.UnderStartDate.IsZero() {
		underStartDate = fmt.Sprintf("%02d/%02d/%d",
			homologation.UnderStartDate.Day(), homologation.UnderStartDate.Month(),
			homologation.UnderStartDate.Year())
	}
	underEndDate := "N/A"
	if !homologation.UnderEndDate.IsZero() {
		underEndDate = fmt.Sprintf("%02d/%02d/%d",
			homologation.UnderEndDate.Day(), homologation.UnderEndDate.Month(),
			homologation.UnderEndDate.Year())
	}
	resultDate := "N/A"
	if !homologation.CompletedDate.IsZero() {
		resultDate = fmt.Sprintf("%02d/%02d/%d",
			homologation.CompletedDate.Day(), homologation.CompletedDate.Month(),
			homologation.CompletedDate.Year())
	}
	var subject string
	var mainMessage string

	switch key {

	case utils.CREATE:
		subject = fmt.Sprintf("Subject: New homologation process was created %s %s %s",
			country.Name, brand.Name, device.CommercialModel)
		mainMessage = utils.CREATE_MAIN_MESSAGE
	case utils.PHASE:
		mainMessage, subject = functions.GetNotificationMessageAndSubject(homologation, country.Name,
			brand.Name, device.CommercialModel)
	default:
		return
	}

	body, err := functions.GetHomologationBodyMessage(subject, mainMessage, projectType, brand.Name,
		device.TechnicalModel, device.CommercialModel,
		homologation.SoftwareVersion, homologation.OsVersion,
		country.Name, homologation.Type, homologation.Carrier,
		homologation.TestingType, planningDate, sampleStartDate,
		sampleEndDate, testStartDate, testEndDate, underStartDate,
		underEndDate, resultDate, utils.TEMPLATE_HOMOLOGATION_PATH)

	if err != nil {
		return
	}
	functions.SendNotifications(toList, body)
}
func (s *homologationService) FailNotification(homologation *responses.HomologationExpanded, failtTest request.TestResultResume, key string) {

	toList, isEmpty := functions.GetEmails(false, homologation.Company.ID)
	if isEmpty {
		return
	}
	projectType := "External"
	if homologation.IsInternalProject {
		projectType = "Internal"
	}
	var subject string
	var mainMessage string

	switch key {

	case utils.CREATE:
		subject = fmt.Sprintf("Subject: New Issue has been created for %s %s",
			homologation.Brand.Name, homologation.Device.CommercialModel)
		mainMessage = utils.CREATE_FAIL_MAIN_MESSAGE
	default:
		return
	}

	body, err := functions.GetFailBodyMessage(subject, mainMessage, homologation.Company.Name,
		homologation.Country.Name, homologation.Brand.Name,
		homologation.Device.TechnicalModel, homologation.Device.CommercialModel,
		homologation.SoftwareVersion, homologation.OsVersion,
		homologation.Type, projectType, failtTest.Code, failtTest.Name, failtTest.OverviewIssue, failtTest.ActualResult,
		failtTest.StepsToReproduce, failtTest.ExpectedResult, failtTest.IssueFrequency,
		failtTest.IssueSeverity, "hiperlinks", "description", utils.TEMPLATE_FAIL_PATH)

	if err != nil {
		return
	}
	functions.SendNotifications(toList, body)
}
