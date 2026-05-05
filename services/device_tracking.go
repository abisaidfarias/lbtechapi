package services

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
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

// ErrDeliveryConfirmNotFound means no device_tracking document matched imei + tracking_id (or log missing).
var ErrDeliveryConfirmNotFound = errors.New("device tracking not found for imei and tracking_id")

var (
	santiagoTZOnce sync.Once
	santiagoTZ     *time.Location
)

// santiagoLocation returns IANA America/Santiago for email/PDF display (Chile time).
func santiagoLocation() *time.Location {
	santiagoTZOnce.Do(func() {
		loc, err := time.LoadLocation("America/Santiago")
		if err != nil {
			log.Printf("timezone America/Santiago unavailable (%v); using UTC-3 fallback for tracking email dates", err)
			santiagoTZ = time.FixedZone("CLT", -3*60*60)
			return
		}
		santiagoTZ = loc
	})
	return santiagoTZ
}

// ErrDeliveryConfirmInconsistent means the tracking log for the same tracking_id differs between devices.
var ErrDeliveryConfirmInconsistent = errors.New("tracking log mismatch across devices for the same tracking_id")

// ErrDeliveryConfirmInvalidSignature means signature_png_data_url is not an allowed data:image PNG/JPEG base64 URL.
var ErrDeliveryConfirmInvalidSignature = errors.New("signature_png_data_url must start with data:image/png;base64, or data:image/jpeg;base64,")

// IDeviceTrackingService is the deviceTracking service
type IDeviceTrackingService interface {
	Create(*request.DeviceTracking, string) error
	Get(string) ([]responses.Tracking, error)
	AddTrakingLog(*request.TrackingLogMultiple, string) (string, error)
	ConfirmMoveDeliveryReport(*request.DeliveryConfirmMoveReportRequest, string) error
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
	storageService           IStorageService
}

// NewDeviceTrackingService is a constructor
func NewDeviceTrackingService(deviceTrackingRepository repositories.IDeviceTrackingRepository,
	userRepository repositories.IUserRepository,
	companyRepository repositories.ICompanyRepository,
	brandRepository repositories.IBrandRepository,
	deviceRepository repositories.IDeviceRepository,
	countryRepository repositories.ICountryRepository,
	storageService IStorageService) IDeviceTrackingService {
	return &deviceTrackingService{
		deviceTrackingRepository: deviceTrackingRepository,
		userRepository:           userRepository,
		companyRepository:        companyRepository,
		brandRepository:          brandRepository,
		deviceRepository:         deviceRepository,
		countryRepository:        countryRepository,
		storageService:           storageService,
	}
}

