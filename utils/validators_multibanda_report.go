package utils

import (
	"fmt"
	"strings"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

// MultibandaReportScope describes which blocks apply to the process being
// reported, so validation only demands what the process type actually needs.
type MultibandaReportScope struct {
	IncludesSimlock   bool
	IncludesMultiband bool
	SupportsGSM       bool
	SupportsUMTS      bool
	SupportsLTE       bool
	Supports5G        bool
}

// ValidateMultibandaReportDraft applies the light checks that hold even for an
// incomplete report: whatever the engineer did fill in must be a legal value.
// It deliberately does not require completeness — that is generation's job.
func ValidateMultibandaReportDraft(req *request.MultibandaReportSave, scope MultibandaReportScope) error {
	if req == nil {
		return NewValidationError("report payload is required")
	}

	di := req.DeviceInfo
	if v := strings.TrimSpace(di.PreferredNetwork); v != "" && !enums.IsValidPreferredNetwork(v) {
		return NewValidationError("invalid preferred_network")
	}
	if v := strings.TrimSpace(di.FMRadio); v != "" && !enums.IsValidFMRadioOption(v) {
		return NewValidationError("invalid fm_radio")
	}
	if v := strings.TrimSpace(di.StampCode); v != "" && !enums.IsValidStampCode(v) {
		return NewValidationError("invalid stamp_code")
	}

	if v := strings.TrimSpace(req.SAEScenario); v != "" && !enums.IsValidSAEScenarioSelection(v) {
		return NewValidationError("invalid sae_scenario")
	}

	for _, r := range req.SimlockResults {
		if v := strings.TrimSpace(r.Result); v != "" && !enums.IsValidPassFailResult(v) {
			return NewValidationError(fmt.Sprintf("invalid simlock result for %s/%s", r.TestID, r.Carrier))
		}
	}
	for _, r := range req.BandResults {
		if v := strings.TrimSpace(r.Result); v != "" && !enums.IsValidBandResult(v) {
			return NewValidationError(fmt.Sprintf("invalid band result for %s %s", r.Technology, r.Band))
		}
	}
	for _, r := range req.SAEResults {
		if v := strings.TrimSpace(r.Result); v != "" && !enums.IsValidSAEResult(v) {
			return NewValidationError(fmt.Sprintf("invalid sae result for %s", r.TestID))
		}
	}
	for _, e := range req.Evidence {
		if v := strings.TrimSpace(e.EvidenceType); v != "" && !enums.IsValidEvidenceType(v) {
			return NewValidationError("invalid evidence type")
		}
	}
	for _, c := range req.CarriersTested {
		if !isKnownCarrier(strings.TrimSpace(c)) {
			return NewValidationError(fmt.Sprintf("unknown carrier: %s", c))
		}
	}

	if v := strings.TrimSpace(req.FMRadioResult); v != "" && !enums.IsValidPassFailResult(v) {
		return NewValidationError("invalid fm_radio_result")
	}

	return nil
}

// ValidateMultibandaReportForGeneration enforces the full mandatory set. The
// spec keeps Generate PDF disabled until everything applicable is complete, and
// asks for this to be checked server-side too — not just in the UI.
func ValidateMultibandaReportForGeneration(req *request.MultibandaReportSave, scope MultibandaReportScope) error {
	if err := ValidateMultibandaReportDraft(req, scope); err != nil {
		return err
	}

	if err := validateReportDeviceInfoComplete(req.DeviceInfo); err != nil {
		return err
	}
	if err := validateReportFMRadio(req); err != nil {
		return err
	}
	if scope.IncludesSimlock {
		if err := validateReportSimlockComplete(req.SimlockResults, req.CarriersTested); err != nil {
			return err
		}
	}
	if scope.IncludesMultiband {
		if err := validateReportBandsComplete(req.BandResults); err != nil {
			return err
		}
	}
	if err := validateReportSAEComplete(req, scope); err != nil {
		return err
	}
	// Evidence is validated after SAE so the scenario selection is known good.
	if err := validateReportEvidence(req.Evidence, enums.SAEScenariosFor(strings.TrimSpace(req.SAEScenario))); err != nil {
		return err
	}

	return nil
}

func validateReportDeviceInfoComplete(di request.MultibandaReportDeviceInfo) error {
	required := []struct {
		field string
		value string
	}{
		{"cbs_package", di.CBSPackage},
		{"google_play_system_update", di.GooglePlaySystemUpdate},
		{"preferred_network", di.PreferredNetwork},
		{"fm_radio", di.FMRadio},
		{"imei", di.IMEI},
		{"serial_number", di.SerialNumber},
		{"stamp_code", di.StampCode},
	}
	for _, f := range required {
		if strings.TrimSpace(f.value) == "" {
			return NewValidationError(fmt.Sprintf("missing required field: %s", f.field))
		}
	}
	if di.TestDate == nil || di.TestDate.IsZero() {
		return NewValidationError("missing required field: test_date")
	}
	// The spec allows exactly one IMEI, so reject anything that looks like a list.
	if strings.ContainsAny(di.IMEI, ",;\n ") {
		return NewValidationError("only one IMEI is allowed")
	}
	return nil
}

// validateReportFMRadio enforces the conditional test: the non-removable
// application result only exists when FM Radio is declared Supported, and a
// FAIL there needs a comment.
func validateReportFMRadio(req *request.MultibandaReportSave) error {
	declared := strings.TrimSpace(req.DeviceInfo.FMRadio)
	result := strings.TrimSpace(req.FMRadioResult)

	if declared != enums.ReportFMRadioSupported {
		return nil
	}
	if result == "" {
		return NewValidationError("fm_radio_result is required when FM Radio is Supported")
	}
	if enums.ResultRequiresComment(result) && strings.TrimSpace(req.FMRadioComment) == "" {
		return NewValidationError("a comment is required when the FM Radio test fails")
	}
	return nil
}

// validateReportSimlockComplete requires a result for each test against every
// carrier the engineer actually tested, rather than the full fixed catalog: a
// device is not necessarily evaluated against all seven carriers.
func validateReportSimlockComplete(
	results []request.MultibandaReportSimlockResult,
	carriersTested []string,
) error {
	if len(carriersTested) == 0 {
		return NewValidationError("at least one tested carrier is required")
	}

	seen := make(map[string]request.MultibandaReportSimlockResult, len(results))
	for _, r := range results {
		seen[simlockKey(r.TestID, r.Carrier)] = r
	}

	for _, test := range enums.ReportSimlockTests {
		for _, carrier := range carriersTested {
			carrier = strings.TrimSpace(carrier)
			r, ok := seen[simlockKey(test.ID, carrier)]
			if !ok || strings.TrimSpace(r.Result) == "" {
				return NewValidationError(fmt.Sprintf(
					"missing SIMLOCK result for %s / %s", test.Name, carrier))
			}
			if enums.ResultRequiresComment(r.Result) && strings.TrimSpace(r.Comment) == "" {
				return NewValidationError(fmt.Sprintf(
					"a comment is required for the failed SIMLOCK result %s / %s", test.Name, carrier))
			}
		}
	}
	return nil
}

func isKnownCarrier(name string) bool {
	for _, c := range enums.ReportSimlockCarriers {
		if c.Name == name {
			return true
		}
	}
	return false
}

func validateReportBandsComplete(results []request.MultibandaReportBandResult) error {
	seen := make(map[string]request.MultibandaReportBandResult, len(results))
	for _, r := range results {
		seen[bandKey(r.Technology, r.Band)] = r
	}

	for _, tech := range enums.ReportBandCatalog {
		for _, band := range tech.Bands {
			r, ok := seen[bandKey(tech.Code, band)]
			if !ok || strings.TrimSpace(r.Result) == "" {
				return NewValidationError(fmt.Sprintf(
					"missing Multi-band result for %s %s", tech.Label, band))
			}
			if enums.ResultRequiresComment(r.Result) && strings.TrimSpace(r.Comment) == "" {
				return NewValidationError(fmt.Sprintf(
					"a comment is required for the NO OK result on %s %s", tech.Label, band))
			}
		}
	}
	return nil
}

func validateReportSAEComplete(req *request.MultibandaReportSave, scope MultibandaReportScope) error {
	scenarios := enums.SAEScenariosFor(strings.TrimSpace(req.SAEScenario))
	if len(scenarios) == 0 {
		return NewValidationError("sae_scenario is required")
	}

	seen := make(map[string]request.MultibandaReportSAEResult, len(req.SAEResults))
	for _, r := range req.SAEResults {
		seen[saeKey(r.TestID, r.Scenario, r.Channel, r.Operator)] = r
	}

	for _, scenario := range scenarios {
		for _, channel := range enums.SAEChannels {
			for _, test := range enums.ReportSAETests {
				for _, operator := range enums.SAEOperators {
					r, ok := seen[saeKey(test.ID, scenario, channel, operator)]
					if !ok || strings.TrimSpace(r.Result) == "" {
						return NewValidationError(fmt.Sprintf(
							"missing SAE result for %s / %s / channel %s / %s",
							test.Name, enums.SAEScenarioLabels[scenario], channel, operator))
					}
					if enums.ResultRequiresComment(r.Result) && strings.TrimSpace(r.Comment) == "" {
						return NewValidationError(fmt.Sprintf(
							"a comment is required for the failed SAE result %s / %s / channel %s / %s",
							test.Name, enums.SAEScenarioLabels[scenario], channel, operator))
					}
				}
			}
		}
	}

	return nil
}

// validateReportEvidence requires the three screenshots. The two SAE ones must
// additionally name a scenario that was actually executed and a fixed operator;
// SW Version documents the build and carries no such metadata.
func validateReportEvidence(evidence []request.MultibandaReportEvidence, scenarios []string) error {
	byType := make(map[string]request.MultibandaReportEvidence, len(evidence))
	for _, e := range evidence {
		byType[strings.TrimSpace(e.EvidenceType)] = e
	}

	for _, required := range enums.RequiredEvidenceTypes {
		e, ok := byType[required]
		label := enums.EvidenceLabels[required]
		if !ok || strings.TrimSpace(e.URL) == "" {
			return NewValidationError(fmt.Sprintf("missing evidence: %s", label))
		}
		if !enums.EvidenceRequiresScenarioAndOperator(required) {
			continue
		}
		if !containsString(scenarios, strings.TrimSpace(e.Scenario)) {
			return NewValidationError(fmt.Sprintf("%s must reference an applicable scenario", label))
		}
		if !containsString(enums.SAEOperators, strings.TrimSpace(e.Operator)) {
			return NewValidationError(fmt.Sprintf("%s must reference a valid operator", label))
		}
	}
	return nil
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func simlockKey(testID, carrier string) string {
	return strings.TrimSpace(testID) + "|" + strings.TrimSpace(carrier)
}

func bandKey(technology, band string) string {
	return strings.TrimSpace(technology) + "|" + strings.TrimSpace(band)
}

func saeKey(testID, scenario, channel, operator string) string {
	return strings.TrimSpace(testID) + "|" + strings.TrimSpace(scenario) + "|" +
		strings.TrimSpace(channel) + "|" + strings.TrimSpace(operator)
}
