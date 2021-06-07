package services

import (
	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IDeviceService is the test case service interface
type IDeviceService interface {
	Create(*request.Device) (*responses.DeviceExpanded, error)
	Get() ([]*responses.DeviceExpanded, error)
	GetById(string) (*responses.Device, error)
	Update(string, *request.Device) error
	Delete(string) error
}

type deviceService struct {
	deviceRepository repositories.IDeviceRepository
}

// NewDeviceService is a constructor
func NewDeviceService(deviceRepository repositories.IDeviceRepository) IDeviceService {
	return &deviceService{
		deviceRepository: deviceRepository,
	}
}

// Create creates a new test case
func (s *deviceService) Create(deviceRequest *request.Device) (*responses.DeviceExpanded, error) {

	device := mapping.DeviceRequestToDevice(deviceRequest)

	deviceResponse, err := s.deviceRepository.Create(device)

	if err != nil {
		return nil, err
	}
	return deviceResponse, nil
}

// Get gets a list of test cases
func (s *deviceService) Get() ([]*responses.DeviceExpanded, error) {
	result, err := s.deviceRepository.Get()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetById gets a case by id
func (s *deviceService) GetById(id string) (*responses.Device, error) {

	device, err := s.deviceRepository.GetById(id)

	if err != nil {
		return nil, err
	}
	return device, nil
}

// Update updates a test case
func (s *deviceService) Update(id string, deviceRequest *request.Device) error {

	device := mapping.DeviceRequestToDevice(deviceRequest)

	err := s.deviceRepository.Update(id, device)
	if err != nil {
		return err
	}
	return nil
}
func (s *deviceService) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	err = s.deviceRepository.Delete(oid)

	if err != nil {
		return err
	}
	return nil
}
