package functions

import (
	"strings"
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func TestFormatMultibandaOsVersionPrefersOsVersionView(t *testing.T) {
	got := FormatMultibandaOsVersion("Android 17", "Android", "17")
	if got != "Android 17" {
		t.Fatalf("expected Android 17, got %q", got)
	}
}

func TestFormatMultibandaOsVersionBuildsFromPlatformAndVersion(t *testing.T) {
	got := FormatMultibandaOsVersion("", "Android", "17")
	if got != "Android 17" {
		t.Fatalf("expected Android 17, got %q", got)
	}
}

func TestFormatMultibandaOsVersionPrefixesOsVersionViewWhenMissingPlatform(t *testing.T) {
	got := FormatMultibandaOsVersion("17", "Android", "17")
	if got != "Android 17" {
		t.Fatalf("expected Android 17, got %q", got)
	}
}

func TestMultibandaReflashURLForEmailOnlyOnCreate(t *testing.T) {
	notify := &request.MultibandaNotify{
		NeedReflash:     true,
		CommentsReflash: "https://example.com/reflash",
	}
	if got := multibandaReflashURLForEmail(utils.CREATE, notify.NeedReflash, notify.CommentsReflash); got != notify.CommentsReflash {
		t.Fatalf("expected reflash url on create, got %q", got)
	}
	if got := multibandaReflashURLForEmail(utils.PHASE, notify.NeedReflash, notify.CommentsReflash); got != "" {
		t.Fatalf("expected empty reflash url on phase email, got %q", got)
	}
}

func TestRenderMultibandaReportEmailShowsHeaderAndDownload(t *testing.T) {
	notify := &request.MultibandaNotify{
		CompanyName:   "Client",
		CurrentPhase:  enums.HomologationPhase_value["UNDER_EVALUATION"],
		TestReportUrl: "https://example.com/report.pdf",
	}
	data := BuildMultibandaPhaseEmailData(notify, "Brand", "TM-1", "Model", "Android", 0,
		"User", utils.MULTIBANDA_REPORT_MAIN_MESSAGE, false, "", utils.REPORT)

	html, err := RenderMultibandaPhaseEmailHTML(data, templatePath(t, utils.TEMPLATE_MULTIBANDA_PHASE_PATH))
	if err != nil {
		t.Fatalf("render report email: %v", err)
	}
	body := string(html)
	for _, want := range []string{
		"Multibanda Test Report Notification",
		utils.MULTIBANDA_REPORT_MAIN_MESSAGE,
		"https://example.com/report.pdf",
		"Download Test Report",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("report email is missing %q", want)
		}
	}
}

func renderReportEmailWith(t *testing.T, details MultibandaReportEmailDetails) string {
	t.Helper()
	data := BuildMultibandaPhaseEmailData(&request.MultibandaNotify{CompanyName: "Client"},
		"Brand", "TM-1", "Model", "Android", 0, "User", utils.MULTIBANDA_REPORT_MAIN_MESSAGE, false, "", utils.REPORT)
	ApplyMultibandaReportEmailDetails(&data, details)
	html, err := RenderMultibandaPhaseEmailHTML(data, templatePath(t, utils.TEMPLATE_MULTIBANDA_PHASE_PATH))
	if err != nil {
		t.Fatalf("render report email: %v", err)
	}
	return string(html)
}

func TestRenderMultibandaReportEmailShowsFMRadioAndStampLabel(t *testing.T) {
	body := renderReportEmailWith(t, MultibandaReportEmailDetails{
		FMRadio:   enums.ReportFMRadioSupported,
		StampCode: "5g_and_below",
	})
	for _, want := range []string{
		"<strong>FM Radio</strong>", ">Supported<",
		"<strong>Stamp Type</strong>", "Approved — all bands (5G)",
		">5G</span>", ">SAE</span>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("report email is missing %q", want)
		}
	}
	if strings.Contains(body, "5g_and_below") {
		t.Error("report email shows the stamp code instead of its label")
	}
}

func TestRenderMultibandaReportEmailDashesMissingValues(t *testing.T) {
	body := renderReportEmailWith(t, MultibandaReportEmailDetails{})
	for _, want := range []string{
		`<td style="padding:10px;"><strong>FM Radio</strong></td>
                  <td style="padding:10px;">—</td>`,
		`<td style="padding:10px;">—</td>
                </tr>`,
	} {
		if !strings.Contains(strings.ReplaceAll(body, "\r\n", "\n"), want) {
			t.Errorf("report email is missing the dash cell %q", want)
		}
	}
	if strings.Contains(body, "<span") {
		t.Error("no stamp chips expected without a stamp")
	}
}

func TestMultibandaStampChipsMarksUnsupportedBands(t *testing.T) {
	chips := MultibandaStampChips("2g_4g")
	got := map[string]bool{}
	var labels []string
	for _, c := range chips {
		got[c.Label] = c.Supported
		labels = append(labels, c.Label)
	}
	if strings.Join(labels, " ") != "2G 3G 4G SAE" {
		t.Fatalf("chips = %v", labels)
	}
	if !got["2G"] || got["3G"] || !got["4G"] || !got["SAE"] {
		t.Fatalf("supported flags = %v", got)
	}
}