// Create creates a new cateogry
func (s *deviceTrackingService) Create(deviceTrackingRequest *request.DeviceTracking, userID string) error {

	if err := enums.ValidateProcessTypes(deviceTrackingRequest.TrackingLog.ProcessTypes); err != nil {
		return err
	}

	trackingID, err := s.deviceTrackingRepository.ReserveNextSequentialMoveTrackingID()
	if err != nil {
		return err
	}
	deviceTrackingRequest.TrackingLog.TrackingID = trackingID

	var docIDs []string
	for _, imei := range deviceTrackingRequest.Imeis {
		deviceTracking := mapping.DeviceTrackinRequestToDeviceTracking(deviceTrackingRequest, imei)
		err := s.deviceTrackingRepository.Create(deviceTracking)
		if err != nil {
			return err
		}
		if !deviceTracking.ID.IsZero() {
			docIDs = append(docIDs, deviceTracking.ID.Hex())
		}
	}
	companyId, _ := primitive.ObjectIDFromHex(deviceTrackingRequest.Company)

	rows, mailErr := s.buildTrackingEmailRowsSingleDevice(deviceTrackingRequest.Company, deviceTrackingRequest.Device,
		deviceTrackingRequest.Imeis, &deviceTrackingRequest.TrackingLog, userID)
	if mailErr != nil || len(rows) == 0 {
		return nil
	}

	go s.ensureMoveReportArtifacts(docIDs, trackingID, rows, nil, true, false, companyId, emailKindCreate)

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
func (s *deviceTrackingService) AddTrakingLog(trackingLogReq *request.TrackingLogMultiple, userID string) (string, error) {
	if err := enums.ValidateProcessTypes(trackingLogReq.TrackingLog.ProcessTypes); err != nil {
		return "", err
	}
	trackingID, err := s.deviceTrackingRepository.ReserveNextSequentialMoveTrackingID()
	if err != nil {
		return "", err
	}
	trackingLogReq.TrackingLog.TrackingID = trackingID

	var devicesTrackingsId []string
	for _, id := range trackingLogReq.DeviceTrackings {
		deviceTranckingID, _ := primitive.ObjectIDFromHex(id)
		devicesTrackingsId = append(devicesTrackingsId, id)
		trackingLog := mapping.TrackinLogRequestToTrackingLog(&trackingLogReq.TrackingLog)
		err = s.deviceTrackingRepository.AddTrakingLog(trackingLog, deviceTranckingID)
		if err != nil {
			return "", err
		}
	}
	go s.MoveTrackingNotification(devicesTrackingsId, trackingLogReq.TrackingLog, userID, trackingLogReq.WithDelivery, nil)
	return trackingID, nil
}

func findTrackingLogInDevice(dt *responses.DeviceTracking, trackingID string) *responses.TrackingLog {
	for i := range dt.TrackingLogs {
		if dt.TrackingLogs[i].TrackingID == trackingID {
			return &dt.TrackingLogs[i]
		}
	}
	return nil
}

func equalStringSlicesSorted(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aa := append([]string{}, a...)
	bb := append([]string{}, b...)
	sort.Strings(aa)
	sort.Strings(bb)
	for i := range aa {
		if aa[i] != bb[i] {
			return false
		}
	}
	return true
}

func moveDeliveryTrackingLogsEqual(a, b responses.TrackingLog) bool {
	if a.TrackingID != b.TrackingID {
		return false
	}
	if a.Country.Name != b.Country.Name || a.Location.Name != b.Location.Name {
		return false
	}
	if a.Person.Name != b.Person.Name {
		return false
	}
	if a.InternalResponsible.Email != b.InternalResponsible.Email ||
		a.InternalResponsible.Name != b.InternalResponsible.Name ||
		a.InternalResponsible.LastName != b.InternalResponsible.LastName {
		return false
	}
	if strings.TrimSpace(a.Comment) != strings.TrimSpace(b.Comment) {
		return false
	}
	if !equalStringSlicesSorted(a.ProcessTypes, b.ProcessTypes) {
		return false
	}
	if a.ExternalDelivery != b.ExternalDelivery {
		return false
	}
	ta := a.TrackingDate.UTC().Truncate(time.Minute)
	tb := b.TrackingDate.UTC().Truncate(time.Minute)
	if !ta.Equal(tb) {
		return false
	}
	return true
}

func responseTrackingLogToRequest(tl responses.TrackingLog) request.TrackingLog {
	uid := ""
	if !tl.InternalResponsible.ID.IsZero() {
		uid = tl.InternalResponsible.ID.Hex()
	}
	return request.TrackingLog{
		TrackingID: tl.TrackingID,
		Country: request.Country{
			Name:       tl.Country.Name,
			BandGsm:    tl.Country.BandGsm,
			BandWcdma:  tl.Country.BandWcdma,
			BandLte:    tl.Country.BandLte,
			Band5g:     tl.Country.Band5g,
			CarrierAgg: tl.Country.CarrierAgg,
		},
		Location: request.Location{Name: tl.Location.Name},
		InternalResponsible: request.UserResume{
			Email:      tl.InternalResponsible.Email,
			Name:       tl.InternalResponsible.Name,
			LastName:   tl.InternalResponsible.LastName,
			IsInternal: tl.InternalResponsible.IsInternal,
			UserID:     uid,
		},
		Person:           request.Person{Name: tl.Person.Name},
		Comment:          tl.Comment,
		DocumentUrl:      tl.DocumentUrl,
		TrackingDate:     tl.TrackingDate,
		ExternalDelivery: tl.ExternalDelivery,
		ProcessTypes:     append([]string{}, tl.ProcessTypes...),
	}
}

func isAllowedSignatureDataURL(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "data:image/png;base64,") || strings.HasPrefix(s, "data:image/jpeg;base64,")
}

func applyMoveDeliveryReceiverToMoveData(moveData *functions.TrackingMoveEmailData, r *request.MoveDeliveryReceiver) {
	if moveData == nil || r == nil {
		return
	}
	moveData.HasDeliveryConfirmation = true
	moveData.ReceiverFullName = strings.TrimSpace(r.FullName)
	moveData.ReceiverRUT = strings.TrimSpace(r.RUT)
	moveData.ReceiverDeliveredAt = formatTrackingDateTimeForEmail(r.DeliveredAt)
	moveData.ReceiverSignatureDataURI = template.URL(strings.TrimSpace(r.SignaturePNGDataURL))
}

