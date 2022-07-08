package services

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/repositories"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"

	//"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IDeviceTrackingService is the deviceTracking service
type IDeviceTrackingService interface {
	Create(*request.DeviceTracking, string) error
	Get(string) ([]responses.Tracking, error)
	AddTrakingLog(*request.TrackingLogMultiple, string) error
	Delete(string) error
	Update(string, *request.DeviceTrackingExpanded) error
	AdvancedSearch(*request.SearchOption, string) ([]responses.Tracking, error)
	AdvancedSearchOptions(userId string) (responses.SearchOption, error)
	ExportDeviceTracking(*request.SearchOption, string) (bytes.Buffer, error)
}

type deviceTrackingService struct {
	deviceTrackingRepository repositories.IDeviceTrackingRepository
	userRepository           repositories.IUserRepository
	companyRepository        repositories.ICompanyRepository
	brandRepository          repositories.IBrandRepository
	deviceRepository         repositories.IDeviceRepository
	countryRepository        repositories.ICountryRepository
}

// NewDeviceTrackingService is a constructor
func NewDeviceTrackingService(deviceTrackingRepository repositories.IDeviceTrackingRepository,
	userRepository repositories.IUserRepository,
	companyRepository repositories.ICompanyRepository,
	brandRepository repositories.IBrandRepository,
	deviceRepository repositories.IDeviceRepository,
	countryRepository repositories.ICountryRepository) IDeviceTrackingService {
	return &deviceTrackingService{
		deviceTrackingRepository: deviceTrackingRepository,
		userRepository:           userRepository,
		companyRepository:        companyRepository,
		brandRepository:          brandRepository,
		deviceRepository:         deviceRepository,
		countryRepository:        countryRepository,
	}
}

