package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/utils/mapping"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IMultibandaReportService interface {
	GetForm(multibandaID, userID string) (*responses.MultibandaReportForm, error)
	SaveDraft(multibandaID string, req *request.MultibandaReportSave, userID string) (*responses.MultibandaReportForm, error)
	Generate(multibandaID string, req *request.MultibandaReportSave, userID string) (*responses.MultibandaReportGenerated, error)
}

type multibandaReportService struct {
	reportRepository     repositoriesMultibandaReport
	multibandaRepository repositoriesMultibanda
	deviceRepository     repositoriesDevice
	userRepository       repositoriesUser
	storageService       IStorageService
}

// Narrow local interfaces keep this service decoupled from the full repository
// surfaces it does not use.
type repositoriesMultibandaReport interface {
	GetByMultibanda(primitive.ObjectID) (*models.MultibandaReport, error)
	Upsert(*models.MultibandaReport) error
	MarkGenerated(primitive.ObjectID, string, string, time.Time) error
}

type repositoriesMultibanda interface {
	GetByIdExpanded(primitive.ObjectID) (*responses.MultibandaExpanded, error)
}

type repositoriesDevice interface {
	GetById(string) (*responses.Device, error)
}

type repositoriesUser interface {
	GetByID(string) (*responses.User, error)
}

func NewMultibandaReportService(
	reportRepository repositoriesMultibandaReport,
	multibandaRepository repositoriesMultibanda,
	deviceRepository repositoriesDevice,
	userRepository repositoriesUser,
	storageService IStorageService,
) IMultibandaReportService {
	return &multibandaReportService{
		reportRepository:     reportRepository,
		multibandaRepository: multibandaRepository,
		deviceRepository:     deviceRepository,
		userRepository:       userRepository,
		storageService:       storageService,
	}
}

// ---- Read ----

func (s *multibandaReportService) GetForm(multibandaID, userID string) (*responses.MultibandaReportForm, error) {
	ctx, err := s.loadContext(multibandaID, userID)
	if err != nil {
		return nil, err
	}

	form := &responses.MultibandaReportForm{
		MultibandaID:      ctx.multibandaOID,
		MultibandaType:    ctx.multibanda.Type,
		Status:            enums.MultibandaReportStatusDraft,
		IncludesSimlock:   ctx.scope.IncludesSimlock,
		IncludesMultiband: ctx.scope.IncludesMultiband,
		Prefilled:         ctx.prefilled,
		Catalogs:          mapping.MultibandaReportCatalogs(),
	}

	if ctx.report != nil {
		form.Status = ctx.report.Status
		form.Saved = mapping.MultibandaReportModelToSaved(ctx.report)
		form.ReportURL = ctx.report.ReportURL
		if !ctx.report.GeneratedAt.IsZero() {
			form.GeneratedAt = ctx.report.GeneratedAt.UTC().Format(time.RFC3339)
		}
	}

	// Always present, even with nothing saved: the screen needs the full list
	// of blockers to render at 0%.
	form.Validation = utils.BuildMultibandaReportValidation(
		mapping.MultibandaReportModelToSaveRequest(ctx.report), ctx.scope)

	return form, nil
}

// ---- Draft ----

func (s *multibandaReportService) SaveDraft(
	multibandaID string,
	req *request.MultibandaReportSave,
	userID string,
) (*responses.MultibandaReportForm, error) {
	ctx, err := s.loadContext(multibandaID, userID)
	if err != nil {
		return nil, err
	}
	if err := utils.ValidateMultibandaReportDraft(req, ctx.scope); err != nil {
		return nil, err
	}

	// A report that already produced a PDF stays "generated" while it is
	// edited; regeneration is what refreshes the stored document.
	status := enums.MultibandaReportStatusDraft
	if ctx.report != nil && ctx.report.Status == enums.MultibandaReportStatusGenerated {
		status = enums.MultibandaReportStatusGenerated
	}

	report := mapping.MultibandaReportRequestToModel(req, ctx.multibandaOID, ctx.multibanda.Type, status)
	if err := s.reportRepository.Upsert(report); err != nil {
		return nil, err
	}

	return s.GetForm(multibandaID, userID)
}

// ---- Generate ----

