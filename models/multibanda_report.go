package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MultibandaReport is the Automatic Multi-Band Certification Report attached to
// a multibanda process. It replaces the manually uploaded test report: the
// engineer fills it in when Testing completes and the platform generates the
// PDF from it.
//
// Results are stored atomically (one document entry per test/carrier, per
// band, per test+scenario+channel+operator) so the report can be re-rendered,
// audited and partially edited without re-deriving anything from the PDF.
type MultibandaReport struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Multibanda primitive.ObjectID `bson:"multibanda"`

	// Status is draft while the engineer is still filling it in, and generated
	// once a PDF exists. A generated report can still be edited and regenerated.
	Status string `bson:"status"`

	// MultibandaType is snapshotted from the process so the report keeps the
	// scope it was filled under even if the process type is later corrected.
	MultibandaType string `bson:"multibanda_type"`

	DeviceInfo MultibandaReportDeviceInfo `bson:"device_info"`

	// CarriersTested are the carriers this device was actually evaluated
	// against. They drive which SIMLOCK combinations are required and are
	// printed in Device Information.
	CarriersTested []string `bson:"carriers_tested,omitempty"`

	// SimlockResults and BandResults only apply to Initial processes.
	SimlockResults []MultibandaReportSimlockResult `bson:"simlock_results,omitempty"`
	BandResults    []MultibandaReportBandResult    `bson:"band_results,omitempty"`

	SAEScenario string                      `bson:"sae_scenario,omitempty"`
	SAEResults  []MultibandaReportSAEResult `bson:"sae_results,omitempty"`

	// Evidence holds all three screenshots (SW Version, Pop-Up, CB History).
	Evidence []MultibandaReportEvidence `bson:"evidence,omitempty"`

	// FMRadioResult is only filled when DeviceInfo.FMRadio is "Supported".
	FMRadioResult  string `bson:"fm_radio_result,omitempty"`
	FMRadioComment string `bson:"fm_radio_comment,omitempty"`

	ReportURL   string    `bson:"report_url,omitempty"`
	GeneratedAt time.Time `bson:"generated_at,omitempty"`
	GeneratedBy string    `bson:"generated_by,omitempty"`

	CreatedDate time.Time `bson:"created_date"`
	UpdatedDate time.Time `bson:"updated_date"`
}

// MultibandaReportDeviceInfo holds the engineer-entered half of the Device
// Information block. The prefilled half (manufacturer, models, versions, SAR)
// is read from the device/multibanda at render time so it never goes stale.
type MultibandaReportDeviceInfo struct {
	CBSPackage             string    `bson:"cbs_package"`
	GooglePlaySystemUpdate string    `bson:"google_play_system_update"`
	PreferredNetwork       string    `bson:"preferred_network"`
	FMRadio                string    `bson:"fm_radio"`
	TestDate               time.Time `bson:"test_date"`
	IMEI                   string    `bson:"imei"`
	SerialNumber           string    `bson:"serial_number"`
	StampCode              string    `bson:"stamp_code"`
}

// MultibandaReportSimlockResult is one of the fourteen SIMLOCK combinations
// (two tests x seven carriers).
type MultibandaReportSimlockResult struct {
	TestID  string `bson:"test_id"`
	Carrier string `bson:"carrier"`
	Result  string `bson:"result"`
	Comment string `bson:"comment,omitempty"`
}

// MultibandaReportBandResult is the consolidated result for one band, covering
// Register, Data and Voice (including VoNR for 5G NR).
type MultibandaReportBandResult struct {
	Technology string `bson:"technology"`
	Band       string `bson:"band"`
	Result     string `bson:"result"`
	Comment    string `bson:"comment,omitempty"`
}

// MultibandaReportSAEResult is one SAE cell: a test evaluated for one
// scenario, channel and operator.
type MultibandaReportSAEResult struct {
	TestID   string `bson:"test_id"`
	Scenario string `bson:"scenario"`
	Channel  string `bson:"channel"`
	Operator string `bson:"operator"`
	Result   string `bson:"result"`
	Comment  string `bson:"comment,omitempty"`
}

// MultibandaReportEvidence is one uploaded screenshot. Scenario and Operator
// are only meaningful for the SAE ones (Pop-Up, CB History); SW Version leaves
// them empty.
type MultibandaReportEvidence struct {
	EvidenceType string    `bson:"evidence_type"`
	URL          string    `bson:"url"`
	Scenario     string    `bson:"scenario,omitempty"`
	Operator     string    `bson:"operator,omitempty"`
	UploadedAt   time.Time `bson:"uploaded_at"`
	UploadedBy   string    `bson:"uploaded_by,omitempty"`
}