// Create creates a new cateogry
func (s *deviceTrackingService) Create(deviceTrackingRequest *request.DeviceTracking, userID string) error {

	for _, imei := range deviceTrackingRequest.Imeis {
		deviceTracking := mapping.DeviceTrackinRequestToDeviceTracking(deviceTrackingRequest, imei)
		err := s.deviceTrackingRepository.Create(deviceTracking)

		if err != nil {
			return err
		}
	}
	companyId, _ := primitive.ObjectIDFromHex(deviceTrackingRequest.Company)

	name := fmt.Sprintf("%s %s", deviceTrackingRequest.TrackingLog.InternalResponsible.Name,
		deviceTrackingRequest.TrackingLog.InternalResponsible.LastName)
	go s.DeviceTrackingNotification(deviceTrackingRequest.Imeis,
		deviceTrackingRequest.TrackingLog.Country.Name,
		deviceTrackingRequest.Device, name,
		deviceTrackingRequest.TrackingLog.Person.Name,
		deviceTrackingRequest.TrackingLog.Comment,
		deviceTrackingRequest.TrackingLog.Location.Name, companyId,
		deviceTrackingRequest.TrackingLog.TrackingDate, utils.CREATE, userID)

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
func (s *deviceTrackingService) AddTrakingLog(trackingLogReq *request.TrackingLogMultiple, userID string) error {
	var devicesTrackingsId []string
	for _, id := range trackingLogReq.DeviceTrackings {
		deviceTranckingID, _ := primitive.ObjectIDFromHex(id)
		devicesTrackingsId = append(devicesTrackingsId, id)
		trackingLog := mapping.TrackinLogRequestToTrackingLog(&trackingLogReq.TrackingLog)
		err := s.deviceTrackingRepository.AddTrakingLog(trackingLog, deviceTranckingID)
		if err != nil {
			return err
		}
	}
	go s.MoveTrackingNotification(devicesTrackingsId, trackingLogReq.TrackingLog, userID)
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
func (s *deviceTrackingService) ExportDeviceTracking(searchOption *request.SearchOption, userId string) (bytes.Buffer, error) {
	user, err := s.userRepository.GetByID(userId)
	var b bytes.Buffer
	if err != nil {
		return b, err
	}
	deviceTrackings, err := s.deviceTrackingRepository.AdvancedSearch(searchOption, user.ID, user.IsInternal)
	if err != nil {
		return b, err
	}
	file, err := exportFileDeviceTracking(deviceTrackings)
	if err != nil {
		return file, err
	}
	return file, nil
}
func exportFileDeviceTracking(deviceTrackings []*responses.DeviceTrackingExpanded) (bytes.Buffer, error) {
	file := excelize.NewFile()
	categories := enums.ExcelDeviceTrackinHeaders
	for k, v := range categories {
		file.SetCellValue(utils.PAGE, k, v)
	}
	for index, d := range deviceTrackings {
		cell, _ := excelize.CoordinatesToCellName(1, index+2)
		file.SetCellValue(utils.PAGE, cell, d.Imei)
		cell, _ = excelize.CoordinatesToCellName(2, index+2)
		file.SetCellValue(utils.PAGE, cell, d.Device.Brand.Name)
		cell, _ = excelize.CoordinatesToCellName(3, index+2)
		file.SetCellValue(utils.PAGE, cell, d.Device.CommercialModel)
		cell, _ = excelize.CoordinatesToCellName(4, index+2)
		file.SetCellValue(utils.PAGE, cell, d.Company.Name)

		lastTrackingRegister := d.TrackingLogs[len(d.TrackingLogs)-1]
		cell, _ = excelize.CoordinatesToCellName(5, index+2)
		file.SetCellValue(utils.PAGE, cell, lastTrackingRegister.Country.Name)
		cell, _ = excelize.CoordinatesToCellName(6, index+2)
		file.SetCellValue(utils.PAGE, cell, lastTrackingRegister.Location.Name)
		cell, _ = excelize.CoordinatesToCellName(7, index+2)
		file.SetCellValue(utils.PAGE, cell, lastTrackingRegister.InternalResponsible.Name)
		cell, _ = excelize.CoordinatesToCellName(8, index+2)
		file.SetCellValue(utils.PAGE, cell, lastTrackingRegister.Person.Name)

		year, month, day := lastTrackingRegister.TrackingDate.Date()
		cell, _ = excelize.CoordinatesToCellName(9, index+2)
		file.SetCellValue(utils.PAGE, cell, fmt.Sprintf("%d/%d/%d", day, month, year))
	}

	var b bytes.Buffer
	if err := file.Write(&b); err != nil {
		return b, err
	}
	return b, nil
}
func (s *deviceTrackingService) DeviceTrackingNotification(imeis []string,
	country string, deviceId string, internal string,
	external string, comment string, location string,
	companyId primitive.ObjectID, date time.Time, key string, userID string) {

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return
	}
	userName := fmt.Sprintf("%s %s", user.Name, user.LastName)
	toList, isEmpty := functions.GetEmails(false, companyId)
	if isEmpty {
		return
	}
	device, err := s.deviceRepository.GetById(deviceId)
	if err != nil {
		return
	}
	brand, err := s.brandRepository.GetById(device.Brand)
	if err != nil {
		return
	}
	company, err := s.companyRepository.GetById(companyId.Hex())
	if err != nil {
		return
	}

	var subject string
	var mainMessage string

	switch key {

	case utils.CREATE:
		subject = fmt.Sprintf("Subject: New Sample(s) received at LB Technology for %s %s %s",
			country, brand.Name, device.CommercialModel)
		mainMessage = utils.CREATE_TRACKING_MAIN_MESSAGE
	case utils.TRACKING_MOVE:
		subject = fmt.Sprintf("Subject: Sample(s) has been moved of location to %s %s %s",
			country, brand.Name, device.CommercialModel)
		mainMessage = utils.MOVE_TRACKING_MAIN_MESSAGE
	default:
		return
	}
	body, err := functions.GetTrackingBodyMessage(subject, mainMessage, company.Name, brand.Name,
		device.TechnicalModel, device.CommercialModel, internal,
		external, country, location, strings.Join(imeis[:], ","),
		comment, date, utils.TEMPLATE_TRACKING_PATH, userName)

	if err != nil {
		return
	}
	functions.SendNotifications(toList, body)
}
func (s *deviceTrackingService) MoveTrackingNotification(deviceTrackingsId []string,
	trackingLog request.TrackingLog, userID string) {

	companyGrouped := make(map[primitive.ObjectID][]string)
	var deviceId string
	for _, id := range deviceTrackingsId {
		deviceTracking, _ := s.deviceTrackingRepository.GetById(id)
		value := companyGrouped[deviceTracking.Company]
		value = append(value, deviceTracking.Imei)
		companyGrouped[deviceTracking.Company] = value
		deviceId = deviceTracking.Device.Hex()
	}
	for companyId, imeis := range companyGrouped {
		name := fmt.Sprintf("%s %s", trackingLog.InternalResponsible.Name,
			trackingLog.InternalResponsible.LastName)
		s.DeviceTrackingNotification(imeis, trackingLog.Country.Name,
			deviceId, name, trackingLog.Person.Name,
			trackingLog.Comment, trackingLog.Location.Name,
			companyId, trackingLog.TrackingDate, utils.TRACKING_MOVE, userID)
	}
}
