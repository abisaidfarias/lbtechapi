package utils_test

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

// scopeNoUMTS makes the UMTS SAE row auto N/A, which is what removes cells from
// the denominator.
func scopeNoUMTS() utils.MultibandaReportScope {
	return utils.MultibandaReportScope{
		IncludesSimlock: true, IncludesMultiband: true,
		SupportsGSM: true, SupportsUMTS: false, SupportsLTE: true, Supports5G: true,
	}
}

// buildPartialReport reproduces the agreed worked example:
//
//	7 carriers · sae_scenario "both" · 1 SAE row auto N/A
//	device_info 6/7 · stamp chosen · evidence 4/7 · FM Radio Supported with no
//	result · 4 SIMLOCK cells · all 17 bands · 40 SAE cells · 2 comments pending
//
// Expected: total 208, filled 73 → 35%, 135 left, 2 comments missing.
func buildPartialReport() *request.MultibandaReportSave {
	testDate := time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC)

	req := &request.MultibandaReportSave{
		DeviceInfo: request.MultibandaReportDeviceInfo{
			// CBSPackage intentionally blank → 6 of 7.
			GooglePlaySystemUpdate: "01-may-2026",
			PreferredNetwork:       "5G/4G/3G/2G",
			FMRadio:                enums.ReportFMRadioSupported,
			TestDate:               &testDate,
			IMEI:                   "862542070112758",
			SerialNumber:           "P16BB/W5UT00079",
			StampCode:              "all_bands",
		},
		CarriersTested: allCarrierNames(),
		SAEScenario:    enums.SAEScenarioBoth,
		// FMRadioResult intentionally blank → the conditional test is pending.
	}

	// Evidence 4 of 7: SW Version image + Pop-Up complete; CB History missing.
	req.Evidence = []request.MultibandaReportEvidence{
		{EvidenceType: enums.EvidenceSWVersion, URL: "https://x/sw.png"},
		{EvidenceType: enums.SAEEvidencePopUp, URL: "https://x/popup.png",
			Scenario: enums.SAEScenarioLaboratory, Operator: "Entel"},
	}

	// SIMLOCK: 4 of 14 filled, one of them a FAIL without comment.
	filledSimlock := 0
	for _, test := range enums.ReportSimlockTests {
		for _, carrier := range enums.ReportSimlockCarriers {
			if filledSimlock >= 4 {
				break
			}
			result := enums.ReportResultPass
			if filledSimlock == 3 {
				result = enums.ReportResultFail // pending comment #1
			}
			req.SimlockResults = append(req.SimlockResults, request.MultibandaReportSimlockResult{
				TestID: test.ID, Carrier: carrier.Name, Result: result,
			})
			filledSimlock++
		}
	}

	// All 17 bands answered; one NO OK without comment (pending comment #2).
	first := true
	for _, tech := range enums.ReportBandCatalog {
		for _, band := range tech.Bands {
			result := enums.ReportResultNotSupported
			if first {
				result = enums.ReportResultNoOK
				first = false
			}
			req.BandResults = append(req.BandResults, request.MultibandaReportBandResult{
				Technology: tech.Code, Band: band, Result: result,
			})
		}
	}

	// 40 SAE cells filled, skipping the auto-N/A UMTS row.
	filledSAE := 0
	for _, scenario := range enums.SAEScenariosFor(enums.SAEScenarioBoth) {
		for _, channel := range enums.SAEChannels {
			for _, test := range enums.ReportSAETests {
				if test.RequiresTechnology == enums.SAETechUMTS {
					continue
				}
				for _, operator := range enums.SAEOperators {
					if filledSAE >= 40 {
						break
					}
					req.SAEResults = append(req.SAEResults, request.MultibandaReportSAEResult{
						TestID: test.ID, Scenario: scenario, Channel: channel,
						Operator: operator, Result: enums.ReportResultPass,
					})
					filledSAE++
				}
			}
		}
	}

	return req
}

func TestValidationMatchesWorkedExample(t *testing.T) {
	v := utils.BuildMultibandaReportValidation(buildPartialReport(), scopeNoUMTS())

	if v.CompletionPercentage != 35 {
		t.Errorf("completion_percentage: got %d, want 35", v.CompletionPercentage)
	}
	if v.RequiredFieldsLeft != 135 {
		t.Errorf("required_fields_left: got %d, want 135", v.RequiredFieldsLeft)
	}
	if v.CommentsMissing != 2 {
		t.Errorf("comments_missing: got %d, want 2", v.CommentsMissing)
	}
}

