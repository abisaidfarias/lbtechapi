package functions

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func FormatMultibandaTypeLabel(value string) string {
	if label, ok := enums.MultibandaTypeLabels[value]; ok {
		return label
	}
	return value
}

func FormatMultibandaEvaluationTypes(values []string) string {
	if len(values) == 0 {
		return "—"
	}

	labels := make([]string, 0, len(values))
	for _, value := range values {
		if label, ok := enums.MultibandaEvaluationTypeLabels[value]; ok {
			labels = append(labels, label)
			continue
		}
		labels = append(labels, value)
	}
	return strings.Join(labels, ", ")
}

func MultibandaPhaseLabel(phase int) string {
	if label, ok := enums.MultibandaPhaseLabels[phase]; ok {
		return label
	}
	return fmt.Sprintf("Phase %d", phase)
}

func FormatMultibandaEmailDate(date time.Time) string {
	if date.IsZero() {
		return "-"
	}
	return fmt.Sprintf("%02d/%02d/%d", date.Day(), date.Month(), date.Year())
}

func FormatMultibandaOsVersion(osVersionView, platformOs, osVersion string) string {
	platformOs = strings.TrimSpace(platformOs)
	osVersionView = strings.TrimSpace(osVersionView)
	osVersion = strings.TrimSpace(osVersion)

	version := osVersionView
	if version == "" {
		version = osVersion
	}

	if version == "" {
		return "—"
	}

	if platformOs == "" {
		return version
	}

	if strings.HasPrefix(strings.ToLower(version), strings.ToLower(platformOs)) {
		return version
	}

	return platformOs + " " + version
}

func FormatMultibandaSarValue(value float64) string {
	if value == 0 {
		return "—"
	}
	formatted := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
	return formatted + " W/Kg"
}

func multibandaReflashURLForEmail(emailKind string, needReflash bool, commentsReflash string) string {
	if emailKind != utils.CREATE || !needReflash {
		return ""
	}
	return strings.TrimSpace(commentsReflash)
}

func multibandaEmailHeaderTitle(emailKind string) string {
	if emailKind == utils.REPORT {
		return "Multibanda Test Report Notification"
	}
	return "Multibanda Phase Notification"
}

func BuildMultibandaPhaseEmailData(
	multibanda *request.MultibandaNotify,
	brandName string,
	technicalModel string,
	commercialModel string,
	platformOs string,
	sarValue float64,
	userName string,
	mainMessage string,
	finished bool,
	decision string,
	emailKind string,
) MultibandaPhaseEmailData {
	projectType := "External"
	if multibanda.IsInternalProject {
		projectType = "Internal"
	}

	now := time.Now()
	clientName := strings.TrimSpace(multibanda.CompanyName)
	if clientName == "" {
		clientName = "Company"
	}

	return MultibandaPhaseEmailData{
		HeaderTitle:             multibandaEmailHeaderTitle(emailKind),
		ClientName:              clientName,
		MainMessage:             mainMessage,
		NotificationDate:        FormatMultibandaEmailDate(now),
		CurrentPhase:            MultibandaPhaseLabel(multibanda.CurrentPhase),
		ProjectType:             projectType,
		ProcessType:             FormatMultibandaTypeLabel(multibanda.Type),
		EvaluationTypes:         FormatMultibandaEvaluationTypes(multibanda.EvaluationType),
		Brand:                   brandName,
		TechnicalModel:          technicalModel,
		CommercialModel:         commercialModel,
		SoftwareVersion:         emptyAsDash(multibanda.SoftwareVersion),
		HardwareVersion:         emptyAsDash(multibanda.HardwareVersion),
		OsVersion:               FormatMultibandaOsVersion(multibanda.OsVersionView, platformOs, multibanda.OsVersion),
		SARValue:                FormatMultibandaSarValue(sarValue),
		UpdatedBy:               emptyAsDash(userName),
		PlanningDate:            FormatMultibandaEmailDate(multibanda.PlanningDate),
		SampleStartDate:         FormatMultibandaEmailDate(multibanda.SampleStartDate),
		SampleEndDate:           FormatMultibandaEmailDate(multibanda.SampleEndDate),
		TestStartDate:           FormatMultibandaEmailDate(multibanda.TestStartDate),
		TestEndDate:             FormatMultibandaEmailDate(multibanda.TestEndDate),
		UnderStartDate:          FormatMultibandaEmailDate(multibanda.UnderStartDate),
		UnderEndDate:            FormatMultibandaEmailDate(multibanda.UnderEndDate),
		ResultDate:              FormatMultibandaEmailDate(multibanda.CompletedDate),
		Finished:                finished,
		Decision:                emptyAsDash(decision),
		TestReportURL:           strings.TrimSpace(multibanda.TestReportUrl),
		MultibandCertificateURL: strings.TrimSpace(multibanda.MultibandCertificateUrl),
		ReflashURL:              multibandaReflashURLForEmail(emailKind, multibanda.NeedReflash, multibanda.CommentsReflash),
		Year:                    now.Year(),
	}
}

