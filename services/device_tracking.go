package services

import (
	"bytes"
	"fmt"
	"sort"
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

	if err := enums.ValidateProcessTypes(deviceTrackingRequest.TrackingLog.ProcessTypes); err != nil {
		return err
	}
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
		deviceTrackingRequest.TrackingLog.TrackingDate, deviceTrackingRequest.TrackingLog.ProcessTypes,
		utils.CREATE, userID)

	return nil
}

// Get gets a list of all categories
func (s *deviceTrackingService) Get(userID string) ([]responses.Tracking, error) {
	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	var countries []string
	if len(user.Countries) > 0 {
		countries, _ = s.countryRepository.GetCountriesById(user.Countries)
	}
	deviceTrackings, err := s.deviceTrackingRepository.Get(user.IsInternal, user.Company,
		user.Brands, countries, user.Clients)
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
	sort.Slice(trakings, func(i, j int) bool {
		return trakings[i].Brand < trakings[j].Brand
	})
	return trakings, nil
}
func (s *deviceTrackingService) AddTrakingLog(trackingLogReq *request.TrackingLogMultiple, userID string) error {
	if err := enums.ValidateProcessTypes(trackingLogReq.TrackingLog.ProcessTypes); err != nil {
		return err
	}
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

	for _, tl := range deviceTrackingRequest.TrackingLogs {
		if err := enums.ValidateProcessTypes(tl.ProcessTypes); err != nil {
			return err
		}
	}
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
	var countries []string
	if len(user.Countries) > 0 {
		countries, _ = s.countryRepository.GetCountriesById(user.Countries)
	}
	deviceTrackings, err := s.deviceTrackingRepository.AdvancedSearch(searchOption,
		user.Company, user.IsInternal, user.Brands, countries, user.Clients)
	if err != nil {
		return nil, err
	}

	trackingGrouped := make(map[string]responses.Tracking)
	for _, deviceTracking := range deviceTrackings {
		if len(deviceTracking.TrackingLogs) == 0 {
			continue
		}

		lastRecord := deviceTracking.TrackingLogs[len(deviceTracking.TrackingLogs)-1]
		if !isContained(lastRecord, searchOption.Locations, searchOption.Countries) {
			continue
		}

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
	if len(trakings) == 0 {
		return []responses.Tracking{}, nil
	}
	for i := range trakings {
		sort.Slice(trakings[i].DeviceTrackings, func(j, k int) bool {
			return trakings[i].DeviceTrackings[j].Device.Brand.Name < trakings[i].DeviceTrackings[k].Device.Brand.Name
		})
	}
	return trakings, nil
}
func (s *deviceTrackingService) AdvancedSearchOptions(userId string) (responses.SearchOption, error) {

	user, err := s.userRepository.GetByID(userId)
	searchOption := responses.SearchOption{
		Brands:           []string{},
		CommercialModels: []string{},
		Countries:        []string{},
		Locations:        []string{},
		Relations: responses.SearchOptionRelations{
			ByBrand:    map[string]responses.SearchOptionRelationByBrand{},
			ByModel:    map[string]responses.SearchOptionRelationByModel{},
			ByCountry:  map[string]responses.SearchOptionRelationByCountry{},
			ByLocation: map[string]responses.SearchOptionRelationByLocation{},
		},
	}
	if err != nil {
		return searchOption, err
	}
	var countries []string
	if !user.IsInternal {
		countries, _ = s.countryRepository.GetCountriesById(user.Countries)
	}
	deviceTrackings, err := s.deviceTrackingRepository.Get(user.IsInternal, user.Company,
		user.Brands, countries, user.Clients)
	if err != nil {
		return searchOption, err
	}
	brandsUnique := make(map[string]struct{})
	modelsUnique := make(map[string]struct{})
	countryUnique := make(map[string]struct{})
	locationUnique := make(map[string]struct{})

	byBrandModels := make(map[string]map[string]struct{})
	byBrandCountries := make(map[string]map[string]struct{})
	byBrandLocations := make(map[string]map[string]struct{})

	byModelBrands := make(map[string]map[string]struct{})
	byModelCountries := make(map[string]map[string]struct{})
	byModelLocations := make(map[string]map[string]struct{})

	byCountryBrands := make(map[string]map[string]struct{})
	byCountryModels := make(map[string]map[string]struct{})
	byCountryLocations := make(map[string]map[string]struct{})

	byLocationBrands := make(map[string]map[string]struct{})
	byLocationModels := make(map[string]map[string]struct{})
	byLocationCountries := make(map[string]map[string]struct{})

	for _, deviceTracking := range deviceTrackings {
		brandName := strings.TrimSpace(deviceTracking.Device.Brand.Name)
		modelName := strings.TrimSpace(deviceTracking.Device.CommercialModel)

		if brandName != "" {
			brandsUnique[brandName] = struct{}{}
		}
		if modelName != "" {
			modelsUnique[modelName] = struct{}{}
		}
		if brandName != "" && modelName != "" {
			addRelation(byBrandModels, brandName, modelName)
			addRelation(byModelBrands, modelName, brandName)
		}

		for _, trackingLog := range deviceTracking.TrackingLogs {
			countryName := strings.TrimSpace(trackingLog.Country.Name)
			locationName := strings.TrimSpace(trackingLog.Location.Name)

			if countryName != "" {
				countryUnique[countryName] = struct{}{}
			}
			if locationName != "" {
				locationUnique[locationName] = struct{}{}
			}

			if brandName != "" {
				if countryName != "" {
					addRelation(byBrandCountries, brandName, countryName)
					addRelation(byCountryBrands, countryName, brandName)
				}
				if locationName != "" {
					addRelation(byBrandLocations, brandName, locationName)
					addRelation(byLocationBrands, locationName, brandName)
				}
			}
			if modelName != "" {
				if countryName != "" {
					addRelation(byModelCountries, modelName, countryName)
					addRelation(byCountryModels, countryName, modelName)
				}
				if locationName != "" {
					addRelation(byModelLocations, modelName, locationName)
					addRelation(byLocationModels, locationName, modelName)
				}
			}
			if countryName != "" && locationName != "" {
				addRelation(byCountryLocations, countryName, locationName)
				addRelation(byLocationCountries, locationName, countryName)
			}
		}
	}

	searchOption.Brands = sortedKeys(brandsUnique)
	searchOption.CommercialModels = sortedKeys(modelsUnique)
	searchOption.Countries = sortedKeys(countryUnique)
	searchOption.Locations = sortedKeys(locationUnique)

	for _, brandName := range searchOption.Brands {
		searchOption.Relations.ByBrand[brandName] = responses.SearchOptionRelationByBrand{
			CommercialModels: sortedValues(byBrandModels, brandName),
			Countries:        sortedValues(byBrandCountries, brandName),
			Locations:        sortedValues(byBrandLocations, brandName),
		}
	}
	for _, modelName := range searchOption.CommercialModels {
		searchOption.Relations.ByModel[modelName] = responses.SearchOptionRelationByModel{
			Brands:    sortedValues(byModelBrands, modelName),
			Countries: sortedValues(byModelCountries, modelName),
			Locations: sortedValues(byModelLocations, modelName),
		}
	}
	for _, countryName := range searchOption.Countries {
		searchOption.Relations.ByCountry[countryName] = responses.SearchOptionRelationByCountry{
			Brands:           sortedValues(byCountryBrands, countryName),
			CommercialModels: sortedValues(byCountryModels, countryName),
			Locations:        sortedValues(byCountryLocations, countryName),
		}
	}
	for _, locationName := range searchOption.Locations {
		searchOption.Relations.ByLocation[locationName] = responses.SearchOptionRelationByLocation{
			Brands:           sortedValues(byLocationBrands, locationName),
			CommercialModels: sortedValues(byLocationModels, locationName),
			Countries:        sortedValues(byLocationCountries, locationName),
		}
	}

	return searchOption, nil
}

func addRelation(container map[string]map[string]struct{}, source string, target string) {
	if source == "" || target == "" {
		return
	}
	if _, exists := container[source]; !exists {
		container[source] = map[string]struct{}{}
	}
	container[source][target] = struct{}{}
}

func sortedKeys(values map[string]struct{}) []string {
	items := make([]string, 0, len(values))
	for key := range values {
		if key == "" {
			continue
		}
		items = append(items, key)
	}
	sort.Strings(items)
	return items
}

func sortedValues(container map[string]map[string]struct{}, key string) []string {
	values, exists := container[key]
	if !exists {
		return []string{}
	}
	return sortedKeys(values)
}
func (s *deviceTrackingService) ExportDeviceTracking(searchOption *request.SearchOption, userId string) (bytes.Buffer, error) {
	user, err := s.userRepository.GetByID(userId)
	var b bytes.Buffer
	if err != nil {
		return b, err
	}
	var countries []string
	if len(user.Countries) > 0 {
		countries, _ = s.countryRepository.GetCountriesById(user.Countries)
	}
	deviceTrackings, err := s.deviceTrackingRepository.AdvancedSearch(searchOption, user.ID,
		user.IsInternal, user.Brands, countries, user.Clients)
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
		cell, _ = excelize.CoordinatesToCellName(10, index+2)
		file.SetCellValue(utils.PAGE, cell, strings.Join(lastTrackingRegister.ProcessTypes, ", "))
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
	companyId primitive.ObjectID, date time.Time, processTypes []string, key string, userID string) {

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
		external, country, location, imeis,
		comment, date, processTypes, utils.TEMPLATE_TRACKING_PATH, userName)

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
			companyId, trackingLog.TrackingDate, trackingLog.ProcessTypes,
			utils.TRACKING_MOVE, userID)
	}
}
func isContained(trackingLog responses.TrackingLog, locations []string, countries []string) bool {

	isContained := true
	if len(countries) > 0 {
		isContained = false
		for _, country := range countries {
			if trackingLog.Country.Name == country {
				isContained = true
				break
			}
		}
	}
	if len(locations) > 0 {
		isContained = false
		for _, location := range locations {
			if trackingLog.Location.Name == location {
				isContained = true
				break
			}
		}
	}
	return isContained
}
