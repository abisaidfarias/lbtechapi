package utils_test

import (
	"strings"
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func initialScope() utils.MultibandaReportScope {
	return utils.MultibandaReportScope{
		IncludesSimlock:   true,
		IncludesMultiband: true,
		SupportsGSM:       true,
		SupportsUMTS:      true,
		SupportsLTE:       true,
		Supports5G:        true,
	}
}

// completeReport builds a payload that satisfies every mandatory rule, so each
// test can knock out exactly one thing and assert on that failure alone.
func completeReport() *request.MultibandaReportSave {
	testDate := time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)
	req := &request.MultibandaReportSave{
		DeviceInfo: request.MultibandaReportDeviceInfo{
			CBSPackage:             "360913060",
			GooglePlaySystemUpdate: "01-may-2026",
			PreferredNetwork:       "5G/4G/3G/2G",
			FMRadio:                enums.ReportFMRadioNotSupported,
			TestDate:               &testDate,
			IMEI:                   "862542070112758",
			SerialNumber:           "P16BB/W5UT00079",
			StampCode:              "all_bands",
		},
		SAEScenario:    enums.SAEScenarioBoth,
		CarriersTested: allCarrierNames(),
	}

	for _, test := range enums.ReportSimlockTests {
		for _, carrier := range enums.ReportSimlockCarriers {
			req.SimlockResults = append(req.SimlockResults, request.MultibandaReportSimlockResult{
				TestID: test.ID, Carrier: carrier.Name, Result: enums.ReportResultPass,
			})
		}
	}
	for _, tech := range enums.ReportBandCatalog {
		for _, band := range tech.Bands {
			req.BandResults = append(req.BandResults, request.MultibandaReportBandResult{
				Technology: tech.Code, Band: band, Result: enums.ReportResultOK,
			})
		}
	}
	for _, scenario := range enums.SAEScenariosFor(enums.SAEScenarioBoth) {
		for _, channel := range enums.SAEChannels {
			for _, test := range enums.ReportSAETests {
				for _, operator := range enums.SAEOperators {
					req.SAEResults = append(req.SAEResults, request.MultibandaReportSAEResult{
						TestID: test.ID, Scenario: scenario, Channel: channel,
						Operator: operator, Result: enums.ReportResultPass,
					})
				}
			}
		}
	}
	req.Evidence = []request.MultibandaReportEvidence{
		{EvidenceType: enums.EvidenceSWVersion, URL: "https://example.com/sw.png"},
		{EvidenceType: enums.SAEEvidencePopUp, URL: "https://example.com/popup.png",
			Scenario: enums.SAEScenarioLaboratory, Operator: "Entel"},
		{EvidenceType: enums.SAEEvidenceCBHistory, URL: "https://example.com/cb.png",
			Scenario: enums.SAEScenarioRoom, Operator: "Movistar"},
	}
	return req
}

func allCarrierNames() []string {
	names := make([]string, 0, len(enums.ReportSimlockCarriers))
	for _, c := range enums.ReportSimlockCarriers {
		names = append(names, c.Name)
	}
	return names
}

func TestCompleteReportPassesGenerationValidation(t *testing.T) {
	if err := utils.ValidateMultibandaReportForGeneration(completeReport(), initialScope()); err != nil {
		t.Fatalf("expected complete report to pass, got: %v", err)
	}
}

// AC-03: a draft may be incomplete.
func TestDraftAllowsIncompleteReport(t *testing.T) {
	req := &request.MultibandaReportSave{
		DeviceInfo: request.MultibandaReportDeviceInfo{IMEI: "862542070112758"},
	}
	if err := utils.ValidateMultibandaReportDraft(req, initialScope()); err != nil {
		t.Fatalf("draft should accept incomplete data, got: %v", err)
	}
}

func TestDraftRejectsInvalidEnumValues(t *testing.T) {
	req := &request.MultibandaReportSave{
		DeviceInfo: request.MultibandaReportDeviceInfo{PreferredNetwork: "6G/5G"},
	}
	if err := utils.ValidateMultibandaReportDraft(req, initialScope()); err == nil {
		t.Fatal("expected invalid preferred_network to be rejected even in a draft")
	}
}

// AC-07: exactly one IMEI.
func TestGenerationRejectsMultipleIMEIs(t *testing.T) {
	req := completeReport()
	req.DeviceInfo.IMEI = "862542070112758, 862542070112759"
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "one IMEI") {
		t.Fatalf("expected single-IMEI rule, got: %v", err)
	}
}

// AC-09: FM Radio Supported enables the test and FAIL requires a comment.
func TestFMRadioSupportedRequiresResult(t *testing.T) {
	req := completeReport()
	req.DeviceInfo.FMRadio = enums.ReportFMRadioSupported
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "fm_radio_result") {
		t.Fatalf("expected fm_radio_result to be required, got: %v", err)
	}
}