func (s *multibandaReportService) Generate(
	multibandaID string,
	req *request.MultibandaReportSave,
	userID string,
) (*responses.MultibandaReportGenerated, error) {
	ctx, err := s.loadContext(multibandaID, userID)
	if err != nil {
		return nil, err
	}

	// Server-side completeness check, as the spec requires in addition to the
	// frontend keeping the button disabled.
	if err := utils.ValidateMultibandaReportForGeneration(req, ctx.scope); err != nil {
		return nil, err
	}

	report := mapping.MultibandaReportRequestToModel(
		req, ctx.multibandaOID, ctx.multibanda.Type, enums.MultibandaReportStatusGenerated)
	if err := s.reportRepository.Upsert(report); err != nil {
		return nil, err
	}

	pdfData := s.buildPDFData(context.Background(), ctx, report)
	pdfBytes, err := functions.BuildMultibandaReportPDF(pdfData)
	if err != nil {
		return nil, fmt.Errorf("generate multibanda report pdf: %w", err)
	}

	fileName := functions.BuildMultibandaReportFileName(
		ctx.prefilled.CommercialName, ctx.prefilled.SoftwareVersion, report.DeviceInfo.TestDate)
	objectKey := multibandaReportObjectKey(ctx.multibandaOID.Hex(), ctx.prefilled.TechnicalModel)

	if s.storageService == nil {
		return nil, fmt.Errorf("storage not configured")
	}
	baseURL, err := s.storageService.UploadFileWithKeyAndName(pdfBytes, objectKey, fileName)
	if err != nil {
		return nil, fmt.Errorf("upload multibanda report: %w", err)
	}

	// Regeneration overwrites the same key, so bust caches the same way the
	// OABI certificate does.
	reportURL := functions.CacheBustCertificateURL(baseURL)
	generatedAt := time.Now()
	if err := s.reportRepository.MarkGenerated(ctx.multibandaOID, reportURL, userID, generatedAt); err != nil {
		return nil, err
	}

	return &responses.MultibandaReportGenerated{
		ReportURL:   reportURL,
		FileName:    fileName,
		GeneratedAt: generatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// ---- Shared loading ----

type multibandaReportContext struct {
	multibandaOID primitive.ObjectID
	multibanda    *responses.MultibandaExpanded
	report        *models.MultibandaReport
	scope         utils.MultibandaReportScope
	prefilled     responses.MultibandaReportPrefilled
}

func (s *multibandaReportService) loadContext(multibandaID, userID string) (*multibandaReportContext, error) {
	oid, err := primitive.ObjectIDFromHex(multibandaID)
	if err != nil {
		return nil, utils.NewValidationError("invalid multibanda id")
	}

	user, err := s.userRepository.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	multibanda, err := s.multibandaRepository.GetByIdExpanded(oid)
	if err != nil {
		return nil, err
	}
	if multibanda == nil {
		return nil, utils.NewValidationError("multibanda not found")
	}

	if !functions.UserHasClientAccess(user, multibanda.Company.ID) {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}
	if !user.IsInternal && multibanda.Company.ID != user.Company {
		return nil, fmt.Errorf("%w", utils.ErrorForbidden)
	}

	report, err := s.reportRepository.GetByMultibanda(oid)
	if err != nil {
		return nil, err
	}

	prefilled := buildReportPrefilled(multibanda)
	includesInitialBlocks := enums.ReportIncludesSimlockAndMultiband(multibanda.Type)

	return &multibandaReportContext{
		multibandaOID: oid,
		multibanda:    multibanda,
		report:        report,
		prefilled:     prefilled,
		scope: utils.MultibandaReportScope{
			IncludesSimlock:   includesInitialBlocks,
			IncludesMultiband: includesInitialBlocks,
			SupportsGSM:       prefilled.SupportsGSM,
			SupportsUMTS:      prefilled.SupportsUMTS,
			SupportsLTE:       prefilled.SupportsLTE,
			Supports5G:        prefilled.Supports5G,
		},
	}, nil
}

func buildReportPrefilled(multibanda *responses.MultibandaExpanded) responses.MultibandaReportPrefilled {
	sar := ""
	if multibanda.Device.SarValue > 0 {
		sar = strconv.FormatFloat(multibanda.Device.SarValue, 'f', 2, 64) + " W/KG"
	}

	return responses.MultibandaReportPrefilled{
		Manufacturer:   strings.TrimSpace(multibanda.Brand.Name),
		CommercialName: strings.TrimSpace(multibanda.Device.CommercialModel),
		TechnicalModel: strings.TrimSpace(multibanda.Device.TechnicalModel),
		// HW/SW versions come from the multibanda process, not the device
		// catalog: they identify the exact build under evaluation.
		HardwareVersion: strings.TrimSpace(multibanda.HardwareVersion),
		SoftwareVersion: strings.TrimSpace(multibanda.SoftwareVersion),
		SARValue:        sar,
		OperativeSystemType: functions.FormatMultibandaOsVersion(
			multibanda.OsVersionView, multibanda.Device.PlatformOs, multibanda.OsVersion),
		SupportsGSM:  multibanda.Device.NetworkGsm,
		SupportsUMTS: multibanda.Device.NetworkWcdma,
		SupportsLTE:  multibanda.Device.NetworkLte,
		Supports5G:   multibanda.Device.Network5g,
	}
}

// ---- PDF data assembly ----

func (s *multibandaReportService) buildPDFData(
	ctx context.Context,
	rc *multibandaReportContext,
	report *models.MultibandaReport,
) functions.MultibandaReportPDFData {
	now := time.Now()
	stampImage, stampLabel := stampImageBytes(report.DeviceInfo.StampCode)

	data := functions.MultibandaReportPDFData{
		ProcessTypeLabel:       enums.MultibandaTypeLabels[rc.multibanda.Type],
		ReportDate:             functions.FormatShipmentControlEmailDate(now),
		Manufacturer:           rc.prefilled.Manufacturer,
		CommercialName:         rc.prefilled.CommercialName,
		TechnicalModel:         rc.prefilled.TechnicalModel,
		HardwareVersion:        rc.prefilled.HardwareVersion,
		SoftwareVersion:        rc.prefilled.SoftwareVersion,
		SARValue:               rc.prefilled.SARValue,
		CBSPackage:             report.DeviceInfo.CBSPackage,
		GooglePlaySystemUpdate: report.DeviceInfo.GooglePlaySystemUpdate,
		OperativeSystemType:    rc.prefilled.OperativeSystemType,
		PreferredNetwork:       report.DeviceInfo.PreferredNetwork,
		FMRadio:                report.DeviceInfo.FMRadio,
		TestDate:               functions.FormatShipmentControlEmailDate(report.DeviceInfo.TestDate),
		IMEI:                   report.DeviceInfo.IMEI,
		SerialNumber:           report.DeviceInfo.SerialNumber,
		StampImage:             stampImage,
		StampLabel:             stampLabel,
		CarriersTested:         report.CarriersTested,
		IncludesSimlock:        rc.scope.IncludesSimlock,
		IncludesMultiband:      rc.scope.IncludesMultiband,
		FMRadioTestApplies:     report.DeviceInfo.FMRadio == enums.ReportFMRadioSupported,
		FMRadioTestResult:      report.FMRadioResult,
		FMRadioTestComment:     report.FMRadioComment,
		GeneratedAt:            now,
		Year:                   now.Year(),
	}

	if data.IncludesSimlock {
		data.SimlockRows = buildSimlockRows(report.SimlockResults, report.CarriersTested)
	}
	if data.IncludesMultiband {
		data.BandGroups = buildBandGroups(report.BandResults)
	}

	scenarios := enums.SAEScenariosFor(report.SAEScenario)
	data.SAEScenario = saeScenarioLabel(report.SAEScenario)
	data.SAEBlocks = buildSAEBlocks(report.SAEResults, scenarios)

	// SW Version is printed in Device Information; the other two under SAE.
	data.SWVersionPhoto = s.fetchEvidenceImage(ctx, report.Evidence, enums.EvidenceSWVersion)
	data.SAEEvidence = s.buildEvidenceItems(ctx, report.Evidence)

	return data
}

// buildSimlockRows walks the fixed test catalog against the carriers actually
// tested, so the PDF keeps a stable order regardless of payload ordering and
// never prints a carrier the device was not evaluated against.
func buildSimlockRows(
	results []models.MultibandaReportSimlockResult,
	carriersTested []string,
) []functions.MultibandaReportSimlockRow {
	index := make(map[string]models.MultibandaReportSimlockResult, len(results))
	for _, r := range results {
		index[r.TestID+"|"+r.Carrier] = r
	}

	mnc := make(map[string]string, len(enums.ReportSimlockCarriers))
	for _, c := range enums.ReportSimlockCarriers {
		mnc[c.Name] = c.MNC
	}

	rows := make([]functions.MultibandaReportSimlockRow, 0, len(results))
	for _, test := range enums.ReportSimlockTests {
		for _, carrier := range carriersTested {
			r := index[test.ID+"|"+carrier]
			rows = append(rows, functions.MultibandaReportSimlockRow{
				TestID:  test.ID,
				Name:    test.Name,
				Carrier: carrier,
				MNC:     mnc[carrier],
				Result:  r.Result,
				Comment: r.Comment,
			})
		}
	}
	return rows
}

func buildBandGroups(results []models.MultibandaReportBandResult) []functions.MultibandaReportBandGroup {
	index := make(map[string]models.MultibandaReportBandResult, len(results))
	for _, r := range results {
		index[r.Technology+"|"+r.Band] = r
	}

	groups := make([]functions.MultibandaReportBandGroup, 0, len(enums.ReportBandCatalog))
	for _, tech := range enums.ReportBandCatalog {
		group := functions.MultibandaReportBandGroup{Technology: tech.Label}
		for _, band := range tech.Bands {
			r := index[tech.Code+"|"+band]
			group.Bands = append(group.Bands, functions.MultibandaReportBandRow{
				Band: band, Result: r.Result, Comment: r.Comment,
			})
		}
		groups = append(groups, group)
	}
	return groups
}

func buildSAEBlocks(results []models.MultibandaReportSAEResult, scenarios []string) []functions.MultibandaReportSAEBlock {
	index := make(map[string]models.MultibandaReportSAEResult, len(results))
	for _, r := range results {
		index[r.TestID+"|"+r.Scenario+"|"+r.Channel+"|"+r.Operator] = r
	}

	blocks := make([]functions.MultibandaReportSAEBlock, 0, len(scenarios)*len(enums.SAEChannels))
	for _, scenario := range scenarios {
		for _, channel := range enums.SAEChannels {
			block := functions.MultibandaReportSAEBlock{
				ScenarioLabel: enums.SAEScenarioLabels[scenario],
				Channel:       channel,
				Operators:     enums.SAEOperators,
			}
			for _, test := range enums.ReportSAETests {
				row := functions.MultibandaReportSAERow{TestID: test.ID, Name: test.Name}
				for _, operator := range enums.SAEOperators {
					r := index[test.ID+"|"+scenario+"|"+channel+"|"+operator]
					row.Results = append(row.Results, r.Result)
					row.Comments = append(row.Comments, r.Comment)
				}
				block.Rows = append(block.Rows, row)
			}
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func (s *multibandaReportService) buildEvidenceItems(
	ctx context.Context,
	evidence []models.MultibandaReportEvidence,
) []functions.MultibandaReportEvidenceItem {
	index := make(map[string]models.MultibandaReportEvidence, len(evidence))
	for _, e := range evidence {
		index[e.EvidenceType] = e
	}

	items := make([]functions.MultibandaReportEvidenceItem, 0, len(enums.RequiredSAEEvidenceTypes))
	for _, evidenceType := range enums.RequiredSAEEvidenceTypes {
		e, ok := index[evidenceType]
		if !ok {
			continue
		}
		items = append(items, functions.MultibandaReportEvidenceItem{
			Label:         enums.EvidenceLabels[evidenceType],
			ScenarioLabel: enums.SAEScenarioLabels[e.Scenario],
			Operator:      e.Operator,
			Image:         fetchReportImage(ctx, e.URL),
		})
	}
	return items
}

// fetchEvidenceImage pulls a single screenshot by type, used for the SW Version
// capture that belongs to Device Information rather than the SAE block.
func (s *multibandaReportService) fetchEvidenceImage(
	ctx context.Context,
	evidence []models.MultibandaReportEvidence,
	evidenceType string,
) []byte {
	for _, e := range evidence {
		if e.EvidenceType == evidenceType {
			return fetchReportImage(ctx, e.URL)
		}
	}
	return nil
}

func saeScenarioLabel(selection string) string {
	switch selection {
	case enums.SAEScenarioBoth:
		return "Both"
	default:
		return enums.SAEScenarioLabels[selection]
	}
}
