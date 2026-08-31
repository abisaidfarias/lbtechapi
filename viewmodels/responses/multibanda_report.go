package responses

import (
	"time"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MultibandaReportForm is what the report screen loads: the prefilled device
// data, whatever the engineer has saved so far, the applicable scope for this
// process type, and the catalogs needed to render the result matrices.
type MultibandaReportForm struct {
	MultibandaID   primitive.ObjectID `json:"multibanda_id"`
	MultibandaType string             `json:"multibanda_type"`
	Status         string             `json:"status"`

	// Scope flags let the frontend show or hide whole blocks without
	// re-deriving the process-type rules.
	IncludesSimlock   bool `json:"includes_simlock"`
	IncludesMultiband bool `json:"includes_multiband"`

	Prefilled MultibandaReportPrefilled `json:"prefilled"`
	Saved     *MultibandaReportSaved    `json:"saved,omitempty"`
	Catalogs  MultibandaReportCatalogs  `json:"catalogs"`

	// Validation drives the completion bar, the section sidebar, the blocker
	// panel and the exceptions list. It is computed server-side so the
	// frontend never has to re-derive what is mandatory.
	Validation MultibandaReportValidation `json:"validation"`

	ReportURL   string `json:"report_url,omitempty"`
	GeneratedAt string `json:"generated_at,omitempty"`
}

// MultibandaReportValidation summarises how much of the report is filled in
// and precisely what is still blocking PDF generation.
type MultibandaReportValidation struct {
	CompletionPercentage int `json:"completion_percentage"`
	RequiredFieldsLeft   int `json:"required_fields_left"`
	CommentsMissing      int `json:"comments_missing"`

	Sections   []MultibandaReportSection   `json:"sections"`
	Blockers   []MultibandaReportBlocker   `json:"blockers"`
	Exceptions []MultibandaReportException `json:"exceptions"`
}

// MultibandaReportSection is one row of the sidebar. Completed/Total are
// omitted for sections that are pass/fail rather than counted.
type MultibandaReportSection struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Completed *int   `json:"completed,omitempty"`
	Total     *int   `json:"total,omitempty"`
	Status    string `json:"status"` // complete | incomplete | conditional
	Detail    string `json:"detail,omitempty"`
}

type MultibandaReportBlocker struct {
	Section string `json:"section"`
	Label   string `json:"label"`
	Count   int    `json:"count"`
}

// MultibandaReportException is one FAIL or NO OK. Ref carries enough to let
// the frontend jump straight to the offending cell.
type MultibandaReportException struct {
	Result         string                       `json:"result"`
	Label          string                       `json:"label"`
	Context        string                       `json:"context"`
	Comment        string                       `json:"comment"`
	MissingComment bool                         `json:"missing_comment"`
	Ref            MultibandaReportExceptionRef `json:"ref"`
}

type MultibandaReportExceptionRef struct {
	Block      string `json:"block"` // multiband | simlock | sae | fm_radio
	Technology string `json:"technology,omitempty"`
	Band       string `json:"band,omitempty"`
	TestID     string `json:"test_id,omitempty"`
	Carrier    string `json:"carrier,omitempty"`
	Scenario   string `json:"scenario,omitempty"`
	Channel    string `json:"channel,omitempty"`
	Operator   string `json:"operator,omitempty"`
}

// MultibandaReportPrefilled is the read-only half of Device Information,
// resolved from the device and multibanda records.
type MultibandaReportPrefilled struct {
	Manufacturer        string `json:"manufacturer"`
	CommercialName      string `json:"commercial_name"`
	TechnicalModel      string `json:"technical_model"`
	HardwareVersion     string `json:"hardware_version"`
	SoftwareVersion     string `json:"software_version"`
	SARValue            string `json:"sar_value"`
	OperativeSystemType string `json:"operative_system_type"`

	// Device technology support drives which SAE technology tests are
	// applicable; unsupported ones are forced to N/A.
	SupportsGSM  bool `json:"supports_gsm"`
	SupportsUMTS bool `json:"supports_umts"`
	SupportsLTE  bool `json:"supports_lte"`
	Supports5G   bool `json:"supports_5g"`
}

type MultibandaReportSaved struct {
	DeviceInfo     MultibandaReportDeviceInfo      `json:"device_info"`
	CarriersTested []string                        `json:"carriers_tested"`
	SimlockResults []MultibandaReportSimlockResult `json:"simlock_results"`
	BandResults    []MultibandaReportBandResult    `json:"band_results"`
	SAEScenario    string                          `json:"sae_scenario"`
	SAEResults     []MultibandaReportSAEResult     `json:"sae_results"`
	Evidence       []MultibandaReportEvidence      `json:"evidence"`
	FMRadioResult  string                          `json:"fm_radio_result"`
	FMRadioComment string                          `json:"fm_radio_comment"`
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
	Comment string `json:"comment,omitempty"`
}

type MultibandaReportBandResult struct {
	Technology string `json:"technology"`
	Band       string `json:"band"`
	Result     string `json:"result"`
	Comment    string `json:"comment,omitempty"`
}

type MultibandaReportSAEResult struct {
	TestID   string `json:"test_id"`
	Scenario string `json:"scenario"`
	Channel  string `json:"channel"`
	Operator string `json:"operator"`
	Result   string `json:"result"`
	Comment  string `json:"comment,omitempty"`
}

type MultibandaReportEvidence struct {
	EvidenceType string `json:"evidence_type"`
	URL          string `json:"url"`
	Scenario     string `json:"scenario"`
	Operator     string `json:"operator"`
}

// MultibandaReportCatalogs carries the fixed catalogs so the frontend renders
// the matrices from server data instead of duplicating the lists.
type MultibandaReportCatalogs struct {
	PreferredNetworkOptions []string                 `json:"preferred_network_options"`
	FMRadioOptions          []string                 `json:"fm_radio_options"`
	Stamps                  []enums.ReportStamp      `json:"stamps"`
	SimlockTests            []enums.ReportTest       `json:"simlock_tests"`
	SimlockCarriers         []enums.ReportCarrier    `json:"simlock_carriers"`
	Technologies            []enums.ReportTechnology `json:"technologies"`
	SAETests                []enums.SAETest          `json:"sae_tests"`
	SAEOperators            []string                 `json:"sae_operators"`
	SAEChannels             []string                 `json:"sae_channels"`
	SAEScenarios            []MultibandaReportOption `json:"sae_scenarios"`
	EvidenceTypes           []MultibandaReportOption `json:"evidence_types"`
	BandResultOptions       []string                 `json:"band_result_options"`
	PassFailOptions         []string                 `json:"pass_fail_options"`
	SAEResultOptions        []string                 `json:"sae_result_options"`
}

type MultibandaReportOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// MultibandaReportGenerated is returned once the PDF exists.
type MultibandaReportGenerated struct {
	ReportURL   string `json:"report_url"`
	FileName    string `json:"file_name"`
	GeneratedAt string `json:"generated_at"`
}
