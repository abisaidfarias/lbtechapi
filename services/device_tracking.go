package services

import (
	"fmt"
	"strings"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IDeviceTrackingService is the deviceTracking service
type IDeviceTrackingService interface {
	Create(*request.DeviceTracking) error
	Get(string) ([]responses.Tracking, error)
	AddTrakingLog(*request.TrackingLogMultiple) error
	Delete(string) error
	Update(string, *request.DeviceTrackingExpanded) error
	AdvancedSearch(*request.SearchOption, string) ([]responses.Tracking, error)
	AdvancedSearchOptions(userId string) (responses.SearchOption, error)
}

type deviceTrackingService struct {
	deviceTrackingRepository repositories.IDeviceTrackingRepository
	userRepository           repositories.IUserRepository
}

// NewDeviceTrackingService is a constructor
func NewDeviceTrackingService(deviceTrackingRepository repositories.IDeviceTrackingRepository,
	userRepository repositories.IUserRepository) IDeviceTrackingService {
	return &deviceTrackingService{
		deviceTrackingRepository: deviceTrackingRepository,
		userRepository:           userRepository,
	}
}

// Create creates a new cateogry
func (s *deviceTrackingService) Create(deviceTrackingRequest *request.DeviceTracking) error {

	for _, imei := range deviceTrackingRequest.Imeis {
		deviceTracking := mapping.DeviceTrackinRequestToDeviceTracking(deviceTrackingRequest, imei)
		err := s.deviceTrackingRepository.Create(deviceTracking)

		if err != nil {
			return err
		}

	}
	return nil
}

// Get gets a list of all categories
func (s *deviceTrackingService) Get(userID string) ([]responses.Tracking, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	deviceTrackings, err := s.deviceTrackingRepository.Get(user.IsInternal, user.Company, user.Brands)
	if err != nil {
		return nil, err
	}

	trackingGrouped := make(map[string]responses.Tracking)
	for _, deviceTracking := range deviceTrackings {
		deviceName := fmt.Sprintf("%s %s",
			deviceTracking.Device.Brand.Name, deviceTracking.Device.CommercialModel)
		existTracking := trackingGrouped[deviceName]
		existTracking.Brand = deviceTracking.Device.Brand.Name
		existTracking.Model = deviceTracking.Device.CommercialModel
		existTracking.ID = deviceTracking.Device.ID
		existTracking.ImageUrl = deviceTracking.Device.ImageUrl
		existTracking.TecnicalModel = deviceTracking.Device.TechnicalModel
		existTracking.DeviceTrackings = append(existTracking.DeviceTrackings, *deviceTracking)
		trackingGrouped[deviceName] = existTracking

	}
	var trakings []responses.Tracking = []responses.Tracking{}
	for _, v := range trackingGrouped {
		trakings = append(trakings, v)
	}

	return trakings, nil
}
func (s *deviceTrackingService) AddTrakingLog(trackingLogReq *request.TrackingLogMultiple) error {

	for _, id := range trackingLogReq.DeviceTrackings {
		deviceTranckingID, _ := primitive.ObjectIDFromHex(id)
		trackingLog := mapping.TrackinLogRequestToTrackingLog(&trackingLogReq.TrackingLog)
		err := s.deviceTrackingRepository.AddTrakingLog(trackingLog, deviceTranckingID)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *deviceTrackingService) Delete(ids string) error {

	deviceTrackingSplits := strings.Split(ids, ",")

	var deviceTrackingIds []primitive.ObjectID = []primitive.ObjectID{}
	for _, id := range deviceTrackingSplits {
		deviceId, _ := primitive.ObjectIDFromHex(id)
		deviceTrackingIds = append(deviceTrackingIds, deviceId)
	}
	err := s.deviceTrackingRepository.Delete(deviceTrackingIds)
	if err != nil {
		return err
	}
	return nil
}
func (s *deviceTrackingService) Update(id string, deviceTrackingRequest *request.DeviceTrackingExpanded) error {

	deviceTracking := mapping.DeviceTrackinRequestToDeviceTrackingUpdate(deviceTrackingRequest)

	err := s.deviceTrackingRepository.Update(id, deviceTracking)
	if err != nil {
		return err
	}
	return nil
}
func (s *deviceTrackingService) AdvancedSearch(searchOption *request.SearchOption, userId string) ([]responses.Tracking, error) {
	user, err := s.userRepository.GetByID(userId)
	if err != nil {
		return nil, err
	}
	deviceTrackings, err := s.deviceTrackingRepository.AdvancedSearch(searchOption, user.Company, user.IsInternal)
	if err != nil {
		return nil, err
	}

	trackingGrouped := make(map[string]responses.Tracking)
	for _, deviceTracking := range deviceTrackings {
		existTracking := trackingGrouped["NoDevice"]
		existTracking.Brand = ""
		existTracking.Model = ""
		existTracking.ID = primitive.NilObjectID
		existTracking.ImageUrl = ""
		existTracking.TecnicalModel = ""
		existTracking.DeviceTrackings = append(existTracking.DeviceTrackings, *deviceTracking)
		trackingGrouped["NoDevice"] = existTracking

	}
	var trakings []responses.Tracking = []responses.Tracking{}
	for _, v := range trackingGrouped {
		trakings = append(trakings, v)
	}

	return trakings, nil
}
func (s *deviceTrackingService) AdvancedSearchOptions(userId string) (responses.SearchOption, error) {

	user, err := s.userRepository.GetByID(userId)
	var searchOption responses.SearchOption = responses.SearchOption{}
	if err != nil {
		return searchOption, err
	}
	deviceTrackings, err := s.deviceTrackingRepository.Get(user.IsInternal, user.Company, user.Brands)
	if err != nil {
		return searchOption, err
	}
	brandsUnique := make(map[string]bool)
	modelsUnique := make(map[string]bool)
	countryUnique := make(map[string]bool)
	locationUnique := make(map[string]bool)
	for _, deviceTracking := range deviceTrackings {

		if _, value := brandsUnique[deviceTracking.Device.Brand.Name]; !value {
			brandsUnique[deviceTracking.Device.Brand.Name] = true
			searchOption.Brands = append(searchOption.Brands, deviceTracking.Device.Brand.Name)
		}
		if _, value := modelsUnique[deviceTracking.Device.CommercialModel]; !value {
			modelsUnique[deviceTracking.Device.CommercialModel] = true
			searchOption.CommercialModels = append(searchOption.CommercialModels, deviceTracking.Device.CommercialModel)
		}
		for _, trackingLog := range deviceTracking.TrackingLogs {
			if _, value := locationUnique[trackingLog.Location.Name]; !value {
				locationUnique[trackingLog.Location.Name] = true
				searchOption.Locations = append(searchOption.Locations, trackingLog.Location.Name)
			}
			if _, value := countryUnique[trackingLog.Country.Name]; !value {
				countryUnique[trackingLog.Country.Name] = true
				searchOption.Countries = append(searchOption.Countries, trackingLog.Country.Name)
			}

		}
	}
	return searchOption, nil
}