// MultibandaReportEmailDetails carries the report values the "Test Report
// Available" email prints, taken from the stored report so they match the PDF.
type MultibandaReportEmailDetails struct {
	FMRadio   string
	StampCode string
	// WithoutTesting: SMR without testing approved. There is no test report
	// to download and the sample/testing dates print "N/A".
	WithoutTesting bool
}

// NotApplicable is how skipped phases print in emails.
const NotApplicable = "N/A"

// ApplySkippedTestingDates prints the sample reception and testing dates as
// N/A for an SMR without testing.
func ApplySkippedTestingDates(data *MultibandaPhaseEmailData) {
	data.SampleStartDate = NotApplicable
	data.SampleEndDate = NotApplicable
	data.TestStartDate = NotApplicable
	data.TestEndDate = NotApplicable
}

// MultibandaRequestURL links an email to the request in the platform.
func MultibandaRequestURL(multibandaID string) string {
	base := FrontendBaseURL()
	if base == "" {
		base = "https://lbonetrack.com"
	}
	// The list opens the validation of that request (there is no detail route).
	return base + "/multibanda?validate=" + multibandaID
}

// MultibandaStampChip is one technology badge of the stamp, rendered as an
// inline-styled span. Unsupported technologies are greyed and struck through,
// as they are crossed out on the stamp artwork.
type MultibandaStampChip struct {
	Label     string
	Style     template.CSS
	Supported bool
}

var stampChipColors = map[string]string{
	"2G":  "#29a9e0",
	"3G":  "#7ab829",
	"4G":  "#8c1d5a",
	"5G":  "#1d4e5f",
	"SAE": "#f07d00",
}

const stampChipBaseStyle = "display:inline-block; padding:2px 6px; margin-right:4px; " +
	"border-radius:3px; font-size:11px; font-weight:bold; line-height:16px; "

// MultibandaStampChips derives the badges from the stamp code, which names
// the bands the device supports ("2g_4g" = 2G and 4G, not 3G). Every stamp
// prints 2G/3G/4G and SAE; 5G is shown only for the stamp approved with 5G.
// An unknown or empty code yields no chips.
func MultibandaStampChips(stampCode string) []MultibandaStampChip {
	code := strings.TrimSpace(stampCode)
	if _, ok := enums.StampByCode(code); !ok {
		return nil
	}

	supported := map[string]bool{}
	switch code {
	case "5g_and_below":
		supported = map[string]bool{"2G": true, "3G": true, "4G": true, "5G": true}
	case "4g_and_below":
		supported = map[string]bool{"2G": true, "3G": true, "4G": true}
	default:
		for _, part := range strings.Split(code, "_") {
			supported[strings.ToUpper(part)] = true
		}
	}

	labels := []string{"2G", "3G", "4G"}
	if supported["5G"] {
		labels = append(labels, "5G")
	}
	labels = append(labels, "SAE")

	chips := make([]MultibandaStampChip, 0, len(labels))
	for _, label := range labels {
		ok := label == "SAE" || supported[label]
		style := stampChipBaseStyle + "color:#ffffff; background-color:" + stampChipColors[label] + ";"
		if !ok {
			style = stampChipBaseStyle + "color:#6b7280; background-color:#e5e7eb; text-decoration:line-through;"
		}
		chips = append(chips, MultibandaStampChip{Label: label, Style: template.CSS(style), Supported: ok})
	}
	return chips
}

// ApplyMultibandaReportEmailDetails fills the FM Radio and Stamp Type rows of
// the report email; missing values print as "—".
func ApplyMultibandaReportEmailDetails(data *MultibandaPhaseEmailData, details MultibandaReportEmailDetails) {
	data.ShowReportDetails = true
	data.FMRadio = emptyAsDash(details.FMRadio)
	data.StampLabel = "—"
	data.StampChips = nil
	if stamp, ok := enums.StampByCode(strings.TrimSpace(details.StampCode)); ok {
		data.StampLabel = stamp.Label
		data.StampChips = MultibandaStampChips(stamp.Code)
	}
	if details.WithoutTesting {
		data.TestReportURL = ""
		ApplySkippedTestingDates(data)
		// There is no test report: the intro must not say one was generated.
		// The template prints "MainMessage NotificationDate", so the date goes
		// inside the message and NotificationDate is cleared.
		data.MainMessage = fmt.Sprintf(SmrWithoutTestingReportMessage, data.NotificationDate)
		data.NotificationDate = ""
	}
}

// SmrWithoutTestingReportMessage replaces the "report available" intro for an
// SMR without testing (%s = validation date).
const SmrWithoutTestingReportMessage = "The following SMR without testing was validated on %s. " +
	"It does not require samples nor testing, so there is no test report; " +
	"FM Radio and the stamp below were confirmed by LB Technology."

func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return strings.TrimSpace(value)
}
