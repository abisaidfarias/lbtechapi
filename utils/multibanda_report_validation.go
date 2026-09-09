package utils

import (
	"fmt"
	"math"
	"strings"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// Completion summary for the report screen.
//
// This lives next to the validators on purpose: "what counts as filled in" and
// "what is mandatory" are the same rules the generation validator enforces, and
// splitting them across packages is how the two drift apart.
//
// Mandatory comments on FAIL / NO OK are deliberately kept out of the field
// count and reported separately as CommentsMissing.

// Section keys are part of the API contract: the frontend scrolls to a section
// by key, so renaming one silently breaks navigation.
const (
	reportSectionDeviceInfo = "device_info"
	reportSectionStamp      = "stamp"
	reportSectionCarriers   = "carriers"
	reportSectionEvidence   = "evidence"
	reportSectionFMRadio    = "fm_radio"
	reportSectionSimlock    = "simlock"
	reportSectionMultiband  = "multiband"
	reportSectionSAE        = "sae"
)

const (
	reportStatusComplete    = "complete"
	reportStatusIncomplete  = "incomplete"
	reportStatusConditional = "conditional"
)

// counter accumulates filled/total for one section.
type counter struct{ filled, total int }

func (c *counter) add(total int, filled bool) {
	c.total += total
	if filled {
		c.filled += total
	}
}

func (c *counter) addFilled(total, filled int) {
	c.total += total
	c.filled += filled
}

// BuildMultibandaReportValidation computes the completion summary for a saved
// report. A nil request (nothing saved yet) yields a zero-progress summary
// listing every blocker, rather than an empty object.
func BuildMultibandaReportValidation(
	req *request.MultibandaReportSave,
	scope MultibandaReportScope,
) responses.MultibandaReportValidation {
	if req == nil {
		req = &request.MultibandaReportSave{}
	}

	v := responses.MultibandaReportValidation{
		Sections:   []responses.MultibandaReportSection{},
		Blockers:   []responses.MultibandaReportBlocker{},
		Exceptions: []responses.MultibandaReportException{},
	}

	totalFields, filledFields := 0, 0
	addSection := func(s responses.MultibandaReportSection, c counter) {
		totalFields += c.total
		filledFields += c.filled
		v.Sections = append(v.Sections, s)
	}
	addBlocker := func(section, label string, count int) {
		if count > 0 {
			v.Blockers = append(v.Blockers, responses.MultibandaReportBlocker{
				Section: section, Label: label, Count: count,
			})
		}
	}

	// ---- Device information ----
	di := req.DeviceInfo
	deviceFields := []struct{ label, value string }{
		{"CBS Package", di.CBSPackage},
		{"Google Play System Update", di.GooglePlaySystemUpdate},
		{"Preferred Network", di.PreferredNetwork},
		{"FM Radio", di.FMRadio},
		{"IMEI", di.IMEI},
		{"S/N", di.SerialNumber},
	}
	var deviceCount counter
	for _, f := range deviceFields {
		filled := strings.TrimSpace(f.value) != ""
		deviceCount.add(1, filled)
		if !filled {
			addBlocker(reportSectionDeviceInfo, "Device Information: "+f.label, 1)
		}
	}
	hasTestDate := di.TestDate != nil && !di.TestDate.IsZero()
	deviceCount.add(1, hasTestDate)
	if !hasTestDate {
		addBlocker(reportSectionDeviceInfo, "Device Information: Test Date", 1)
	}
	addSection(responses.MultibandaReportSection{
		Key:       reportSectionDeviceInfo,
		Label:     "Device Information",
		Completed: intPtr(deviceCount.filled),
		Total:     intPtr(deviceCount.total),
		Status:    statusFor(deviceCount),
	}, deviceCount)

	// ---- Stamp ----
	var stampCount counter
	stampChosen := strings.TrimSpace(di.StampCode) != ""
	stampCount.add(1, stampChosen)
	stampDetail := "Not selected"
	if stamp, ok := enums.StampByCode(di.StampCode); ok {
		stampDetail = stamp.Label
	} else if !stampChosen {
		addBlocker(reportSectionStamp, "Stamp Type not selected", 1)
	}
	addSection(responses.MultibandaReportSection{
		Key:    reportSectionStamp,
		Label:  "Stamp Type",
		Status: statusFor(stampCount),
		Detail: stampDetail,
	}, stampCount)

	// ---- Carriers tested ----
	// Counts as a single field: the requirement is "at least one", not seven.
	carriers := trimmedNonEmpty(req.CarriersTested)
	var carrierCount counter
	carrierCount.add(1, len(carriers) > 0)
	if len(carriers) == 0 {
		addBlocker(reportSectionCarriers, "No carriers selected", 1)
	}
	addSection(responses.MultibandaReportSection{
		Key:    reportSectionCarriers,
		Label:  "Carriers Tested",
		Status: statusFor(carrierCount),
		Detail: fmt.Sprintf("%d of %d selected", len(carriers), len(enums.ReportSimlockCarriers)),
	}, carrierCount)

	// ---- Evidence: three images plus scenario/operator on the two SAE ones ----
	evidenceByType := map[string]request.MultibandaReportEvidence{}
	for _, e := range req.Evidence {
		evidenceByType[strings.TrimSpace(e.EvidenceType)] = e
	}
	var evidenceCount counter
	for _, evidenceType := range enums.RequiredEvidenceTypes {
		e := evidenceByType[evidenceType]
		label := enums.EvidenceLabels[evidenceType]

		hasImage := strings.TrimSpace(e.URL) != ""
		evidenceCount.add(1, hasImage)
		if !hasImage {
			addBlocker(reportSectionEvidence, "Photos: "+label+" image", 1)
		}

		if !enums.EvidenceRequiresScenarioAndOperator(evidenceType) {
			continue
		}
		hasScenario := strings.TrimSpace(e.Scenario) != ""
		hasOperator := strings.TrimSpace(e.Operator) != ""
		evidenceCount.add(1, hasScenario)
		evidenceCount.add(1, hasOperator)
		if !hasScenario {
			addBlocker(reportSectionEvidence, "Photos: "+label+" scenario", 1)
		}
		if !hasOperator {
			addBlocker(reportSectionEvidence, "Photos: "+label+" operator", 1)
		}
	}
	addSection(responses.MultibandaReportSection{
		Key:       reportSectionEvidence,
		Label:     "Evidence Photos",
		Completed: intPtr(evidenceCount.filled),
		Total:     intPtr(evidenceCount.total),
		Status:    statusFor(evidenceCount),
		Detail:    "3 uploads + metadata",
	}, evidenceCount)

	// ---- FM Radio: only counted when the device declares support ----
	var fmCount counter
	fmApplies := strings.TrimSpace(di.FMRadio) == enums.ReportFMRadioSupported
	fmStatus := reportStatusConditional
	fmDetail := "Conditional section"
	if fmApplies {
		hasResult := strings.TrimSpace(req.FMRadioResult) != ""
		fmCount.add(1, hasResult)
		fmStatus = statusFor(fmCount)
		fmDetail = "Non-removable application test"
		if !hasResult {
			addBlocker(reportSectionFMRadio, "FM Radio test result not set", 1)
		}
	}
	addSection(responses.MultibandaReportSection{
		Key:    reportSectionFMRadio,
		Label:  "FM Radio",
		Status: fmStatus,
		Detail: fmDetail,
	}, fmCount)

	// ---- SIMLOCK: two tests against every tested carrier ----
	if scope.IncludesSimlock {
		simlockIndex := map[string]request.MultibandaReportSimlockResult{}
		for _, r := range req.SimlockResults {
			simlockIndex[simlockKey(r.TestID, r.Carrier)] = r
		}
		var simlockCount counter
		empty := 0
		for _, test := range enums.ReportSimlockTests {
			for _, carrier := range carriers {
				r := simlockIndex[simlockKey(test.ID, carrier)]
				filled := strings.TrimSpace(r.Result) != ""
				simlockCount.add(1, filled)
				if !filled {
					empty++
				}
			}
		}
		addBlocker(reportSectionSimlock, "SIMLOCK cells empty", empty)
		addSection(responses.MultibandaReportSection{
			Key:       reportSectionSimlock,
			Label:     "SIMLOCK",
			Completed: intPtr(simlockCount.filled),
			Total:     intPtr(simlockCount.total),
			Status:    statusFor(simlockCount),
			Detail: fmt.Sprintf("%d × %d matrix",
				len(enums.ReportSimlockTests), len(carriers)),
		}, simlockCount)
	}

	// ---- Multi-band: every band in the catalog ----
	if scope.IncludesMultiband {
		bandIndex := map[string]request.MultibandaReportBandResult{}
		for _, r := range req.BandResults {
			bandIndex[bandKey(r.Technology, r.Band)] = r
		}
		var bandCount counter
		empty := 0
		for _, tech := range enums.ReportBandCatalog {
			for _, band := range tech.Bands {
				r := bandIndex[bandKey(tech.Code, band)]
				// NOT SUPPORTED is a real answer, so a band defaulted to it
				// already counts as filled.
				filled := strings.TrimSpace(r.Result) != ""
				bandCount.add(1, filled)
				if !filled {
					empty++
				}
			}
		}
		addBlocker(reportSectionMultiband, "Multi-band results missing", empty)
		addSection(responses.MultibandaReportSection{
			Key:       reportSectionMultiband,
			Label:     "Multi-band",
			Completed: intPtr(bandCount.filled),
			Total:     intPtr(bandCount.total),
			Status:    statusFor(bandCount),
			Detail: fmt.Sprintf("%d bands · %d technologies",
				bandCount.total, len(enums.ReportBandCatalog)),
		}, bandCount)
	}

	// ---- SAE: tests x operators x channels x scenarios, minus auto N/A ----
	saeCount, autoNA, saeEmpty := countSAE(req, scope)
	scenarios := enums.SAEScenariosFor(strings.TrimSpace(req.SAEScenario))
	matrices := len(scenarios) * len(enums.SAEChannels)
	addBlocker(reportSectionSAE, "SAE cells empty", saeEmpty)
	saeDetail := "Scenario not selected"
	if matrices > 0 {
		saeDetail = fmt.Sprintf("%d matrices · %d auto N/A", matrices, autoNA)
	} else {
		addBlocker(reportSectionSAE, "SAE scenario not selected", 1)
	}
	addSection(responses.MultibandaReportSection{
		Key:       reportSectionSAE,
		Label:     "SAE",
		Completed: intPtr(saeCount.filled),
		Total:     intPtr(saeCount.total),
		Status:    statusFor(saeCount),
		Detail:    saeDetail,
	}, saeCount)

	// ---- Exceptions: every FAIL / NO OK, wherever it lives ----
	v.Exceptions = collectReportExceptions(req, scope, carriers)
	for _, e := range v.Exceptions {
		if e.MissingComment {
			v.CommentsMissing++
		}
	}
	addBlocker(reportSectionSAE, "Required comments missing on FAIL / NO OK", v.CommentsMissing)

	v.RequiredFieldsLeft = totalFields - filledFields
	if totalFields > 0 {
		v.CompletionPercentage = int(math.Round(float64(filledFields) / float64(totalFields) * 100))
	}
	return v
}

// countSAE returns the SAE counter plus how many cells were excluded as auto
// N/A and how many remain empty. Rows whose technology the device does not
// support are not fillable, so they leave the denominator entirely.
func countSAE(req *request.MultibandaReportSave, scope MultibandaReportScope) (counter, int, int) {
	var c counter
	autoNA, empty := 0, 0

	scenarios := enums.SAEScenariosFor(strings.TrimSpace(req.SAEScenario))
	if len(scenarios) == 0 {
		return c, 0, 0
	}

	index := map[string]request.MultibandaReportSAEResult{}
	for _, r := range req.SAEResults {
		index[saeKey(r.TestID, r.Scenario, r.Channel, r.Operator)] = r
	}

	for _, scenario := range scenarios {
		for _, channel := range enums.SAEChannels {
			for _, test := range enums.ReportSAETests {
				if !saeTestApplies(test, scope) {
					autoNA += len(enums.SAEOperators)
					continue
				}
				for _, operator := range enums.SAEOperators {
					r := index[saeKey(test.ID, scenario, channel, operator)]
					filled := strings.TrimSpace(r.Result) != ""
					c.add(1, filled)
					if !filled {
						empty++
					}
				}
			}
		}
	}
	return c, autoNA, empty
}

// saeTestApplies reports whether a technology test is fillable for this device.
func saeTestApplies(test enums.SAETest, scope MultibandaReportScope) bool {
	switch test.RequiresTechnology {
	case enums.SAETechGSM:
		return scope.SupportsGSM
	case enums.SAETechUMTS:
		return scope.SupportsUMTS
	case enums.SAETechLTE:
		return scope.SupportsLTE
	case enums.SAETechNR:
		return scope.Supports5G
	default:
		return true
	}
}

// collectReportExceptions lists every failing result across all blocks so a
// missing comment can never hide inside a matrix.
func collectReportExceptions(
	req *request.MultibandaReportSave,
	scope MultibandaReportScope,
	carriers []string,
) []responses.MultibandaReportException {
	exceptions := []responses.MultibandaReportException{}

	if scope.IncludesMultiband {
		labels := map[string]string{}
		for _, tech := range enums.ReportBandCatalog {
			labels[tech.Code] = tech.Label
		}
		for _, r := range req.BandResults {
			if !enums.ResultRequiresComment(r.Result) {
				continue
			}
			exceptions = append(exceptions, responses.MultibandaReportException{
				Result:         r.Result,
				Label:          fmt.Sprintf("%s %s", labels[r.Technology], r.Band),
				Context:        "Multi-band",
				Comment:        r.Comment,
				MissingComment: strings.TrimSpace(r.Comment) == "",
				Ref: responses.MultibandaReportExceptionRef{
					Block: reportSectionMultiband, Technology: r.Technology, Band: r.Band,
				},
			})
		}
	}

	if scope.IncludesSimlock {
		names := map[string]string{}
		for _, t := range enums.ReportSimlockTests {
			names[t.ID] = t.Name
		}
		for _, r := range req.SimlockResults {
			if !enums.ResultRequiresComment(r.Result) || !containsString(carriers, r.Carrier) {
				continue
			}
			exceptions = append(exceptions, responses.MultibandaReportException{
				Result:         r.Result,
				Label:          fmt.Sprintf("%s — %s", names[r.TestID], r.Carrier),
				Context:        fmt.Sprintf("SIMLOCK · %s", r.TestID),
				Comment:        r.Comment,
				MissingComment: strings.TrimSpace(r.Comment) == "",
				Ref: responses.MultibandaReportExceptionRef{
					Block: reportSectionSimlock, TestID: r.TestID, Carrier: r.Carrier,
				},
			})
		}
	}

	saeNames := map[string]string{}
	for _, t := range enums.ReportSAETests {
		saeNames[t.ID] = t.Name
	}
	for _, r := range req.SAEResults {
		if !enums.ResultRequiresComment(r.Result) {
			continue
		}
		exceptions = append(exceptions, responses.MultibandaReportException{
			Result: r.Result,
			Label:  fmt.Sprintf("%s — %s", saeNames[r.TestID], r.Operator),
			Context: fmt.Sprintf("SAE · %s · Channel %s · %s",
				enums.SAEScenarioLabels[r.Scenario], r.Channel, r.TestID),
			Comment:        r.Comment,
			MissingComment: strings.TrimSpace(r.Comment) == "",
			Ref: responses.MultibandaReportExceptionRef{
				Block: reportSectionSAE, TestID: r.TestID, Scenario: r.Scenario,
				Channel: r.Channel, Operator: r.Operator,
			},
		})
	}

	if enums.ResultRequiresComment(strings.TrimSpace(req.FMRadioResult)) {
		exceptions = append(exceptions, responses.MultibandaReportException{
			Result:         req.FMRadioResult,
			Label:          enums.ReportFMRadioTestName,
			Context:        "FM Radio",
			Comment:        req.FMRadioComment,
			MissingComment: strings.TrimSpace(req.FMRadioComment) == "",
			Ref:            responses.MultibandaReportExceptionRef{Block: reportSectionFMRadio},
		})
	}

	return exceptions
}

func statusFor(c counter) string {
	if c.total == 0 {
		return reportStatusConditional
	}
	if c.filled >= c.total {
		return reportStatusComplete
	}
	return reportStatusIncomplete
}

func trimmedNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func intPtr(v int) *int { return &v }
