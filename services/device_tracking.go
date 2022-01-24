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
	deviceTrackings, err := s.deviceTrackingRepository.Get(user.IsInternal, user.Company)
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
