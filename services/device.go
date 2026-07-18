package services

import (
	"fmt"
	"strings"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"

	//"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IDeviceService is the test case service interface
type IDeviceService interface {
	Create(*request.Device, string) (*responses.DeviceExpanded, error)
	Get(string) ([]*responses.DeviceExpanded, error)
	GetById(string, string) (*responses.Device, error)
	Update(string, *request.Device, string) error
	Delete(string) (bool, error)
}

type deviceService struct {
	deviceRepository         repositories.IDeviceRepository
	deviceTrackingRepository repositories.IDeviceTrackingRepository
	homologationRepository   repositories.IHomologationRepository
	brandRepository          repositories.IBrandRepository
	userRepository           repositories.IUserRepository
}

// NewDeviceService is a constructor
func NewDeviceService(deviceRepository repositories.IDeviceRepository,
	deviceTrackingRepository repositories.IDeviceTrackingRepository,
	homologationRepository repositories.IHomologationRepository,
	brandRepository repositories.IBrandRepository,
	userRepository repositories.IUserRepository) IDeviceService {
	return &deviceService{
		deviceRepository:         deviceRepository,
		deviceTrackingRepository: deviceTrackingRepository,
		homologationRepository:   homologationRepository,
		brandRepository:          brandRepository,
		userRepository:           userRepository,
	}
}

// Create creates a new test case
func (s *deviceService) Create(deviceRequest *request.Device, userID string) (*responses.DeviceExpanded, error) {
	if err := s.ensureDeviceModelsUnique("", deviceRequest.TechnicalModel, deviceRequest.CommercialModel); err != nil {
		return nil, err
	}

	device := mapping.DeviceRequestToDevice(deviceRequest)

	deviceResponse, err := s.deviceRepository.Create(device)

	if err != nil {
		return nil, err
	}
	go s.DeviceNotification(*deviceRequest, utils.CREATE, userID)
	return deviceResponse, nil
}

// Get gets devices visible to the user.
func (s *deviceService) Get(userID string) ([]*responses.DeviceExpanded, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	var brands []primitive.ObjectID
	if !user.IsInternal {
		brands = user.Brands
	}

	result, err := s.deviceRepository.Get(brands)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetById gets a device by id when the user is allowed to see its brand.
func (s *deviceService) GetById(id string, userID string) (*responses.Device, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	device, err := s.deviceRepository.GetById(id)
	if err != nil {
		return nil, err
	}

	if !user.IsInternal {
		brandID, parseErr := primitive.ObjectIDFromHex(device.Brand)
		if parseErr != nil || !userHasBrandAccess(user, brandID) {
			return nil, fmt.Errorf("%w", utils.ErrorForbidden)
		}
	}

	return device, nil
}

// Update updates a test case
func (s *deviceService) Update(id string, deviceRequest *request.Device, userID string) error {
	if err := s.ensureDeviceModelsUnique(id, deviceRequest.TechnicalModel, deviceRequest.CommercialModel); err != nil {
		return err
	}

	device := mapping.DeviceRequestToDevice(deviceRequest)

	err := s.deviceRepository.Update(id, device)
	if err != nil {
		return err
	}
	go s.DeviceNotification(*deviceRequest, utils.EDIT, userID)
	return nil
}
func (s *deviceService) Delete(id string) (bool, error) {
	deviceId, _ := primitive.ObjectIDFromHex(id)

	homologation, err := s.homologationRepository.GetByDevice(deviceId)
	if err != nil {
		return false, err
	}
	if homologation != nil {
		return true, err
	}
	deviceTacking, err := s.deviceTrackingRepository.GetByDevice(deviceId)
	if err != nil {
		return false, err
	}
	if deviceTacking != nil {
		return true, err
	}
	err = s.deviceRepository.Delete(deviceId)

	if err != nil {
		return false, err
	}
	return false, nil
}
func (s *deviceService) DeviceNotification(device request.Device, key string,userID string) {

	toList, isEmpty := functions.GetEmails(true, primitive.NewObjectID())
	if isEmpty {
		return
	}
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return
	}
	userName := fmt.Sprintf("%s %s", user.Name, user.LastName)
	brand, err := s.brandRepository.GetById(device.Brand)
	if err != nil {
		return
	}
	var subject string
	var mainMessage string

	switch key {

	case utils.CREATE:
		subject = fmt.Sprintf("Subject: New Device created for %s %s",
			brand.Name, device.CommercialModel)
		mainMessage = utils.CREATE_DEVICE_MAIN_MESSAGE
	case utils.EDIT:
		subject = fmt.Sprintf("Subject: Device updated %s %s",
			brand.Name, device.CommercialModel)
		mainMessage = utils.UPDATE_DEVICE_MAIN_MESSAGE
	default:
		return
	}

	body, err := functions.GetDeviceBodyMessage(subject, mainMessage, brand.Name,
		device, utils.TEMPLATE_DEVICE_PATH,userName)

	if err != nil {
		return
	}
	functions.SendNotifications(toList, body)
}

func (s *deviceService) ensureDeviceModelsUnique(excludeID, technicalModel, commercialModel string) error {
	technicalModel = strings.TrimSpace(technicalModel)
	commercialModel = strings.TrimSpace(commercialModel)
	if technicalModel == "" {
		return utils.NewValidationError("technical_model is required")
	}
	if commercialModel == "" {
		return utils.NewValidationError("commercial_model is required")
	}

	if excludeID == "" {
		byTechnical, err := s.deviceRepository.FindByTechnicalModel(technicalModel)
		if err != nil {
			return err
		}
		if len(byTechnical) > 0 {
			return utils.NewValidationError("a device with this technical_model already exists")
		}

		byCommercial, err := s.deviceRepository.FindByCommercialModel(commercialModel)
		if err != nil {
			return err
		}
		if len(byCommercial) > 0 {
			return utils.NewValidationError("a device with this commercial_model already exists")
		}
		return nil
	}

	oid, err := primitive.ObjectIDFromHex(excludeID)
	if err != nil {
		return utils.NewValidationError("invalid device id")
	}

	existsTechnical, err := s.deviceRepository.ExistsByTechnicalModelExcludingID(oid, technicalModel)
	if err != nil {
		return err
	}
	if existsTechnical {
		return utils.NewValidationError("a device with this technical_model already exists")
	}

	existsCommercial, err := s.deviceRepository.ExistsByCommercialModelExcludingID(oid, commercialModel)
	if err != nil {
		return err
	}
	if existsCommercial {
		return utils.NewValidationError("a device with this commercial_model already exists")
	}
	return nil
}

func userHasBrandAccess(user *responses.User, brandID primitive.ObjectID) bool {
	if user.IsInternal && len(user.Brands) == 0 {
		return true
	}
	for _, allowedBrand := range user.Brands {
		if allowedBrand == brandID {
			return true
		}
	}
	return false
}
