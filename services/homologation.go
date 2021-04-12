package services

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
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
