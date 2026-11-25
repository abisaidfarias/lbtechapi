package functions

import (
	"os"
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

// TestWriteMultibandaReportEmailSample renders a realistic "Test Report
// Available" email for visual review. It only runs when
// MULTIBANDA_REPORT_EMAIL_SAMPLE names the output file:
//
//	MULTIBANDA_REPORT_EMAIL_SAMPLE=/tmp/report_email.html go test ./utils/functions -run WriteMultibandaReportEmailSample
func TestWriteMultibandaReportEmailSample(t *testing.T) {
	out := os.Getenv("MULTIBANDA_REPORT_EMAIL_SAMPLE")
	if out == "" {
		t.Skip("set MULTIBANDA_REPORT_EMAIL_SAMPLE to write the sample")
	}

	day := func(d int) time.Time { return time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC) }
	notify := &request.MultibandaNotify{
		CompanyName:     "Samsung Electronics Chile",
		CurrentPhase:    enums.HomologationPhase_value["UNDER_EVALUATION"],
		Type:            enums.MultibandaTypeSMR,
		SoftwareVersion: "S928BXXU3AXK1",
		HardwareVersion: "REV1.0",
		OsVersionView:   "Android 15",
		PlanningDate:    day(3),
		SampleStartDate: day(5),
		SampleEndDate:   day(6),
		TestStartDate:   day(10),
		TestEndDate:     day(18),
		UnderStartDate:  day(18),
		TestReportUrl:   "https://example.com/multibanda/report.pdf",
	}
	data := BuildMultibandaPhaseEmailData(notify, "Samsung", "SM-S928B", "Galaxy S24 Ultra", "Android", 1.09,
		"Camilo Pérez", utils.MULTIBANDA_REPORT_MAIN_MESSAGE, false, "", utils.REPORT)
	ApplyMultibandaReportEmailDetails(&data, MultibandaReportEmailDetails{
		FMRadio:   enums.ReportFMRadioSupported,
		StampCode: "5g_and_below",
	})
	data.LogoDataURI = resolveBrandedEmailLogo(utils.LBOneTrackLogoPNG)

	html, err := RenderMultibandaPhaseEmailHTML(data, templatePath(t, utils.TEMPLATE_MULTIBANDA_PHASE_PATH))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if err := os.WriteFile(out, html, 0o644); err != nil {
		t.Fatalf("write sample: %v", err)
	}
	t.Logf("sample written to %s", out)
}