func TestMultibandaEmailGreetingFallsBackToCompany(t *testing.T) {
	data := BuildMultibandaPhaseEmailData(&request.MultibandaNotify{CompanyName: "  "}, "Brand", "TM-1", "Model",
		"Android", 0, "User", utils.MULTIBANDA_UNDER_MAIN_MESSAGE, false, "", utils.PHASE)
	if data.ClientName != "Company" {
		t.Fatalf("greeting fallback = %q, want Company", data.ClientName)
	}
}

func TestExcelMultibandaFirstHeaderIsCompany(t *testing.T) {
	if got := enums.ExcelMultibandaHeaders["A1"]; got != "Company" {
		t.Fatalf("multibanda export A1 = %q, want Company", got)
	}
}

// Only the report email carries the FM Radio / Stamp Type rows.
func TestRenderMultibandaPhaseEmailOmitsReportDetails(t *testing.T) {
	data := BuildMultibandaPhaseEmailData(&request.MultibandaNotify{}, "Brand", "TM-1", "Model", "Android", 0,
		"User", utils.MULTIBANDA_UNDER_MAIN_MESSAGE, false, "", utils.PHASE)
	html, err := RenderMultibandaPhaseEmailHTML(data, templatePath(t, utils.TEMPLATE_MULTIBANDA_PHASE_PATH))
	if err != nil {
		t.Fatalf("render phase email: %v", err)
	}
	if strings.Contains(string(html), "Stamp Type") {
		t.Error("phase email shows report details")
	}
}

// Phase emails keep their original header.
func TestRenderMultibandaPhaseEmailKeepsPhaseHeader(t *testing.T) {
	data := BuildMultibandaPhaseEmailData(&request.MultibandaNotify{}, "Brand", "TM-1", "Model", "Android", 0,
		"User", utils.MULTIBANDA_UNDER_MAIN_MESSAGE, false, "", utils.PHASE)

	html, err := RenderMultibandaPhaseEmailHTML(data, templatePath(t, utils.TEMPLATE_MULTIBANDA_PHASE_PATH))
	if err != nil {
		t.Fatalf("render phase email: %v", err)
	}
	if !strings.Contains(string(html), "Multibanda Phase Notification") {
		t.Error("phase email lost its header")
	}
}

func TestFormatMultibandaSarValue(t *testing.T) {
	if got := FormatMultibandaSarValue(0); got != "—" {
		t.Fatalf("expected dash for zero SAR, got %q", got)
	}
	if got := FormatMultibandaSarValue(1.25); got != "1.25 W/Kg" {
		t.Fatalf("expected 1.25 W/Kg, got %q", got)
	}
	if got := FormatMultibandaSarValue(1.09); got != "1.09 W/Kg" {
		t.Fatalf("expected 1.09 W/Kg, got %q", got)
	}
}

func TestApplyMultibandaPhaseDateRulesSetsUnderStartFromTestEndOnUnderEvaluation(t *testing.T) {
	testEnd := time.Date(2026, 5, 29, 6, 0, 0, 0, time.UTC)

	multibanda := &models.Multibanda{
		CurrentPhase: enums.HomologationPhase_value["UNDER_EVALUATION"],
		TestEndDate:  &testEnd,
	}

	ApplyMultibandaPhaseDateRules(multibanda, nil)

	if !multibanda.UnderStartDate.Equal(testEnd) {
		t.Fatalf("expected under_start_date %v, got %v", testEnd, multibanda.UnderStartDate)
	}
}

func TestApplyMultibandaPhaseDateRulesUsesExistingTestEndWhenRequestIsEmpty(t *testing.T) {
	testEnd := time.Date(2026, 5, 29, 6, 0, 0, 0, time.UTC)

	multibanda := &models.Multibanda{
		CurrentPhase: enums.HomologationPhase_value["UNDER_EVALUATION"],
	}
	existing := &responses.MultibandaExpanded{
		TestEndDate: &testEnd,
	}

	ApplyMultibandaPhaseDateRules(multibanda, existing)

	if !multibanda.UnderStartDate.Equal(testEnd) {
		t.Fatalf("expected under_start_date %v, got %v", testEnd, multibanda.UnderStartDate)
	}
	if !multibanda.TestEndDate.Equal(testEnd) {
		t.Fatalf("expected test_end_date %v, got %v", testEnd, multibanda.TestEndDate)
	}
}

func TestApplyMultibandaPhaseDateRulesSkipsOtherPhases(t *testing.T) {
	testEnd := time.Date(2026, 5, 29, 6, 0, 0, 0, time.UTC)

	multibanda := &models.Multibanda{
		CurrentPhase: enums.HomologationPhase_value["TEST"],
		TestEndDate:  &testEnd,
	}

	ApplyMultibandaPhaseDateRules(multibanda, nil)

	if multibanda.UnderStartDate != nil {
		t.Fatalf("expected empty under_start_date, got %v", multibanda.UnderStartDate)
	}
}