// ConfirmMoveDeliveryReport validates IMEIs share the same tracking log for each tracking_id, then sends move email and generates PDF like a move with with_delivery false.
func (s *deviceTrackingService) ConfirmMoveDeliveryReport(body *request.DeliveryConfirmMoveReportRequest, userID string) error {
	if !isAllowedSignatureDataURL(body.Receiver.SignaturePNGDataURL) {
		return ErrDeliveryConfirmInvalidSignature
	}
	for _, item := range body.Items {
		trackingID := strings.TrimSpace(item.TrackingID)
		if trackingID == "" {
			return fmt.Errorf("delivery_confirm: tracking_id is required")
		}
		seen := make(map[string]struct{})
		var docIDs []string
		var ref *responses.TrackingLog
		for _, rawImei := range item.Imeis {
			imei := strings.TrimSpace(rawImei)
			if imei == "" {
				continue
			}
			if _, dup := seen[imei]; dup {
				continue
			}
			seen[imei] = struct{}{}

			dt, err := s.deviceTrackingRepository.GetByImeiAndTrackingLogID(imei, trackingID)
			if err != nil {
				return err
			}
			if dt == nil {
				return fmt.Errorf("%w: imei=%s tracking_id=%s", ErrDeliveryConfirmNotFound, imei, trackingID)
			}
			tl := findTrackingLogInDevice(dt, trackingID)
			if tl == nil {
				return fmt.Errorf("%w: imei=%s tracking_id=%s", ErrDeliveryConfirmNotFound, imei, trackingID)
			}
			if ref == nil {
				copyTL := *tl
				ref = &copyTL
			} else if !moveDeliveryTrackingLogsEqual(*ref, *tl) {
				return fmt.Errorf("%w: imei=%s", ErrDeliveryConfirmInconsistent, imei)
			}
			docIDs = append(docIDs, dt.ID.Hex())
		}
		if len(docIDs) == 0 {
			return fmt.Errorf("delivery_confirm: no imeis resolved for tracking_id=%s", trackingID)
		}
		reqLog := responseTrackingLogToRequest(*ref)
		rcv := body.Receiver
		s.MoveTrackingNotification(docIDs, reqLog, userID, false, &rcv)
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
// sendTrackingNotificationMail sends the move/create notification with a link
// (documentURL) to the PDF report when it is already in S3, or with a fallback
// note when it is still pending. The PDF is never attached to the email itself.
func (s *deviceTrackingService) sendTrackingNotificationMail(rows []functions.TrackingEmailDeviceRow,
	companyId primitive.ObjectID, key string, receiver *request.MoveDeliveryReceiver, moveTrackingID string,
	documentURL string) {

	if len(rows) == 0 {
		return
	}
	toList, isEmpty := functions.GetEmails(false, companyId)
	if isEmpty {
		return
	}

	first := rows[0]
	var subject string
	var mainMessage string

	switch key {
	case utils.CREATE:
		subject = fmt.Sprintf("Subject: New Sample(s) received at LB Technology for %s %s %s",
			first.Country, first.Brand, first.CommercialModel)
		mainMessage = utils.CREATE_TRACKING_MAIN_MESSAGE
		moveData := createTrackingRowsToEmailData(rows, mainMessage, moveTrackingID)
		applyMoveReportLinkToMoveData(&moveData, documentURL)
		if err := functions.SendNotificationsMoveEmail(toList, subject, moveData, utils.TEMPLATE_TRACKING_MOVE_PATH, utils.LBOneTrackLogoPNG, documentURL); err != nil {
			log.Printf("tracking create notification email: %v", err)
		}
		return
	case utils.TRACKING_MOVE:
		if len(rows) > 1 {
			subject = fmt.Sprintf("Subject: Sample(s) has been moved of location to %s", first.NewLocation)
		} else {
			subject = fmt.Sprintf("Subject: Sample(s) has been moved of location to %s %s %s",
				first.Country, first.Brand, first.CommercialModel)
		}
		mainMessage = utils.MOVE_TRACKING_MAIN_MESSAGE
		moveData := moveTrackingRowsToEmailData(rows, mainMessage, moveTrackingID)
		if receiver != nil {
			applyMoveDeliveryReceiverToMoveData(&moveData, receiver)
		}
		applyMoveReportLinkToMoveData(&moveData, documentURL)
		if err := functions.SendNotificationsMoveEmail(toList, subject, moveData, utils.TEMPLATE_TRACKING_MOVE_PATH, utils.LBOneTrackLogoPNG, documentURL); err != nil {
			log.Printf("movement notification email: %v", err)
		}
		return
	default:
		return
	}
}

// applyMoveReportLinkToMoveData fills the link / pending-note fields the email
// template uses to render either the download button or the fallback message.
func applyMoveReportLinkToMoveData(d *functions.TrackingMoveEmailData, documentURL string) {
	d.DocumentURL = strings.TrimSpace(documentURL)
	if d.DocumentURL == "" {
		note := strings.TrimSpace(os.Getenv("MOVE_REPORT_PENDING_NOTE"))
		if note == "" {
			note = "The PDF report is being generated and will be available in the LB Technology platform shortly."
		}
		d.DocumentPendingNote = note
	} else {
		d.DocumentPendingNote = ""
	}
}

func moveTrackingRowsToEmailData(rows []functions.TrackingEmailDeviceRow, mainMessage string, trackingID string) functions.TrackingMoveEmailData {
	first := rows[0]
	return functions.TrackingMoveEmailData{
		IsMovement:          true,
		ClientName:          first.Client,
		Country:             first.Country,
		TrackingID:          strings.TrimSpace(trackingID),
		TotalSamples:        len(rows),
		PreviousLocation:    summarizePreviousLocations(rows),
		NewLocation:         first.NewLocation,
		RegistrationDate:    first.RegistrationDate,
		LBResponsible:       first.LBResponsible,
		ExternalResponsible: first.ExternalResponsible,
		RegisteredBy:        first.RegisteredBy,
		Comments:            first.Comments,
		Year:                time.Now().Year(),
		MainMessage:         mainMessage,
		Rows:                rows,
	}
}

func createTrackingRowsToEmailData(rows []functions.TrackingEmailDeviceRow, mainMessage string, trackingID string) functions.TrackingMoveEmailData {
	first := rows[0]
	return functions.TrackingMoveEmailData{
		IsMovement:          false,
		ClientName:          first.Client,
		Country:             first.Country,
		TrackingID:          strings.TrimSpace(trackingID),
		TotalSamples:        len(rows),
		PreviousLocation:    "—",
		NewLocation:         first.NewLocation,
		RegistrationDate:    first.RegistrationDate,
		LBResponsible:       first.LBResponsible,
		ExternalResponsible: first.ExternalResponsible,
		RegisteredBy:        first.RegisteredBy,
		Comments:            first.Comments,
		Year:                time.Now().Year(),
		MainMessage:         mainMessage,
		Rows:                rows,
	}
}

func sanitizeMoveReportIDForPath(id string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(id) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'), r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" {
		return "unknown"
	}
	return out
}

// moveReportDebugArtifactDir: default <cwd>/move-report-debug; override or set TRACKING_MOVE_REPORT_DEBUG_DIR=none to skip.
func moveReportDebugArtifactDir() string {
	raw := strings.TrimSpace(os.Getenv("TRACKING_MOVE_REPORT_DEBUG_DIR"))
	switch strings.ToLower(raw) {
	case "none", "false", "0", "off", "-", "disable", "disabled":
		return ""
	}
	if raw != "" {
		return raw
	}
	wd, err := os.Getwd()
	if err != nil {
		return filepath.Clean("move-report-debug")
	}
	return filepath.Join(wd, "move-report-debug")
}

// saveMoveReportDebugPDF mirrors only the generated PDF to debugDir.
func saveMoveReportDebugPDF(debugDir, trackingID string, pdfBytes []byte) {
	if debugDir == "" {
		return
	}
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		log.Printf("move report debug mkdir %s: %v", debugDir, err)
		return
	}
	stamp := time.Now().Format("20060102-150405")
	base := filepath.Join(debugDir, fmt.Sprintf("move-%s-%s", sanitizeMoveReportIDForPath(trackingID), stamp))
	if err := os.WriteFile(base+".pdf", pdfBytes, 0600); err != nil {
		log.Printf("move report debug write %s.pdf: %v", base, err)
	} else {
		log.Printf("move report debug wrote %s.pdf (%d bytes)", base, len(pdfBytes))
	}
}

