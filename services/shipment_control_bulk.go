package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *shipmentControlService) BulkValidate(
	fileHeader *multipart.FileHeader,
	userID string,
	companyParam string,
) (*responses.ShipmentControlBulkValidateResponse, error) {
	if err := s.requireProfileClaim(userID, enums.CanWriteShipmentControl); err != nil {
		return nil, err
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	companyID, err := s.resolveCompanyScope(user, companyParam)
	if err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	parsedRows, err := functions.ParseShipmentControlBulkCSVFromBytes(raw)
	if err != nil {
		return nil, utils.NewValidationError(err.Error())
	}

	response := &responses.ShipmentControlBulkValidateResponse{
		Summary: responses.ShipmentControlBulkValidateSummary{
			Total: len(parsedRows),
		},
		Rows: make([]responses.ShipmentControlBulkValidateRow, 0, len(parsedRows)),
	}

	for _, row := range parsedRows {
		result := s.validateBulkCSVRow(user, companyID, row)
		response.Rows = append(response.Rows, result)
		if result.Status == "valid" {
			response.Summary.Valid++
		} else {
			response.Summary.Invalid++
		}
	}

	return response, nil
}

func (s *shipmentControlService) BulkConfirm(
	req *request.ShipmentControlBulkConfirm,
	userID string,
	companyParam string,
	countryParam string,
) (*responses.ShipmentControlBulkConfirmResponse, error) {
	if err := s.requireProfileClaim(userID, enums.CanWriteShipmentControl); err != nil {
		return nil, err
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}

	companyID, err := s.resolveCompanyScope(user, companyParam)
	if err != nil {
		return nil, err
	}

	countryID, err := s.resolveBulkConfirmCountry(user, countryParam)
	if err != nil {
		return nil, err
	}

	country, err := s.countryRepository.GetById(countryID.Hex())
	if err != nil || country == nil {
		return nil, utils.NewValidationError("country not found")
	}

	response := &responses.ShipmentControlBulkConfirmResponse{
		Summary: responses.ShipmentControlBulkConfirmSummary{},
		Results: make([]responses.ShipmentControlBulkConfirmResult, 0, len(req.Rows)),
	}

	for _, row := range req.Rows {
		result := s.confirmBulkRow(user, companyID, countryID, row)
		response.Results = append(response.Results, result)
		if result.Status == "created" {
			response.Summary.Created++
		} else {
			response.Summary.Failed++
		}
	}

	return response, nil
}

func (s *shipmentControlService) validateBulkCSVRow(
	user *responses.User,
	companyID primitive.ObjectID,
	row functions.ShipmentControlBulkCSVRow,
) responses.ShipmentControlBulkValidateRow {
	result := responses.ShipmentControlBulkValidateRow{
		RowNumber:       row.RowNumber,
		Client:          row.Client,
		ReworkNumber:    row.ReworkNumber,
		TechnicalModel:  row.TechnicalModel,
		SoftwareVersion: row.SoftwareVersion,
		ImeiQuantity:    row.ImeiQuantity,
		Status:          "invalid",
	}

	fieldErrors := functions.ValidateShipmentControlBulkCSVFields(row)
	if len(fieldErrors) > 0 {
		result.Errors = fieldErrors
		return result
	}

	matchErrors, multibanda := s.resolveBulkMultibanda(user, companyID, row.TechnicalModel, row.SoftwareVersion)
	if len(matchErrors) > 0 {
		result.Errors = matchErrors
		return result
	}

	result.Status = "valid"
	result.MultibandaID = multibanda.ID.Hex()
	result.SubtelCertificateNumber = multibanda.SubtelCertificateNumber
	result.Device = &responses.ShipmentControlBulkValidateDevicePreview{
		CommercialModel: multibanda.Device.CommercialModel,
		HardwareVersion: multibanda.HardwareVersion,
	}

	return result
}

func (s *shipmentControlService) resolveBulkMultibanda(
	user *responses.User,
	companyID primitive.ObjectID,
	technicalModel string,
	softwareVersion string,
) ([]string, *responses.MultibandaExpanded) {
	devices, err := s.deviceRepository.FindByTechnicalModel(technicalModel)
	if err != nil {
		return []string{functions.BulkErrorMultibandaNotFound}, nil
	}

	multibandaCount := 0
	var multibanda *responses.MultibandaExpanded
	subtelCertificate := ""

	if len(devices) == 1 {
		multibandas, lookupErr := s.multibandaRepository.FindApprovedByCompanyDeviceSoftwareVersion(
			companyID,
			devices[0].ID,
			softwareVersion,
			user.Brands,
		)
		if lookupErr != nil {
			return []string{functions.BulkErrorMultibandaNotFound}, nil
		}
		multibandaCount = len(multibandas)
		if multibandaCount == 1 {
			multibanda = multibandas[0]
			subtelCertificate = multibanda.SubtelCertificateNumber
		}
	}

	matchErrors, ok := functions.EvaluateBulkMultibandaMatch(len(devices), multibandaCount, subtelCertificate)
	if !ok {
		return matchErrors, nil
	}

	if !functions.UserHasClientAccess(user, multibanda.Company.ID) {
		return []string{functions.BulkErrorForbidden}, nil
	}
	if !user.IsInternal && multibanda.Company.ID != user.Company {
		return []string{functions.BulkErrorForbidden}, nil
	}
	if !userHasBrandAccess(user, multibanda.Brand.ID) {
		return []string{functions.BulkErrorForbidden}, nil
	}

	return nil, multibanda
}

func (s *shipmentControlService) confirmBulkRow(
	user *responses.User,
	companyID primitive.ObjectID,
	countryID primitive.ObjectID,
	row request.ShipmentControlBulkConfirmRow,
) responses.ShipmentControlBulkConfirmResult {
	result := responses.ShipmentControlBulkConfirmResult{
		RowNumber: row.RowNumber,
		Status:    "failed",
	}

	errors := []string{}
	if strings.TrimSpace(row.ImeiFileUrl) == "" {
		errors = append(errors, functions.BulkErrorMissingImeiFileURL)
	}
	if row.ImeiQuantity <= 0 {
		errors = append(errors, functions.BulkErrorInvalidQty)
	}

	multibandaID, err := primitive.ObjectIDFromHex(strings.TrimSpace(row.MultibandaID))
	if err != nil {
		errors = append(errors, functions.BulkErrorInvalidMultibandaID)
	}
	if len(errors) > 0 {
		result.Errors = errors
		return result
	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(multibandaID)
	if err != nil {
		result.Errors = []string{functions.BulkErrorMultibandaNotFound}
		return result
	}
	if multibanda == nil {
		result.Errors = []string{functions.BulkErrorMultibandaNotFound}
		return result
	}

	result.TechnicalModel = multibanda.Device.TechnicalModel
	result.SoftwareVersion = multibanda.SoftwareVersion

	validationErrors := s.validateBulkMultibandaForConfirm(user, companyID, multibanda)
	if len(validationErrors) > 0 {
		result.Errors = validationErrors
		return result
	}

	createReq := &request.ShipmentControl{
		MultibandaID: multibandaID.Hex(),
		ImeiQuantity: row.ImeiQuantity,
		ImeiFileUrl:  row.ImeiFileUrl,
		Client:       row.Client,
		ReworkNumber: row.ReworkNumber,
		Country:      countryID.Hex(),
	}

	shipmentControl := mapping.ShipmentControlRequestToShipmentControl(
		createReq,
		multibandaID,
		multibanda.Company.ID,
		countryID,
	)

	id, err := s.shipmentControlRepository.Create(shipmentControl)
	if err != nil {
		result.Errors = []string{fmt.Sprintf("create_failed: %v", err)}
		return result
	}

	shipmentControlID, _ := primitive.ObjectIDFromHex(id)
	go s.ShipmentControlNotification(shipmentControlID, nil, utils.CREATE, user.ID.Hex())

	result.Status = "created"
	result.ShipmentControlID = id
	return result
}

func (s *shipmentControlService) validateBulkMultibandaForConfirm(
	user *responses.User,
	companyID primitive.ObjectID,
	multibanda *responses.MultibandaExpanded,
) []string {
	if multibanda.Status != enums.HomologationStatus_value["APPROVED"] {
		return []string{functions.BulkErrorMultibandaNotApproved}
	}
	if multibanda.RequestDelete {
		return []string{functions.BulkErrorMultibandaDeletePending}
	}
	if multibanda.Company.ID != companyID {
		return []string{functions.BulkErrorMultibandaNotFound}
	}
	if strings.TrimSpace(multibanda.SubtelCertificateNumber) == "" {
		return []string{functions.BulkErrorCertificateMissing}
	}
	if !functions.UserHasClientAccess(user, multibanda.Company.ID) {
		return []string{functions.BulkErrorForbidden}
	}
	if !user.IsInternal && multibanda.Company.ID != user.Company {
		return []string{functions.BulkErrorForbidden}
	}
	if !userHasBrandAccess(user, multibanda.Brand.ID) {
		return []string{functions.BulkErrorForbidden}
	}
	return nil
}

func (s *shipmentControlService) resolveBulkConfirmCountry(user *responses.User, countryParam string) (primitive.ObjectID, error) {
	if user.IsInternal {
		if strings.TrimSpace(countryParam) == "" {
			return primitive.NilObjectID, utils.NewValidationError("country query parameter is required")
		}
		return parseReferenceID("country", countryParam)
	}

	if user.Company == primitive.NilObjectID {
		return primitive.NilObjectID, utils.NewValidationError("user company is not configured")
	}

	chile, err := s.countryRepository.GetByName(enums.ShipmentControlExternalCountryName)
	if err != nil || chile == nil {
		return primitive.NilObjectID, utils.NewValidationError("Chile country not found")
	}

	return chile.ID, nil
}
