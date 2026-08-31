package request

import "time"

// MultibandaReportSave is the payload for saving the automatic report, used
// both for draft saves and for the save that precedes PDF generation.
// Validation depth differs between the two: a draft accepts incomplete data,
// while generation enforces every mandatory field (see utils validators).
type MultibandaReportSave struct {
	DeviceInfo MultibandaReportDeviceInfo `json:"device_info"`

	// CarriersTested are the carriers the engineer checked. SIMLOCK results are
	// only required for these.
	CarriersTested []string `json:"carriers_tested"`

	SimlockResults []MultibandaReportSimlockResult `json:"simlock_results"`
	BandResults    []MultibandaReportBandResult    `json:"band_results"`

	SAEScenario string                      `json:"sae_scenario"`
	SAEResults  []MultibandaReportSAEResult `json:"sae_results"`

	// Evidence carries all three screenshots: sw_version, pop_up, cb_history.
	Evidence []MultibandaReportEvidence `json:"evidence"`

	FMRadioResult  string `json:"fm_radio_result"`
	FMRadioComment string `json:"fm_radio_comment"`
}

type MultibandaReportDeviceInfo struct {
	CBSPackage             string     `json:"cbs_package"`
	GooglePlaySystemUpdate string     `json:"google_play_system_update"`
	PreferredNetwork       string     `json:"preferred_network"`
	FMRadio                string     `json:"fm_radio"`
	TestDate               *time.Time `json:"test_date"`
	IMEI                   string     `json:"imei"`
	SerialNumber           string     `json:"serial_number"`
	StampCode              string     `json:"stamp_code"`
}

type MultibandaReportSimlockResult struct {
	TestID  string `json:"test_id"`
	Carrier string `json:"carrier"`
	Result  string `json:"result"`
	Comment string `json:"comment"`
}

type MultibandaReportBandResult struct {
	Technology string `json:"technology"`
	Band       string `json:"band"`
	Result     string `json:"result"`
	Comment    string `json:"comment"`
}

type MultibandaReportSAEResult struct {
	TestID   string `json:"test_id"`
	Scenario string `json:"scenario"`
	Channel  string `json:"channel"`
	Operator string `json:"operator"`
	Result   string `json:"result"`
	Comment  string `json:"comment"`
}

type MultibandaReportEvidence struct {
	EvidenceType string `json:"evidence_type"`
	URL          string `json:"url"`
	Scenario     string `json:"scenario"`
	Operator     string `json:"operator"`
}