// moveReportEmailKind selects which copy/subject family the orchestrator uses
// when sending the notification email (or skips the email entirely).
type moveReportEmailKind int

const (
	emailKindNone moveReportEmailKind = iota
	emailKindCreate
	emailKindMove
)

func emailKindToKey(k moveReportEmailKind) string {
	switch k {
	case emailKindCreate:
		return utils.CREATE
	case emailKindMove:
		return utils.TRACKING_MOVE
	default:
		return ""
	}
}

// renderMoveReportPDFBytes builds the same HTML as the move/registration email and renders it to PDF (Chrome).
// registration=true uses CREATE copy and Registration Summary layout; otherwise movement layout.
// Pass a non-zero timeout to override the default Chrome PDF timeout (used for fast vs slow phases).
func renderMoveReportPDFBytes(trackingID string,
	rows []functions.TrackingEmailDeviceRow, receiver *request.MoveDeliveryReceiver, registration bool, timeout time.Duration) ([]byte, error) {
	if trackingID == "" || len(rows) == 0 {
		return nil, fmt.Errorf("move report pdf: missing tracking id or rows")
	}
	if strings.TrimSpace(os.Getenv("TRACKING_SKIP_MOVE_PDF")) == "true" {
		return nil, fmt.Errorf("TRACKING_SKIP_MOVE_PDF=true")
	}
	var moveData functions.TrackingMoveEmailData
	if registration {
		moveData = createTrackingRowsToEmailData(rows, utils.CREATE_TRACKING_MAIN_MESSAGE, trackingID)
	} else {
		moveData = moveTrackingRowsToEmailData(rows, utils.MOVE_TRACKING_MAIN_MESSAGE, trackingID)
	}
	publicLogoURL := strings.TrimSpace(os.Getenv("TRACKING_LOGO_URL"))
	switch {
	case publicLogoURL != "":
		moveData.LogoDataURI = template.URL(publicLogoURL)
	case len(utils.LBOneTrackLogoPNG) > 0:
		moveData.LogoDataURI = template.URL("data:image/png;base64," +
			base64.StdEncoding.EncodeToString(utils.LBOneTrackLogoPNG))
	default:
		moveData.LogoDataURI = ""
	}
	if receiver != nil {
		applyMoveDeliveryReceiverToMoveData(&moveData, receiver)
	}
	htmlBytes, err := functions.RenderTrackingMoveEmailHTML(moveData, utils.TEMPLATE_TRACKING_MOVE_PATH)
	if err != nil {
		return nil, err
	}
	pdfBytes, pdfErr := movementHTMLToPDFWithTimeout(htmlBytes, timeout)
	if pdfErr != nil {
		return nil, pdfErr
	}
	return pdfBytes, nil
}

