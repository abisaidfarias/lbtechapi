package services

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IHomologationService is the homologation service
type IHomologationService interface {
	Create(*request.Homologation) (*models.CustomError, error)
	// Get() ([]*responses.Homologation, error)
}

type homologationService struct {
	homologationRepository repositories.IHomologationRepository
	testCategoryRepository repositories.ITestCategoryRepository
}

// NewHomologationService is a constructor
func NewHomologationService(homologationRepository repositories.IHomologationRepository,
	testCategoryRepository repositories.ITestCategoryRepository) IHomologationService {
	return &homologationService{
		homologationRepository: homologationRepository,
		testCategoryRepository: testCategoryRepository,
	}
}

// Create creates a new cateogry
func (s *homologationService) Create(homologationRequest *request.Homologation) (*models.CustomError, error) {

	companyId, _ := primitive.ObjectIDFromHex(homologationRequest.Company)
	deviceId, _ := primitive.ObjectIDFromHex(homologationRequest.Device)
	countryId, _ := primitive.ObjectIDFromHex(homologationRequest.Country)

	existHomologation, err := s.homologationRepository.GetHomologationByValidations(deviceId,
		countryId, companyId)
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
		categories, companyId, deviceId, countryId)

	err = s.homologationRepository.Create(homologation)

	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Get gets a list of all categories
// func (s *homologationService) Get() ([]*responses.Homologation, error) {
// 	result, err := s.homologationRepository.Get()

// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }
