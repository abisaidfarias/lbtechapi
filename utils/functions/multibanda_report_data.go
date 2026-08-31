package functions

import (
	"strings"
	"time"
)

// BuildMultibandaReportFileName is the download name for the generated report:
// "<commercial name> - <sw version> - <test date>.pdf", skipping any part that
// is missing rather than leaving a stray separator.
func BuildMultibandaReportFileName(commercialName, softwareVersion string, testDate time.Time) string {
	segments := []string{
		strings.TrimSpace(commercialName),
		strings.TrimSpace(softwareVersion),
	}
	if !testDate.IsZero() {
		segments = append(segments, testDate.Format("02-01-2006"))
	}

	nonEmpty := make([]string, 0, len(segments))
	for _, s := range segments {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	if len(nonEmpty) == 0 {
		return "device-evaluation-report.pdf"
	}
	return strings.Join(nonEmpty, " - ") + ".pdf"
}

// MultibandaReportPDFData is the fully resolved input for the Automatic
// Multi-Band Certification Report PDF. Everything the renderer needs is
// already flattened here: no lookups, no conditional resolution, so the
// rendering code stays about layout only.
type MultibandaReportPDFData struct {
	// Header
	ProcessTypeLabel string
	ReportDate       string

	// Device information (prefilled + engineer-entered)
	Manufacturer           string
	CommercialName         string
	TechnicalModel         string
	HardwareVersion        string
	SoftwareVersion        string
	SARValue               string
	CBSPackage             string
	GooglePlaySystemUpdate string
	OperativeSystemType    string
	PreferredNetwork       string
	FMRadio                string
	TestDate               string
	IMEI                   string
	SerialNumber           string

	// Images already fetched as bytes so the renderer never does I/O.
	StampImage []byte
	StampLabel string

	// SWVersionPhoto is the build screenshot shown in Device Information; the
	// Pop-Up and CB History captures live in SAEEvidence.
	SWVersionPhoto []byte

	// CarriersTested are the carriers this device was evaluated against.
	CarriersTested []string

	// Block scope: SIMLOCK and Multi-band only apply to Initial processes.
	IncludesSimlock   bool
	IncludesMultiband bool

	SimlockRows []MultibandaReportSimlockRow
	BandGroups  []MultibandaReportBandGroup
	SAEScenario string
	SAEBlocks   []MultibandaReportSAEBlock
	SAEEvidence []MultibandaReportEvidenceItem

	// FM Radio declaration plus the conditional non-removable test.
	FMRadioTestApplies bool
	FMRadioTestResult  string
	FMRadioTestComment string

	GeneratedAt time.Time
	Year        int
}

type MultibandaReportSimlockRow struct {
	TestID  string
	Name    string
	Carrier string
	MNC     string
	Result  string
	Comment string
}

type MultibandaReportBandGroup struct {
	Technology string
	Bands      []MultibandaReportBandRow
}

type MultibandaReportBandRow struct {
	Band    string
	Result  string
	Comment string
}

// MultibandaReportSAEBlock is one scenario+channel matrix: eleven tests by
// four operators.
type MultibandaReportSAEBlock struct {
	ScenarioLabel string
	Channel       string
	Operators     []string
	Rows          []MultibandaReportSAERow
}

type MultibandaReportSAERow struct {
	TestID string
	Name   string
	// Results is aligned with the block's Operators slice.
	Results  []string
	Comments []string
}

type MultibandaReportEvidenceItem struct {
	Label         string
	ScenarioLabel string
	Operator      string
	Image         []byte
}
