package mapping

import (
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MultibandaReportRequestToModel converts the save payload into the stored
// report. Status is decided by the caller (draft vs generated).
func MultibandaReportRequestToModel(
	req *request.MultibandaReportSave,
	multibandaID primitive.ObjectID,
	multibandaType string,
	status string,
) *models.MultibandaReport {
	report := &models.MultibandaReport{
		Multibanda:     multibandaID,
		Status:         status,
		MultibandaType: multibandaType,
		DeviceInfo: models.MultibandaReportDeviceInfo{
			CBSPackage:             strings.TrimSpace(req.DeviceInfo.CBSPackage),
			GooglePlaySystemUpdate: strings.TrimSpace(req.DeviceInfo.GooglePlaySystemUpdate),
			PreferredNetwork:       strings.TrimSpace(req.DeviceInfo.PreferredNetwork),
			FMRadio:                strings.TrimSpace(req.DeviceInfo.FMRadio),
			IMEI:                   strings.TrimSpace(req.DeviceInfo.IMEI),
			SerialNumber:           strings.TrimSpace(req.DeviceInfo.SerialNumber),
			StampCode:              strings.TrimSpace(req.DeviceInfo.StampCode),
		},
		SAEScenario:    strings.TrimSpace(req.SAEScenario),
		CarriersTested: trimAll(req.CarriersTested),
		FMRadioResult:  strings.TrimSpace(req.FMRadioResult),
		FMRadioComment: strings.TrimSpace(req.FMRadioComment),
	}
	if req.DeviceInfo.TestDate != nil {
		report.DeviceInfo.TestDate = *req.DeviceInfo.TestDate
	}

	for _, r := range req.SimlockResults {
		report.SimlockResults = append(report.SimlockResults, models.MultibandaReportSimlockResult{
			TestID:  strings.TrimSpace(r.TestID),
			Carrier: strings.TrimSpace(r.Carrier),
			Result:  strings.TrimSpace(r.Result),
			Comment: strings.TrimSpace(r.Comment),
		})
	}
	for _, r := range req.BandResults {
		report.BandResults = append(report.BandResults, models.MultibandaReportBandResult{
			Technology: strings.TrimSpace(r.Technology),
			Band:       strings.TrimSpace(r.Band),
			Result:     strings.TrimSpace(r.Result),
			Comment:    strings.TrimSpace(r.Comment),
		})
	}
	for _, r := range req.SAEResults {
		report.SAEResults = append(report.SAEResults, models.MultibandaReportSAEResult{
			TestID:   strings.TrimSpace(r.TestID),
			Scenario: strings.TrimSpace(r.Scenario),
			Channel:  strings.TrimSpace(r.Channel),
			Operator: strings.TrimSpace(r.Operator),
			Result:   strings.TrimSpace(r.Result),
			Comment:  strings.TrimSpace(r.Comment),
		})
	}
	now := time.Now()
	for _, e := range req.Evidence {
		report.Evidence = append(report.Evidence, models.MultibandaReportEvidence{
			EvidenceType: strings.TrimSpace(e.EvidenceType),
			URL:          strings.TrimSpace(e.URL),
			Scenario:     strings.TrimSpace(e.Scenario),
			Operator:     strings.TrimSpace(e.Operator),
			UploadedAt:   now,
		})
	}

	return report
}

// MultibandaReportModelToSaved exposes the stored report back to the form.
func MultibandaReportModelToSaved(report *models.MultibandaReport) *responses.MultibandaReportSaved {
	if report == nil {
		return nil
	}

	saved := &responses.MultibandaReportSaved{
		DeviceInfo: responses.MultibandaReportDeviceInfo{
			CBSPackage:             report.DeviceInfo.CBSPackage,
			GooglePlaySystemUpdate: report.DeviceInfo.GooglePlaySystemUpdate,
			PreferredNetwork:       report.DeviceInfo.PreferredNetwork,
			FMRadio:                report.DeviceInfo.FMRadio,
			IMEI:                   report.DeviceInfo.IMEI,
			SerialNumber:           report.DeviceInfo.SerialNumber,
			StampCode:              report.DeviceInfo.StampCode,
		},
		SAEScenario:    report.SAEScenario,
		CarriersTested: report.CarriersTested,
		FMRadioResult:  report.FMRadioResult,
		FMRadioComment: report.FMRadioComment,
		SimlockResults: []responses.MultibandaReportSimlockResult{},
		BandResults:    []responses.MultibandaReportBandResult{},
		SAEResults:     []responses.MultibandaReportSAEResult{},
		Evidence:       []responses.MultibandaReportEvidence{},
	}
	if !report.DeviceInfo.TestDate.IsZero() {
		d := report.DeviceInfo.TestDate
		saved.DeviceInfo.TestDate = &d
	}

	for _, r := range report.SimlockResults {
		saved.SimlockResults = append(saved.SimlockResults, responses.MultibandaReportSimlockResult{
			TestID: r.TestID, Carrier: r.Carrier, Result: r.Result, Comment: r.Comment,
		})
	}
	for _, r := range report.BandResults {
		saved.BandResults = append(saved.BandResults, responses.MultibandaReportBandResult{
			Technology: r.Technology, Band: r.Band, Result: r.Result, Comment: r.Comment,
		})
	}
	for _, r := range report.SAEResults {
		saved.SAEResults = append(saved.SAEResults, responses.MultibandaReportSAEResult{
			TestID: r.TestID, Scenario: r.Scenario, Channel: r.Channel,
			Operator: r.Operator, Result: r.Result, Comment: r.Comment,
		})
	}
	for _, e := range report.Evidence {
		saved.Evidence = append(saved.Evidence, responses.MultibandaReportEvidence{
			EvidenceType: e.EvidenceType, URL: e.URL, Scenario: e.Scenario, Operator: e.Operator,
		})
	}

	return saved
}

// MultibandaReportModelToSaveRequest converts a stored report back into the
// save payload shape, so the completion summary can be computed from what is
// persisted using exactly the same code path as an incoming request.
// Returns nil for a report that does not exist yet.
func MultibandaReportModelToSaveRequest(report *models.MultibandaReport) *request.MultibandaReportSave {
	if report == nil {
		return nil
	}

	req := &request.MultibandaReportSave{
		DeviceInfo: request.MultibandaReportDeviceInfo{
			CBSPackage:             report.DeviceInfo.CBSPackage,
			GooglePlaySystemUpdate: report.DeviceInfo.GooglePlaySystemUpdate,
			PreferredNetwork:       report.DeviceInfo.PreferredNetwork,
			FMRadio:                report.DeviceInfo.FMRadio,
			IMEI:                   report.DeviceInfo.IMEI,
			SerialNumber:           report.DeviceInfo.SerialNumber,
			StampCode:              report.DeviceInfo.StampCode,
		},
		CarriersTested: report.CarriersTested,
		SAEScenario:    report.SAEScenario,
		FMRadioResult:  report.FMRadioResult,
		FMRadioComment: report.FMRadioComment,
	}
	if !report.DeviceInfo.TestDate.IsZero() {
		d := report.DeviceInfo.TestDate
		req.DeviceInfo.TestDate = &d
	}

	for _, r := range report.SimlockResults {
		req.SimlockResults = append(req.SimlockResults, request.MultibandaReportSimlockResult{
			TestID: r.TestID, Carrier: r.Carrier, Result: r.Result, Comment: r.Comment,
		})
	}
	for _, r := range report.BandResults {
		req.BandResults = append(req.BandResults, request.MultibandaReportBandResult{
			Technology: r.Technology, Band: r.Band, Result: r.Result, Comment: r.Comment,
		})
	}
	for _, r := range report.SAEResults {
		req.SAEResults = append(req.SAEResults, request.MultibandaReportSAEResult{
			TestID: r.TestID, Scenario: r.Scenario, Channel: r.Channel,
			Operator: r.Operator, Result: r.Result, Comment: r.Comment,
		})
	}
	for _, e := range report.Evidence {
		req.Evidence = append(req.Evidence, request.MultibandaReportEvidence{
			EvidenceType: e.EvidenceType, URL: e.URL,
			Scenario: e.Scenario, Operator: e.Operator,
		})
	}

	return req
}

// MultibandaReportCatalogs assembles the fixed catalogs the form renders from.
func MultibandaReportCatalogs() responses.MultibandaReportCatalogs {
	scenarios := []responses.MultibandaReportOption{
		{Value: enums.SAEScenarioLaboratory, Label: enums.SAEScenarioLabels[enums.SAEScenarioLaboratory]},
		{Value: enums.SAEScenarioRoom, Label: enums.SAEScenarioLabels[enums.SAEScenarioRoom]},
		{Value: enums.SAEScenarioBoth, Label: "Both"},
	}

	evidenceTypes := make([]responses.MultibandaReportOption, 0, len(enums.RequiredEvidenceTypes))
	for _, t := range enums.RequiredEvidenceTypes {
		evidenceTypes = append(evidenceTypes, responses.MultibandaReportOption{
			Value: t, Label: enums.EvidenceLabels[t],
		})
	}

	return responses.MultibandaReportCatalogs{
		PreferredNetworkOptions: enums.ReportPreferredNetworkOptions,
		FMRadioOptions:          enums.ReportFMRadioOptions,
		Stamps:                  enums.ReportStampCatalog,
		SimlockTests:            enums.ReportSimlockTests,
		SimlockCarriers:         enums.ReportSimlockCarriers,
		Technologies:            enums.ReportBandCatalog,
		SAETests:                enums.ReportSAETests,
		SAEOperators:            enums.SAEOperators,
		SAEChannels:             enums.SAEChannels,
		SAEScenarios:            scenarios,
		EvidenceTypes:           evidenceTypes,
		BandResultOptions: []string{
			enums.ReportResultOK, enums.ReportResultNoOK, enums.ReportResultNotSupported,
		},
		PassFailOptions:  []string{enums.ReportResultPass, enums.ReportResultFail},
		SAEResultOptions: []string{enums.ReportResultPass, enums.ReportResultFail, enums.ReportResultNA},
	}
}

// trimAll normalizes a list of free-form strings coming from the frontend.
func trimAll(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	return out
}
