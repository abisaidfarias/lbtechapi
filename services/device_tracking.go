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

	idsCopy := append([]string(nil), docIDs...)
	tid := trackingID
	rws := rows
	cid := companyId
	go func() {
		var pdfBytes []byte
		if tid != "" && len(rws) > 0 {
			b, err := renderMoveReportPDFBytes(tid, rws, nil, true)
			if err != nil {
				log.Printf("create report pdf: %v — %s", err, moveReportPDFErrHint(err))
			} else {
				pdfBytes = b
				saveMoveReportDebugPDF(moveReportDebugArtifactDir(), tid, pdfBytes)
			}
		}
		pdfName := fmt.Sprintf("move-report-%s.pdf", sanitizeMoveReportIDForPath(tid))
		s.sendTrackingNotificationMail(rws, cid, utils.CREATE, nil, tid, pdfBytes, pdfName)
		s.persistMoveReportDocumentURL(idsCopy, tid, pdfBytes)
	}()

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
func (s *deviceTrackingService) sendTrackingNotificationMail(rows []functions.TrackingEmailDeviceRow,
	companyId primitive.ObjectID, key string, receiver *request.MoveDeliveryReceiver, moveTrackingID string,
	pdfAttachment []byte, pdfFileName string) {

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
		if err := functions.SendNotificationsMoveEmail(toList, subject, moveData, utils.TEMPLATE_TRACKING_MOVE_PATH, utils.LBOneTrackLogoPNG, pdfAttachment, pdfFileName); err != nil {
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
		if err := functions.SendNotificationsMoveEmail(toList, subject, moveData, utils.TEMPLATE_TRACKING_MOVE_PATH, utils.LBOneTrackLogoPNG, pdfAttachment, pdfFileName); err != nil {
			log.Printf("movement notification email: %v", err)
		}
		return
	default:
		return
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

// renderMoveReportPDFBytes builds the same HTML as the move/registration email and renders it to PDF (Chrome).
// registration=true uses CREATE copy and Registration Summary layout; otherwise movement layout.
func renderMoveReportPDFBytes(trackingID string,
	rows []functions.TrackingEmailDeviceRow, receiver *request.MoveDeliveryReceiver, registration bool) ([]byte, error) {
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
	pdfBytes, pdfErr := movementHTMLToPDF(htmlBytes)
	if pdfErr != nil {
		return nil, pdfErr
	}
	return pdfBytes, nil
}

func (s *deviceTrackingService) persistMoveReportDocumentURL(docIDHexes []string, trackingID string, pdfBytes []byte) {
	if trackingID == "" || len(docIDHexes) == 0 || len(pdfBytes) == 0 {
		return
	}
	if s.storageService == nil {
		log.Printf("move report: storage not configured, document_url not saved (tracking_id=%s)", trackingID)
		return
	}
	var url string
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		url, err = s.storageService.UploadFile(pdfBytes)
		if err == nil {
			break
		}
		log.Printf("move report s3 upload attempt %d/3: %v", attempt, err)
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 400 * time.Millisecond)
		}
	}
	if err != nil {
		log.Printf("move report s3 upload: %v", err)
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
		tid := trackingLog.TrackingID
		idsCopy := append([]string(nil), docIDs...)
		rws := rows
		cid := companyId
		wDel := withDelivery
		rcv := receiver
		go func() {
			var pdfBytes []byte
			if tid != "" && len(rws) > 0 {
				b, err := renderMoveReportPDFBytes(tid, rws, rcv, false)
				if err != nil {
					log.Printf("move report pdf: %v — %s", err, moveReportPDFErrHint(err))
				} else {
					pdfBytes = b
					saveMoveReportDebugPDF(moveReportDebugArtifactDir(), tid, pdfBytes)
				}
			}
			pdfName := fmt.Sprintf("move-report-%s.pdf", sanitizeMoveReportIDForPath(tid))
			if !wDel {
				s.sendTrackingNotificationMail(rws, cid, utils.TRACKING_MOVE, rcv, tid, pdfBytes, pdfName)
			}
			s.persistMoveReportDocumentURL(idsCopy, tid, pdfBytes)
		}()
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
