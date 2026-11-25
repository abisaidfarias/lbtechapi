package functions

import (
	"strings"
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func boolPtr(v bool) *bool { return &v }

func TestComputePhaseStates(t *testing.T) {
	got := ComputePhaseStates(2, nil)
	want := models.MultibandaPhaseStates{Planning: "done", SampleReception: "done", Testing: "pending",
		UnderEvaluation: "pending", Complete: "pending"}
	if *got != want {
		t.Fatalf("phase 2 = %+v", *got)
	}
	if done := ComputePhaseStates(4, nil); done.Complete != "done" || done.UnderEvaluation != "done" {
		t.Fatalf("phase 4 = %+v", *done)
	}
	// Skipped phases stay skipped as the request advances.
	got = ComputePhaseStates(4, SmrApprovedStates())
	want = models.MultibandaPhaseStates{Planning: "done", SampleReception: "skipped", Testing: "skipped",
		UnderEvaluation: "done", Complete: "done"}
	if *got != want {
		t.Fatalf("approved SMR closed = %+v", *got)
	}
}

func TestRequiresTestingApplies(t *testing.T) {
	sae := []string{enums.MultibandaEvalSAEMultibandaCertificate}
	if !RequiresTestingApplies("smr", sae, true) {
		t.Error("smr + sae + company allows must apply")
	}
	for _, c := range []struct {
		typ   string
		evals []string
		allow bool
	}{
		{"smr", sae, false}, {"initial", sae, true}, {"smr", []string{"other"}, true}, {"smr", nil, true},
	} {
		if RequiresTestingApplies(c.typ, c.evals, c.allow) {
			t.Errorf("%+v must not apply", c)
		}
	}
}

func TestNormalizeRequiresTesting(t *testing.T) {
	// Enabled: required.
	if _, err := NormalizeRequiresTesting(true, nil); !utils.IsValidationError(err) || err.Error() != SmrErrRequiresTestingNeeded {
		t.Errorf("enabled without value: %v", err)
	}
	if v, err := NormalizeRequiresTesting(true, boolPtr(false)); err != nil || v == nil || *v {
		t.Errorf("enabled false: %v %v", v, err)
	}
	// Not enabled: false rejected, true/null stored as null.
	if _, err := NormalizeRequiresTesting(false, boolPtr(false)); !utils.IsValidationError(err) || err.Error() != SmrErrNotEnabled {
		t.Errorf("not enabled false: %v", err)
	}
	for _, v := range []*bool{nil, boolPtr(true)} {
		if got, err := NormalizeRequiresTesting(false, v); err != nil || got != nil {
			t.Errorf("not enabled %v: %v %v", v, got, err)
		}
	}
}

func TestGuardSmrChange(t *testing.T) {
	cases := []struct {
		status            string
		phase, newPhase   int
		statusV, newState int
		want              string
	}{
		{models.SmrValidationPending, 0, 1, 0, 0, SmrErrPendingValidation}, // can't go to sample reception
		{models.SmrValidationPending, 0, 0, 0, 2, SmrErrPendingValidation}, // can't change status
		{models.SmrValidationPending, 0, 0, 0, 0, ""},                      // editing data is fine
		{models.SmrValidationRejected, 4, 4, 2, 2, SmrErrRejectedClosed},   // closed
		{models.SmrValidationRejected, 4, 3, 2, 0, SmrErrRejectedClosed},   // can't reopen
		{models.SmrValidationApproved, 3, 2, 0, 0, SmrErrNoTestingPhases},  // no testing
		{models.SmrValidationApproved, 3, 4, 0, 1, ""},                     // normal close
		{"", 0, 1, 0, 0, ""}, // not an SMR w/o testing
	}
	for _, c := range cases {
		err := GuardSmrChange(c.status, c.phase, c.statusV, c.newPhase, c.newState)
		got := ""
		if err != nil {
			got = err.Error()
		}
		if got != c.want {
			t.Errorf("%+v: got %q", c, got)
		}
	}
}

func TestSmrApproveCheck(t *testing.T) {
	stamp := "5g_and_below"
	cases := []struct {
		name          string
		origFM        *bool
		origStamp     string
		fm            *bool
		stampCode     string
		comment, want string
	}{
		{"confirm both", boolPtr(true), stamp, boolPtr(true), stamp, "", ""},
		{"fill empty values without comment", nil, "", boolPtr(false), stamp, "", ""},
		{"correct FM without comment", boolPtr(true), stamp, boolPtr(false), stamp, "", SmrErrCommentOnCorrection},
		{"correct stamp without comment", boolPtr(true), stamp, boolPtr(true), "4g", "", SmrErrCommentOnCorrection},
		{"correct with comment", boolPtr(true), stamp, boolPtr(false), "4g", "radio removed", ""},
		{"missing stamp", boolPtr(true), stamp, boolPtr(true), " ", "", SmrErrStampRequired},
		{"invalid stamp", boolPtr(true), stamp, boolPtr(true), "nope", "x", SmrErrStampInvalid},
		{"missing fm", boolPtr(true), stamp, nil, stamp, "", SmrErrFMRadioRequired},
	}
	for _, c := range cases {
		err := SmrApproveCheck(c.origFM, c.origStamp, c.fm, c.stampCode, c.comment)
		got := ""
		if err != nil {
			got = err.Error()
		}
		if got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestComputeSmrKpi(t *testing.T) {
	created := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	in2 := created.Add(48 * time.Hour)
	in4 := created.Add(96 * time.Hour)
	items := []SmrKpiInput{
		{Type: "smr", RequiresTesting: boolPtr(true)},
		{Type: "smr", RequiresTesting: nil}, // not eligible → with testing
		{Type: "smr", RequiresTesting: boolPtr(false), ValidationStatus: "approved", CreatedDate: created, DecidedAt: &in2},
		{Type: "smr", RequiresTesting: boolPtr(false), ValidationStatus: "rejected", CreatedDate: created, DecidedAt: &in4},
		{Type: "smr", RequiresTesting: boolPtr(false), ValidationStatus: "pending", CreatedDate: created},
		{Type: "initial", RequiresTesting: nil},
	}
	k := ComputeSmrKpi(items)
	want := SmrKpi{WithTesting: 2, WithoutTesting: 3, Approved: 1, Rejected: 1, Pending: 1, AvgDaysToDecision: 3}
	if k != want {
		t.Fatalf("kpi = %+v, want %+v", k, want)
	}
}

// The approval email of an SMR without testing has no download and N/A for
// the skipped sample/testing dates.
func TestReportEmailWithoutTesting(t *testing.T) {
	data := MultibandaPhaseEmailData{TestReportURL: "https://x/report.pdf", SampleStartDate: "01/11/2026", TestStartDate: "02/11/2026"}
	ApplyMultibandaReportEmailDetails(&data, MultibandaReportEmailDetails{
		FMRadio: "Supported", StampCode: "5g_and_below", WithoutTesting: true,
	})
	if data.TestReportURL != "" || data.SampleStartDate != "N/A" || data.SampleEndDate != "N/A" ||
		data.TestStartDate != "N/A" || data.TestEndDate != "N/A" || data.FMRadio != "Supported" ||
		data.StampLabel == "—" {
		t.Fatalf("data = %+v", data)
	}
}

// The intro of the SMR-without-testing email must not claim a test report
// was generated; the subject stays "Test Report Available".
func TestReportEmailWithoutTestingIntro(t *testing.T) {
	notify := &request.MultibandaNotify{CompanyName: "VIVO", TestReportUrl: "https://x/report.pdf"}
	data := BuildMultibandaPhaseEmailData(notify, "VIVO", "V2410", "Y29", "Android", 0,
		"Ana Soto", utils.MULTIBANDA_REPORT_MAIN_MESSAGE, false, "", utils.REPORT)
	date := data.NotificationDate
	ApplyMultibandaReportEmailDetails(&data, MultibandaReportEmailDetails{
		FMRadio: "Not Supported", StampCode: "5g_and_below", WithoutTesting: true,
	})
	html, err := RenderMultibandaPhaseEmailHTML(data, templatePath(t, utils.TEMPLATE_MULTIBANDA_PHASE_PATH))
	if err != nil {
		t.Fatal(err)
	}
	body := string(html)
	for _, want := range []string{
		"The following SMR without testing was validated on " + date + ".",
		"so there is no test report", ">Not Supported<", "N/A",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, unwanted := range []string{utils.MULTIBANDA_REPORT_MAIN_MESSAGE, "Download Test Report"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("must not contain %q", unwanted)
		}
	}

	// A normal report keeps its intro.
	normal := BuildMultibandaPhaseEmailData(notify, "VIVO", "V2410", "Y29", "Android", 0,
		"Ana Soto", utils.MULTIBANDA_REPORT_MAIN_MESSAGE, false, "", utils.REPORT)
	ApplyMultibandaReportEmailDetails(&normal, MultibandaReportEmailDetails{FMRadio: "Supported", StampCode: "2g"})
	if normal.MainMessage != utils.MULTIBANDA_REPORT_MAIN_MESSAGE || normal.NotificationDate == "" {
		t.Errorf("normal report intro changed: %q %q", normal.MainMessage, normal.NotificationDate)
	}
}