func TestFMRadioFailRequiresComment(t *testing.T) {
	req := completeReport()
	req.DeviceInfo.FMRadio = enums.ReportFMRadioSupported
	req.FMRadioResult = enums.ReportResultFail
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "comment") {
		t.Fatalf("expected a comment to be required on FM Radio FAIL, got: %v", err)
	}
}

func TestFMRadioNotSupportedSkipsTest(t *testing.T) {
	req := completeReport()
	req.DeviceInfo.FMRadio = enums.ReportFMRadioNotSupported
	req.FMRadioResult = ""
	if err := utils.ValidateMultibandaReportForGeneration(req, initialScope()); err != nil {
		t.Fatalf("Not Supported must not require the conditional test, got: %v", err)
	}
}

// AC-10: all 14 SIMLOCK results, each FAIL commented.
func TestSimlockRequiresAllFourteenResults(t *testing.T) {
	req := completeReport()
	req.SimlockResults = req.SimlockResults[:13]
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "SIMLOCK") {
		t.Fatalf("expected the missing SIMLOCK combination to be caught, got: %v", err)
	}
}

func TestSimlockFailRequiresCarrierSpecificComment(t *testing.T) {
	req := completeReport()
	req.SimlockResults[2].Result = enums.ReportResultFail
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "comment") {
		t.Fatalf("expected a comment to be required for a failed carrier, got: %v", err)
	}

	req.SimlockResults[2].Comment = "Carrier not detected."
	if err := utils.ValidateMultibandaReportForGeneration(req, initialScope()); err != nil {
		t.Fatalf("commented failure should pass, got: %v", err)
	}
}

// AC-12: every band needs a result and NO OK needs a comment.
func TestBandNoOKRequiresComment(t *testing.T) {
	req := completeReport()
	req.BandResults[0].Result = enums.ReportResultNoOK
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "comment") {
		t.Fatalf("expected NO OK to require a comment, got: %v", err)
	}
}

func TestBandNotSupportedNeedsNoComment(t *testing.T) {
	req := completeReport()
	req.BandResults[0].Result = enums.ReportResultNotSupported
	if err := utils.ValidateMultibandaReportForGeneration(req, initialScope()); err != nil {
		t.Fatalf("NOT SUPPORTED must not require a comment, got: %v", err)
	}
}

// AC-02: SMR/MR/OS Upgrade skip SIMLOCK and Multi-band entirely.
func TestNonInitialScopeSkipsSimlockAndMultiband(t *testing.T) {
	req := completeReport()
	req.SimlockResults = nil
	req.BandResults = nil

	scope := utils.MultibandaReportScope{IncludesSimlock: false, IncludesMultiband: false}
	if err := utils.ValidateMultibandaReportForGeneration(req, scope); err != nil {
		t.Fatalf("SMR/MR/OS Upgrade should not require Initial-only blocks, got: %v", err)
	}
}

// AC-14/AC-15: both scenarios, both channels, all four operators.
func TestSAERequiresEveryScenarioChannelOperatorCombination(t *testing.T) {
	req := completeReport()
	req.SAEResults = req.SAEResults[:len(req.SAEResults)-1]
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "SAE") {
		t.Fatalf("expected the missing SAE combination to be caught, got: %v", err)
	}
}

func TestSAEScenarioIsRequired(t *testing.T) {
	req := completeReport()
	req.SAEScenario = ""
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "sae_scenario") {
		t.Fatalf("expected sae_scenario to be required, got: %v", err)
	}
}

func TestSingleScenarioDoesNotRequireTheOther(t *testing.T) {
	req := completeReport()
	req.SAEScenario = enums.SAEScenarioLaboratory
	// Keep only Laboratory results and point both evidences at it.
	filtered := req.SAEResults[:0]
	for _, r := range req.SAEResults {
		if r.Scenario == enums.SAEScenarioLaboratory {
			filtered = append(filtered, r)
		}
	}
	req.SAEResults = filtered
	req.Evidence[2].Scenario = enums.SAEScenarioLaboratory

	if err := utils.ValidateMultibandaReportForGeneration(req, initialScope()); err != nil {
		t.Fatalf("a single-scenario report should pass, got: %v", err)
	}
}

// AC-16: SAE FAIL requires a comment.
func TestSAEFailRequiresComment(t *testing.T) {
	req := completeReport()
	req.SAEResults[0].Result = enums.ReportResultFail
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "comment") {
		t.Fatalf("expected a comment to be required on SAE FAIL, got: %v", err)
	}
}

// AC-17: exactly two evidence images, each with scenario and operator.
func TestSAEEvidenceRequiresBothImages(t *testing.T) {
	req := completeReport()
	req.Evidence = req.Evidence[:2]
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "CB History") {
		t.Fatalf("expected the missing CB History evidence to be caught, got: %v", err)
	}
}