func TestValidationSectionTotals(t *testing.T) {
	v := utils.BuildMultibandaReportValidation(buildPartialReport(), scopeNoUMTS())

	want := map[string]struct{ completed, total int }{
		"device_info": {6, 7},
		"evidence":    {4, 7},
		"simlock":     {4, 14},
		"multiband":   {17, 17},
		"sae":         {40, 160}, // 176 minus 16 auto N/A
	}

	for _, s := range v.Sections {
		exp, tracked := want[s.Key]
		if !tracked {
			continue
		}
		if s.Completed == nil || s.Total == nil {
			t.Errorf("section %s: expected counts, got nil", s.Key)
			continue
		}
		if *s.Completed != exp.completed || *s.Total != exp.total {
			t.Errorf("section %s: got %d/%d, want %d/%d",
				s.Key, *s.Completed, *s.Total, exp.completed, exp.total)
		}
	}
}

// The sidebar keys are an API contract: the frontend scrolls by key.
func TestValidationSectionKeysAndOrder(t *testing.T) {
	v := utils.BuildMultibandaReportValidation(buildPartialReport(), scopeNoUMTS())

	want := []string{"device_info", "stamp", "carriers", "evidence",
		"fm_radio", "simlock", "multiband", "sae"}
	if len(v.Sections) != len(want) {
		t.Fatalf("got %d sections, want %d", len(v.Sections), len(want))
	}
	for i, key := range want {
		if v.Sections[i].Key != key {
			t.Errorf("section %d: got key %q, want %q", i, v.Sections[i].Key, key)
		}
	}
}

func TestValidationExceptionRefs(t *testing.T) {
	v := utils.BuildMultibandaReportValidation(buildPartialReport(), scopeNoUMTS())

	if len(v.Exceptions) != 2 {
		t.Fatalf("got %d exceptions, want 2", len(v.Exceptions))
	}

	var band, simlock *struct {
		result, block string
	}
	for _, e := range v.Exceptions {
		got := &struct{ result, block string }{e.Result, e.Ref.Block}
		switch e.Ref.Block {
		case "multiband":
			band = got
			if e.Ref.Technology == "" || e.Ref.Band == "" {
				t.Error("multiband exception must carry technology and band")
			}
			// technology must be the code, not the label
			if e.Ref.Technology != "gsm" {
				t.Errorf("technology: got %q, want the catalog code", e.Ref.Technology)
			}
		case "simlock":
			simlock = got
			if e.Ref.TestID == "" || e.Ref.Carrier == "" {
				t.Error("simlock exception must carry test_id and carrier")
			}
		}
		if !e.MissingComment {
			t.Errorf("exception %s should be flagged as missing its comment", e.Label)
		}
	}
	if band == nil || simlock == nil {
		t.Fatal("expected one multiband and one simlock exception")
	}
}

// A brand-new report must still return the full picture at 0%.
func TestValidationOnEmptyReport(t *testing.T) {
	v := utils.BuildMultibandaReportValidation(nil, scopeNoUMTS())

	if v.CompletionPercentage != 0 {
		t.Errorf("completion_percentage: got %d, want 0", v.CompletionPercentage)
	}
	if len(v.Sections) == 0 {
		t.Error("sections must be present even with nothing saved")
	}
	if len(v.Blockers) == 0 {
		t.Error("blockers must list what is missing on an empty report")
	}
}

// Scope drives the denominator: non-Initial processes drop two whole sections.
func TestValidationExcludesNonApplicableSections(t *testing.T) {
	scope := utils.MultibandaReportScope{
		IncludesSimlock: false, IncludesMultiband: false,
		SupportsGSM: true, SupportsUMTS: true, SupportsLTE: true, Supports5G: true,
	}
	v := utils.BuildMultibandaReportValidation(buildPartialReport(), scope)

	for _, s := range v.Sections {
		if s.Key == "simlock" || s.Key == "multiband" {
			t.Errorf("section %s must not appear when out of scope", s.Key)
		}
	}
}

// Unselecting carriers shrinks the SIMLOCK denominator.
func TestValidationSimlockFollowsSelectedCarriers(t *testing.T) {
	req := buildPartialReport()
	req.CarriersTested = []string{"Entel", "Movistar"}

	v := utils.BuildMultibandaReportValidation(req, scopeNoUMTS())
	for _, s := range v.Sections {
		if s.Key != "simlock" {
			continue
		}
		if s.Total == nil || *s.Total != 4 {
			t.Fatalf("simlock total: got %v, want 4 (2 tests × 2 carriers)", s.Total)
		}
	}
}

// Switching from both to a single scenario halves the SAE denominator.
func TestValidationSAEFollowsScenarioSelection(t *testing.T) {
	req := buildPartialReport()
	req.SAEScenario = enums.SAEScenarioLaboratory

	v := utils.BuildMultibandaReportValidation(req, scopeNoUMTS())
	for _, s := range v.Sections {
		if s.Key != "sae" {
			continue
		}
		// 11 tests minus the auto-N/A UMTS row = 10, × 4 operators × 2 matrices
		if s.Total == nil || *s.Total != 80 {
			t.Fatalf("sae total: got %v, want 80", s.Total)
		}
	}
}
