package enums

// Catalogs for the Automatic Multi-Band Certification Report.
//
// The functional spec fixes these values for the first release but explicitly
// asks that technologies, bands, tests, operators and stamps live as catalog
// data so future regulatory changes do not require redesigning the form.
// They are defined here (not hardcoded at the call sites) so a later move to
// a database-backed catalog only changes the source of these slices.

// ---- Report block identifiers ----

const (
	ReportBlockDeviceInfo = "device_info"
	ReportBlockSimlock    = "simlock"
	ReportBlockMultiband  = "multiband"
	ReportBlockSAE        = "sae"
	ReportBlockFMRadio    = "fm_radio"
)

// ---- Report status ----

const (
	MultibandaReportStatusDraft     = "draft"
	MultibandaReportStatusGenerated = "generated"
)

// ---- Result values ----

const (
	ReportResultPass = "PASS"
	ReportResultFail = "FAIL"
	ReportResultNA   = "N/A"

	// Multi-band consolidated results.
	ReportResultOK           = "OK"
	ReportResultNoOK         = "NO OK"
	ReportResultNotSupported = "NOT SUPPORTED"
)

// ResultRequiresComment reports whether a result value obliges the engineer to
// justify it. The spec makes a comment mandatory on every FAIL and NO OK.
func ResultRequiresComment(result string) bool {
	return result == ReportResultFail || result == ReportResultNoOK
}

var allowedPassFailResults = map[string]struct{}{
	ReportResultPass: {},
	ReportResultFail: {},
}

var allowedSAEResults = map[string]struct{}{
	ReportResultPass: {},
	ReportResultFail: {},
	ReportResultNA:   {},
}

var allowedBandResults = map[string]struct{}{
	ReportResultOK:           {},
	ReportResultNoOK:         {},
	ReportResultNotSupported: {},
}

func IsValidPassFailResult(v string) bool {
	_, ok := allowedPassFailResults[v]
	return ok
}

func IsValidSAEResult(v string) bool {
	_, ok := allowedSAEResults[v]
	return ok
}

func IsValidBandResult(v string) bool {
	_, ok := allowedBandResults[v]
	return ok
}

// ---- Device information dropdowns ----

var ReportPreferredNetworkOptions = []string{
	"5G/4G/3G/2G",
	"4G/3G/2G",
	"3G/2G",
}

const (
	ReportFMRadioSupported    = "Supported"
	ReportFMRadioNotSupported = "Not Supported"
	ReportFMRadioNA           = "N/A"
)

var ReportFMRadioOptions = []string{
	ReportFMRadioSupported,
	ReportFMRadioNotSupported,
	ReportFMRadioNA,
}

func IsValidPreferredNetwork(v string) bool {
	for _, o := range ReportPreferredNetworkOptions {
		if o == v {
			return true
		}
	}
	return false
}

func IsValidFMRadioOption(v string) bool {
	for _, o := range ReportFMRadioOptions {
		if o == v {
			return true
		}
	}
	return false
}

// ---- Stamps ----

// ReportStamp is one selectable stamp image. Code is the stable identifier
// stored on the report; ImageKey points at the artwork embedded in the PDF.
//
// Both name the bands the device DOES support, matching the artwork files:
// "2g_4g" is a device supporting 2G and 4G but not 3G. The two "and_below"
// stamps are the fully approved ones — they differ only in whether the 5G
// advisory strip is printed.
type ReportStamp struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description"`
	ImageKey    string `json:"image_key"`
}

var ReportStampCatalog = []ReportStamp{
	{
		Code:        "5g_and_below",
		Label:       "Approved — all bands (5G)",
		Description: "Approved for all carrier bands, including the 5G advisory",
		ImageKey:    "5G_and_below",
	},
	{
		Code:        "4g_and_below",
		Label:       "Approved — all bands (up to 4G)",
		Description: "Approved for all carrier bands, without the 5G advisory",
		ImageKey:    "4G_and_below",
	},
	{
		Code:        "2g_3g",
		Label:       "Supports 2G and 3G",
		Description: "Not approved for 4G",
		ImageKey:    "2G_3G",
	},
	{
		Code:        "2g_4g",
		Label:       "Supports 2G and 4G",
		Description: "Not approved for 3G",
		ImageKey:    "2G_4G",
	},
	{
		Code:        "3g_4g",
		Label:       "Supports 3G and 4G",
		Description: "Not approved for 2G",
		ImageKey:    "3G_4G",
	},
	{
		Code:        "2g",
		Label:       "Supports 2G only",
		Description: "Not approved for 3G or 4G",
		ImageKey:    "2G",
	},
	{
		Code:        "3g",
		Label:       "Supports 3G only",
		Description: "Not approved for 2G or 4G",
		ImageKey:    "3G",
	},
	{
		Code:        "4g",
		Label:       "Supports 4G only",
		Description: "Not approved for 2G or 3G",
		ImageKey:    "4G",
	},
}

func IsValidStampCode(code string) bool {
	for _, s := range ReportStampCatalog {
		if s.Code == code {
			return true
		}
	}
	return false
}

func StampByCode(code string) (ReportStamp, bool) {
	for _, s := range ReportStampCatalog {
		if s.Code == code {
			return s, true
		}
	}
	return ReportStamp{}, false
}

// ---- SIMLOCK catalog (Initial processes only) ----

type ReportTest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var ReportSimlockTests = []ReportTest{
	{ID: "001SL", Name: "Carriers Check"},
	{ID: "002SL", Name: "Preferred Network Check"},
}

// ReportCarrier is a fixed SIMLOCK carrier with its MNC code.
type ReportCarrier struct {
	Name string `json:"name"`
	MNC  string `json:"mnc"`
}