// uploadMoveReportToS3 uploads the PDF bytes to S3 with up to 3 retries.
// Returns the resulting public URL, or an error if every attempt failed.
func (s *deviceTrackingService) uploadMoveReportToS3(trackingID string, pdfBytes []byte, objectKey string) (string, error) {
	if s.storageService == nil {
		return "", fmt.Errorf("storage not configured")
	}
	if len(pdfBytes) == 0 {
		return "", fmt.Errorf("empty pdf bytes")
	}
	var url string
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		url, err = s.storageService.UploadFileWithKey(pdfBytes, objectKey)
		if err == nil {
			return url, nil
		}
		log.Printf("move report s3 upload attempt %d/3 (tracking_id=%s key=%s): %v", attempt, trackingID, objectKey, err)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 400 * time.Millisecond)
		}
	}
	return "", err
}

// moveReportS3Prefix returns the folder prefix for move-report PDFs (no leading/trailing slash).
func moveReportS3Prefix() string {
	p := strings.TrimSpace(os.Getenv("MOVE_REPORT_S3_PREFIX"))
	if p == "" {
		return "move-reports"
	}
	return strings.Trim(p, "/")
}

func digitsOnlyImei(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// moveReportCreationNumSegment picks the "number of IMEI" segment for the filename:
// one device → digits of that IMEI; several devices → count of samples (IMEI rows).
func moveReportCreationNumSegment(rows []functions.TrackingEmailDeviceRow) string {
	if len(rows) == 1 {
		d := digitsOnlyImei(rows[0].Imei)
		if d != "" {
			return d
		}
	}
	if len(rows) == 0 {
		return "1"
	}
	return strconv.Itoa(len(rows))
}

// moveReportS3ObjectKey builds the S3 object key for the move-report PDF.
//
// Creation (POST device-tracking): C-<imei or count>-<tracking_id>.pdf
//   — one IMEI: C-<15 dígitos>-<tracking_id>.pdf; varios: C-<cantidad>-<tracking_id>.pdf
//
// Movement (PUT move / POST confirm-delivery): M{C|D}-<tracking_id>.pdf
//   — MC = direct client (external_delivery false), MD = external delivery.
func moveReportS3ObjectKey(registration bool, trackingID string, externalDelivery bool, rows []functions.TrackingEmailDeviceRow) string {
	prefix := moveReportS3Prefix()
	tid := sanitizeMoveReportIDForPath(trackingID)
	if tid == "" {
		tid = "unknown"
	}
	if registration {
		num := moveReportCreationNumSegment(rows)
		return fmt.Sprintf("%s/C-%s-%s.pdf", prefix, num, tid)
	}
	channel := "C"
	if externalDelivery {
		channel = "D"
	}
	return fmt.Sprintf("%s/M%s-%s.pdf", prefix, channel, tid)
}

// persistMoveReportDocumentURL writes the (already-uploaded) URL into every
// tracking_log entry that matches the trackingID for the given device-tracking docs.
func (s *deviceTrackingService) persistMoveReportDocumentURL(docIDHexes []string, trackingID, url string) {
	if trackingID == "" || len(docIDHexes) == 0 || strings.TrimSpace(url) == "" {
		return
	}
	log.Printf("move report stored document_url=%s tracking_id=%s format=pdf", url, trackingID)
	oids := make([]primitive.ObjectID, 0, len(docIDHexes))
	for _, id := range docIDHexes {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		oids = append(oids, oid)
	}
	if len(oids) == 0 {
		return
	}
	if err := s.deviceTrackingRepository.SetTrackingLogDocumentURLByTrackingID(oids, trackingID, url); err != nil {
		log.Printf("move report document_url update: %v", err)
	}
}

func envIntPositive(name string, def int) int {
	s := strings.TrimSpace(os.Getenv(name))
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

// ensureMoveReportArtifacts is the orchestrator that, in a single goroutine,
// (a) tries to produce the PDF quickly so the notification email can carry a
// download link, and (b) keeps retrying in background until the PDF lands in S3
// and tracking_log.document_url is populated. The two concerns are independent:
// even if the email already went out without a link, the slow phase keeps
// retrying so the platform's "consult record" flow eventually shows the PDF.
func (s *deviceTrackingService) ensureMoveReportArtifacts(
	docIDHexes []string,
	trackingID string,
	rows []functions.TrackingEmailDeviceRow,
	receiver *request.MoveDeliveryReceiver,
	registration bool,
	externalDelivery bool,
	companyId primitive.ObjectID,
	emailKind moveReportEmailKind,
) {
	if trackingID == "" || len(rows) == 0 {
		return
	}

	idsCopy := append([]string(nil), docIDHexes...)
	emailKey := emailKindToKey(emailKind)
	objectKey := moveReportS3ObjectKey(registration, trackingID, externalDelivery, rows)

	fastAttempts := envIntPositive("MOVE_REPORT_PDF_FAST_ATTEMPTS", 3)
	fastTimeout := moveReportPDFFastTimeout()

	var pdfBytes []byte
	var lastErr error
	for attempt := 1; attempt <= fastAttempts; attempt++ {
		b, err := renderMoveReportPDFBytes(trackingID, rows, receiver, registration, fastTimeout)
		if err == nil && len(b) > 0 {
			pdfBytes = b
			break
		}
		lastErr = err
		log.Printf("move report pdf fast attempt %d/%d (tracking_id=%s): %v — %s",
			attempt, fastAttempts, trackingID, err, moveReportPDFErrHint(err))
		if attempt < fastAttempts {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}
	}

	if len(pdfBytes) > 0 {
		saveMoveReportDebugPDF(moveReportDebugArtifactDir(), trackingID, pdfBytes)
		url, upErr := s.uploadMoveReportToS3(trackingID, pdfBytes, objectKey)
		if upErr != nil {
			log.Printf("move report fast s3 upload (tracking_id=%s): %v", trackingID, upErr)
		} else {
			s.persistMoveReportDocumentURL(idsCopy, trackingID, url)
		}
		if emailKey != "" {
			s.sendTrackingNotificationMail(rows, companyId, emailKey, receiver, trackingID, url)
		}
		if upErr == nil {
			return
		}
		// Fall through to slow phase: PDF was rendered but upload failed; we
		// retry render+upload to make sure document_url gets populated.
	} else {
		log.Printf("move report pdf fast phase exhausted for tracking_id=%s after %d attempts: last err=%v",
			trackingID, fastAttempts, lastErr)
		if emailKey != "" {
			// Send the notification anyway so the customer is informed; the
			// link will appear in-platform once the slow phase succeeds.
			s.sendTrackingNotificationMail(rows, companyId, emailKey, receiver, trackingID, "")
		}
	}

	slowAttempts := envIntPositive("MOVE_REPORT_PDF_SLOW_ATTEMPTS", 8)
	baseBackoffSec := envIntPositive("MOVE_REPORT_PDF_SLOW_BACKOFF_BASE_SEC", 30)
	maxBackoffSec := envIntPositive("MOVE_REPORT_PDF_SLOW_MAX_BACKOFF_SEC", 1800)
	if maxBackoffSec < baseBackoffSec {
		maxBackoffSec = baseBackoffSec
	}
	slowTimeout := moveReportPDFTimeout()

	for attempt := 1; attempt <= slowAttempts; attempt++ {
		backoffSec := baseBackoffSec << uint(attempt-1)
		if backoffSec <= 0 || backoffSec > maxBackoffSec {
			backoffSec = maxBackoffSec
		}
		log.Printf("move report pdf slow attempt %d/%d for tracking_id=%s scheduled in %ds",
			attempt, slowAttempts, trackingID, backoffSec)
		time.Sleep(time.Duration(backoffSec) * time.Second)

		b, err := renderMoveReportPDFBytes(trackingID, rows, receiver, registration, slowTimeout)
		if err != nil || len(b) == 0 {
			log.Printf("move report pdf slow attempt %d/%d failed (tracking_id=%s): %v — %s",
				attempt, slowAttempts, trackingID, err, moveReportPDFErrHint(err))
			continue
		}
		saveMoveReportDebugPDF(moveReportDebugArtifactDir(), trackingID, b)
		url, upErr := s.uploadMoveReportToS3(trackingID, b, objectKey)
		if upErr != nil {
			log.Printf("move report slow s3 upload attempt %d/%d (tracking_id=%s): %v",
				attempt, slowAttempts, trackingID, upErr)
			continue
		}
		s.persistMoveReportDocumentURL(idsCopy, trackingID, url)
		log.Printf("move report pdf slow phase succeeded on attempt %d/%d for tracking_id=%s",
			attempt, slowAttempts, trackingID)
		return
	}

	log.Printf("move report pdf ABANDONED tracking_id=%s after %d slow attempts; document_url not stored",
		trackingID, slowAttempts)
}

func summarizePreviousLocations(rows []functions.TrackingEmailDeviceRow) string {
	seen := make(map[string]struct{})
	var list []string
	for _, r := range rows {
		v := strings.TrimSpace(r.PreviousLocation)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		list = append(list, v)
	}
	if len(list) == 0 {
		return "—"
	}
	return strings.Join(list, ", ")
}

// mergeCommentsAndProcessTypes appends process types as bullet lines after the comment when both exist.
func mergeCommentsAndProcessTypes(comment string, processTypes []string) string {
	comment = strings.TrimSpace(comment)
	var bullets strings.Builder
	for _, pt := range processTypes {
		pt = strings.TrimSpace(pt)
		if pt == "" {
			continue
		}
		if bullets.Len() > 0 {
			bullets.WriteString("\n")
		}
		bullets.WriteString("• ")
		bullets.WriteString(pt)
	}
	procBlock := bullets.String()
	switch {
	case comment != "" && procBlock != "":
		return comment + "\n" + procBlock
	case comment != "":
		return comment
	default:
		return procBlock
	}
}

func formatTrackingDateTimeForEmail(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	// DB stores an instant (UTC); show Chile (Santiago) civil time in notifications.
	return t.In(santiagoLocation()).Format("02/01/2006 15:04")
}

func (s *deviceTrackingService) buildTrackingEmailRowsSingleDevice(companyHex, deviceHex string, imeis []string,
	tl *request.TrackingLog, actingUserID string) ([]functions.TrackingEmailDeviceRow, error) {

	company, err := s.companyRepository.GetById(companyHex)
	if err != nil || company == nil {
		return nil, err
	}
	device, err := s.deviceRepository.GetById(deviceHex)
	if err != nil || device == nil {
		return nil, err
	}
	brand, err := s.brandRepository.GetById(device.Brand)
	if err != nil || brand == nil {
		return nil, err
	}
	actingUser, err := s.userRepository.GetByID(actingUserID)
	if err != nil {
		return nil, err
	}
	registeredBy := fmt.Sprintf("%s %s", actingUser.Name, actingUser.LastName)

	lbResp := fmt.Sprintf("%s %s", tl.InternalResponsible.Name, tl.InternalResponsible.LastName)
	commentsCell := mergeCommentsAndProcessTypes(tl.Comment, tl.ProcessTypes)
	regDate := formatTrackingDateTimeForEmail(tl.TrackingDate)

	rows := make([]functions.TrackingEmailDeviceRow, 0, len(imeis))
	for _, imei := range imeis {
		rows = append(rows, functions.TrackingEmailDeviceRow{
			Client:              company.Name,
			Country:             tl.Country.Name,
			Brand:               brand.Name,
			TechnicalModel:      device.TechnicalModel,
			CommercialModel:     device.CommercialModel,
			Imei:                imei,
			NewLocation:         tl.Location.Name,
			LBResponsible:       lbResp,
			ExternalResponsible: tl.Person.Name,
			Comments:            commentsCell,
			RegistrationDate:    regDate,
			RegisteredBy:        registeredBy,
		})
	}
	return rows, nil
}

func (s *deviceTrackingService) buildTrackingEmailRowsFromTrackingDocIDs(trackingDocIDs []string,
	tl *request.TrackingLog, actingUserID string) ([]functions.TrackingEmailDeviceRow, error) {

	actingUser, err := s.userRepository.GetByID(actingUserID)
	if err != nil {
		return nil, err
	}
	registeredBy := fmt.Sprintf("%s %s", actingUser.Name, actingUser.LastName)

	lbResp := fmt.Sprintf("%s %s", tl.InternalResponsible.Name, tl.InternalResponsible.LastName)
	commentsCell := mergeCommentsAndProcessTypes(tl.Comment, tl.ProcessTypes)
	regDate := formatTrackingDateTimeForEmail(tl.TrackingDate)

	rows := make([]functions.TrackingEmailDeviceRow, 0, len(trackingDocIDs))
	for _, id := range trackingDocIDs {
		dt, err := s.deviceTrackingRepository.GetById(id)
		if err != nil || dt == nil {
			continue
		}
		company, err := s.companyRepository.GetById(dt.Company.Hex())
		if err != nil || company == nil {
			continue
		}
		device, err := s.deviceRepository.GetById(dt.Device.Hex())
		if err != nil || device == nil {
			continue
		}
		brand, err := s.brandRepository.GetById(device.Brand)
		if err != nil || brand == nil {
			continue
		}
		prevLoc := ""
		if len(dt.TrackingLogs) >= 2 {
			prevLoc = dt.TrackingLogs[len(dt.TrackingLogs)-2].Location.Name
		}
		rows = append(rows, functions.TrackingEmailDeviceRow{
			Client:              company.Name,
			Country:             tl.Country.Name,
			Brand:               brand.Name,
			TechnicalModel:      device.TechnicalModel,
			CommercialModel:     device.CommercialModel,
			Imei:                dt.Imei,
			PreviousLocation:    prevLoc,
			NewLocation:         tl.Location.Name,
			LBResponsible:       lbResp,
			ExternalResponsible: tl.Person.Name,
			Comments:            commentsCell,
			RegistrationDate:    regDate,
			RegisteredBy:        registeredBy,
		})
	}
	return rows, nil
}

func (s *deviceTrackingService) MoveTrackingNotification(deviceTrackingsId []string,
	trackingLog request.TrackingLog, userID string, withDelivery bool, receiver *request.MoveDeliveryReceiver) {

	companyGrouped := make(map[primitive.ObjectID][]string)
	for _, id := range deviceTrackingsId {
		deviceTracking, err := s.deviceTrackingRepository.GetById(id)
		if err != nil || deviceTracking == nil {
			continue
		}
		companyGrouped[deviceTracking.Company] = append(companyGrouped[deviceTracking.Company], id)
	}
	for companyId, docIDs := range companyGrouped {
		rows, err := s.buildTrackingEmailRowsFromTrackingDocIDs(docIDs, &trackingLog, userID)
		if err != nil || len(rows) == 0 {
			continue
		}
		emailKind := emailKindMove
		if withDelivery {
			// withDelivery=true means the customer will confirm delivery later
			// (see ConfirmMoveDeliveryReport); the move email is suppressed here
			// but the PDF/document_url must still be produced for the platform.
			emailKind = emailKindNone
		}
		go s.ensureMoveReportArtifacts(docIDs, trackingLog.TrackingID, rows, receiver, false, trackingLog.ExternalDelivery, companyId, emailKind)
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