func TestSAEEvidenceRequiresValidOperator(t *testing.T) {
	req := completeReport()
	req.Evidence[1].Operator = "Nextel"
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "operator") {
		t.Fatalf("expected an invalid operator to be rejected, got: %v", err)
	}
}

func TestSAEEvidenceRequiresApplicableScenario(t *testing.T) {
	req := completeReport()
	req.SAEScenario = enums.SAEScenarioLaboratory
	filtered := req.SAEResults[:0]
	for _, r := range req.SAEResults {
		if r.Scenario == enums.SAEScenarioLaboratory {
			filtered = append(filtered, r)
		}
	}
	req.SAEResults = filtered
	// Evidence still points at SAE Room, which was not executed.
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "scenario") {
		t.Fatalf("expected evidence to require an applicable scenario, got: %v", err)
	}
}

func TestGenerationRequiresMandatoryDeviceFields(t *testing.T) {
	for _, field := range []struct {
		name   string
		mutate func(*request.MultibandaReportSave)
	}{
		{"cbs_package", func(r *request.MultibandaReportSave) { r.DeviceInfo.CBSPackage = "" }},
		{"google_play_system_update", func(r *request.MultibandaReportSave) { r.DeviceInfo.GooglePlaySystemUpdate = "" }},
		{"preferred_network", func(r *request.MultibandaReportSave) { r.DeviceInfo.PreferredNetwork = "" }},
		{"serial_number", func(r *request.MultibandaReportSave) { r.DeviceInfo.SerialNumber = "" }},
		{"stamp_code", func(r *request.MultibandaReportSave) { r.DeviceInfo.StampCode = "" }},
		{"test_date", func(r *request.MultibandaReportSave) { r.DeviceInfo.TestDate = nil }},
	} {
		t.Run(field.name, func(t *testing.T) {
			req := completeReport()
			field.mutate(req)
			err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
			if err == nil || !strings.Contains(err.Error(), field.name) {
				t.Fatalf("expected %s to be required, got: %v", field.name, err)
			}
		})
	}
}

// AC-08: the stamp must be one of the five supplied images.
func TestStampMustComeFromCatalog(t *testing.T) {
	req := completeReport()
	req.DeviceInfo.StampCode = "made_up_stamp"
	if err := utils.ValidateMultibandaReportForGeneration(req, initialScope()); err == nil {
		t.Fatal("expected an unknown stamp code to be rejected")
	}
	if len(enums.ReportStampCatalog) != 5 {
		t.Fatalf("expected exactly five stamps, got %d", len(enums.ReportStampCatalog))
	}
}

// The engineer only reports SIMLOCK for the carriers actually tested, not the
// full catalog of seven.
func TestSimlockOnlyRequiresTestedCarriers(t *testing.T) {
	req := completeReport()
	req.CarriersTested = []string{"Entel", "Movistar"}

	kept := req.SimlockResults[:0]
	for _, r := range req.SimlockResults {
		if r.Carrier == "Entel" || r.Carrier == "Movistar" {
			kept = append(kept, r)
		}
	}
	req.SimlockResults = kept

	if err := utils.ValidateMultibandaReportForGeneration(req, initialScope()); err != nil {
		t.Fatalf("only tested carriers should be required, got: %v", err)
	}
}

func TestAtLeastOneCarrierMustBeTested(t *testing.T) {
	req := completeReport()
	req.CarriersTested = nil
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "carrier") {
		t.Fatalf("expected at least one tested carrier to be required, got: %v", err)
	}
}

func TestUnknownCarrierIsRejected(t *testing.T) {
	req := completeReport()
	req.CarriersTested = append(req.CarriersTested, "Nextel")
	if err := utils.ValidateMultibandaReportDraft(req, initialScope()); err == nil {
		t.Fatal("expected an unknown carrier to be rejected")
	}
}

// The three screenshots are all mandatory; SW Version carries no scenario.
func TestSWVersionEvidenceIsRequired(t *testing.T) {
	req := completeReport()
	req.Evidence = req.Evidence[1:]
	err := utils.ValidateMultibandaReportForGeneration(req, initialScope())
	if err == nil || !strings.Contains(err.Error(), "SW Version") {
		t.Fatalf("expected SW Version evidence to be required, got: %v", err)
	}
}

func TestSWVersionEvidenceNeedsNoScenarioOrOperator(t *testing.T) {
	req := completeReport()
	req.Evidence[0].Scenario = ""
	req.Evidence[0].Operator = ""
	if err := utils.ValidateMultibandaReportForGeneration(req, initialScope()); err != nil {
		t.Fatalf("SW Version must not require scenario/operator, got: %v", err)
	}
}