var ReportSimlockCarriers = []ReportCarrier{
	{Name: "Entel", MNC: "01"},
	{Name: "Movistar", MNC: "02"},
	{Name: "Claro", MNC: "03"},
	{Name: "Virgin", MNC: "07"},
	{Name: "VTR", MNC: "08"},
	{Name: "WOM", MNC: "09"},
	{Name: "Mundo Pacífico", MNC: "02/09/28"},
}

// ---- Multi-band catalog (Initial processes only) ----

type ReportTechnology struct {
	Code  string   `json:"code"`
	Label string   `json:"label"`
	Bands []string `json:"bands"`
}

var ReportBandCatalog = []ReportTechnology{
	{Code: "gsm", Label: "GSM (2G)", Bands: []string{"900 MHz", "850 MHz", "1900 MHz"}},
	{Code: "umts", Label: "UMTS (3G)", Bands: []string{"B2", "B4", "B5", "B8"}},
	{Code: "lte", Label: "LTE (4G)", Bands: []string{"B2", "B4", "B66", "B5", "B7", "B28"}},
	{Code: "nr", Label: "5G NR", Bands: []string{"n28", "n66", "n7", "n78"}},
}

// ---- SAE catalog ----

const (
	SAEScenarioLaboratory = "laboratory"
	SAEScenarioRoom       = "sae_room"
	SAEScenarioBoth       = "both"
)

var SAEScenarioLabels = map[string]string{
	SAEScenarioLaboratory: "Laboratory",
	SAEScenarioRoom:       "SAE Room",
}

// SAEScenariosFor expands the engineer's scenario selection into the concrete
// scenarios whose result matrices must be filled in.
func SAEScenariosFor(selection string) []string {
	switch selection {
	case SAEScenarioLaboratory:
		return []string{SAEScenarioLaboratory}
	case SAEScenarioRoom:
		return []string{SAEScenarioRoom}
	case SAEScenarioBoth:
		return []string{SAEScenarioLaboratory, SAEScenarioRoom}
	default:
		return nil
	}
}

func IsValidSAEScenarioSelection(v string) bool {
	return len(SAEScenariosFor(v)) > 0
}

var SAEChannels = []string{"919", "4370"}

var SAEOperators = []string{"Entel", "Movistar", "Claro", "WOM"}

// SAETest is one of the eleven fixed SAE tests. RequiresTechnology names the
// device capability that makes the test applicable; when the device lacks it
// the result is forced to N/A. Empty means the test always applies.
type SAETest struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	RequiresTechnology string `json:"requires_technology,omitempty"`
}

const (
	SAETechGSM  = "gsm"
	SAETechUMTS = "umts"
	SAETechLTE  = "lte"
	SAETechNR   = "nr"
)

var ReportSAETests = []SAETest{
	{ID: "001SAE", Name: "GSM Technology", RequiresTechnology: SAETechGSM},
	{ID: "002SAE", Name: "UMTS Technology", RequiresTechnology: SAETechUMTS},
	{ID: "003SAE", Name: "LTE Technology", RequiresTechnology: SAETechLTE},
	{ID: "004SAE", Name: "5G Technology", RequiresTechnology: SAETechNR},
	{ID: "005SAE", Name: "Pop-Up Notification"},
	{ID: "006SAE", Name: "Sound Alert"},
	{ID: "007SAE", Name: "Vibration"},
	{ID: "008SAE", Name: "Notification Mode"},
	{ID: "009SAE", Name: "CB History"},
	{ID: "010SAE", Name: "CB Configuration"},
	{ID: "011SAE", Name: "Forbidden Channels"},
}

// ---- Photographic evidence ----

// Three screenshots are captured. SW Version documents the build under test and
// is printed in Device Information; the other two are the SAE evidence and
// additionally record the scenario and operator they were taken for.
const (
	EvidenceSWVersion    = "sw_version"
	SAEEvidencePopUp     = "pop_up"
	SAEEvidenceCBHistory = "cb_history"
)

var EvidenceLabels = map[string]string{
	EvidenceSWVersion:    "SW Version",
	SAEEvidencePopUp:     "Pop-Up",
	SAEEvidenceCBHistory: "CB History",
}

// RequiredEvidenceTypes is every screenshot the report needs, in display order.
var RequiredEvidenceTypes = []string{EvidenceSWVersion, SAEEvidencePopUp, SAEEvidenceCBHistory}

// RequiredSAEEvidenceTypes are the two that must also name a scenario and an
// operator, regardless of how many scenarios were executed.
var RequiredSAEEvidenceTypes = []string{SAEEvidencePopUp, SAEEvidenceCBHistory}

func IsValidEvidenceType(v string) bool {
	_, ok := EvidenceLabels[v]
	return ok
}

// EvidenceRequiresScenarioAndOperator reports whether an evidence type must be
// tagged with the scenario and operator it belongs to.
func EvidenceRequiresScenarioAndOperator(evidenceType string) bool {
	return evidenceType == SAEEvidencePopUp || evidenceType == SAEEvidenceCBHistory
}

// ---- FM Radio conditional test ----

const ReportFMRadioTestID = "001FM"
const ReportFMRadioTestName = "FM Radio Application — Non-removable"

// ---- Process type scope ----

// ReportIncludesSimlockAndMultiband reports whether the process type carries
// the SIMLOCK and Multi-band blocks. Per the spec only Initial does; SMR, MR
// and OS Upgrade carry Device Information and SAE only.
func ReportIncludesSimlockAndMultiband(multibandaType string) bool {
	return multibandaType == MultibandaTypeInitialProcess
}
