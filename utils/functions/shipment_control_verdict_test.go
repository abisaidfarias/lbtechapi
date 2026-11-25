package functions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func closedControl(registered, requested int, comment string) *models.ShipmentControl {
	return &models.ShipmentControl{
		CurrentPhase:          enums.ShipmentControlPhaseCompleted,
		RegisteredImeiCount:   registered,
		ImeiQuantity:          requested,
		Comment:               comment,
		OabiCertificate:       "OABI-1",
		OabiCertificateUrl:    "https://s3/oabi.pdf",
		OabiCertificateNumber: "REG-1",
	}
}

func TestApplyShipmentControlVerdictRules(t *testing.T) {
	cases := []struct {
		name       string
		registered int
		requested  int
		comment    string
		want       string
		wantErr    string
	}{
		{"complete", 500, 500, "", enums.ShipmentControlVerdictApproved, ""},
		{"partial with reason", 480, 500, "20 IMEI rejected by OABI", enums.ShipmentControlVerdictApproved, ""},
		{"partial without reason", 480, 500, "  ", "", ShipmentControlErrCommentRequired},
		{"rejected with reason", 0, 500, "OABI rejected the batch", enums.ShipmentControlVerdictRejected, ""},
		{"rejected without reason", 0, 500, "", "", ShipmentControlErrCommentRequired},
		{"above requested", 501, 500, "x", "", ShipmentControlErrRegisteredAboveRequested},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sc := closedControl(tc.registered, tc.requested, tc.comment)
			err := ApplyShipmentControlVerdict(sc)
			if tc.wantErr != "" {
				if !utils.IsValidationError(err) || err.Error() != tc.wantErr {
					t.Fatalf("err = %v, want validation error %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sc.Verdict != tc.want {
				t.Fatalf("verdict = %q, want %q", sc.Verdict, tc.want)
			}
		})
	}
}

// A rejected control never carries OABI data, whatever the client sent.
func TestApplyShipmentControlVerdictRejectedClearsOabi(t *testing.T) {
	sc := closedControl(0, 500, "rejected")
	if err := ApplyShipmentControlVerdict(sc); err != nil {
		t.Fatal(err)
	}
	if sc.OabiCertificate != "" || sc.OabiCertificateUrl != "" || sc.OabiCertificateNumber != "" {
		t.Fatalf("oabi fields not cleared: %+v", sc)
	}
}

func TestApplyShipmentControlVerdictApprovedKeepsOabi(t *testing.T) {
	sc := closedControl(500, 500, "")
	if err := ApplyShipmentControlVerdict(sc); err != nil {
		t.Fatal(err)
	}
	if sc.OabiCertificateUrl == "" {
		t.Fatal("approved control lost its OABI certificate")
	}
}

// Open controls have no verdict, even if the client sends one.
func TestApplyShipmentControlVerdictOpenClearsVerdict(t *testing.T) {
	sc := &models.ShipmentControl{CurrentPhase: enums.ShipmentControlPhaseUnderRevision,
		RegisteredImeiCount: 0, ImeiQuantity: 10, Verdict: enums.ShipmentControlVerdictRejected}
	if err := ApplyShipmentControlVerdict(sc); err != nil {
		t.Fatal(err)
	}
	if sc.Verdict != "" {
		t.Fatalf("verdict = %q on an open control", sc.Verdict)
	}
}

func TestEnsureShipmentControlNotRejected(t *testing.T) {
	err := EnsureShipmentControlNotRejected(&models.ShipmentControl{Verdict: enums.ShipmentControlVerdictRejected})
	if !utils.IsValidationError(err) || err.Error() != ShipmentControlErrRejectedClosed {
		t.Fatalf("rejected control: err = %v", err)
	}
	for _, sc := range []*models.ShipmentControl{
		nil,
		{},
		{Verdict: enums.ShipmentControlVerdictApproved},
	} {
		if err := EnsureShipmentControlNotRejected(sc); err != nil {
			t.Fatalf("control %+v should stay editable: %v", sc, err)
		}
	}
}

func TestEnrichShipmentControlVerdict(t *testing.T) {
	cases := []struct {
		name       string
		item       responses.ShipmentControlExpanded
		verdict    string
		statusView string
	}{
		{"historical closed", responses.ShipmentControlExpanded{CurrentPhase: 3, Status: 1},
			enums.ShipmentControlVerdictApproved, "Approved"},
		{"rejected", responses.ShipmentControlExpanded{CurrentPhase: 3, Status: 1, Verdict: "rejected"},
			enums.ShipmentControlVerdictRejected, "Rejected"},
		{"open", responses.ShipmentControlExpanded{CurrentPhase: 2, Status: 0}, "", "Ongoing"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := tc.item
			EnrichShipmentControlVerdict(&item)
			if item.Verdict != tc.verdict || item.StatusView != tc.statusView {
				t.Fatalf("verdict=%q status_view=%q, want %q / %q",
					item.Verdict, item.StatusView, tc.verdict, tc.statusView)
			}
		})
	}
}

func TestShipmentControlExportStatus(t *testing.T) {
	if got := ShipmentControlExportStatus("approved", 1, 500, 500); got != "Approved" {
		t.Errorf("complete = %q", got)
	}
	if got := ShipmentControlExportStatus("approved", 1, 480, 500); got != "Approved (partial)" {
		t.Errorf("partial = %q", got)
	}
	if got := ShipmentControlExportStatus("rejected", 1, 0, 500); got != "Rejected" {
		t.Errorf("rejected = %q", got)
	}
	if got := ShipmentControlExportStatus("", 0, 0, 500); got != "Ongoing" {
		t.Errorf("open = %q", got)
	}
}

// ---- closing emails ----

func closingNotify(registered int, verdict, comment string) *request.ShipmentControlNotify {
	day := func(d int) time.Time { return time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC) }
	return &request.ShipmentControlNotify{
		CompanyName:             "Xiaomi Chile SpA",
		CountryName:             "Chile",
		Client:                  "Entel",
		ReferenceID:             "SC-2026-0142",
		Validation:              "V-88",
		ReworkNumber:            "RW-3",
		ImeiQuantity:            500,
		ImeiFileUrl:             "https://example.com/imei.xlsx",
		RegisteredImeiCount:     registered,
		SubtelCertificateNumber: "SUB-12345",
		MultibandCertificateUrl: "https://example.com/multibanda.pdf",
		SoftwareVersion:         "OS2.0.210.0.VMRMIXM",
		HardwareVersion:         "P1",
		OsVersionView:           "Android 15",
		CurrentPhase:            enums.ShipmentControlPhaseCompleted,
		PlanningDate:            day(3),
		ValidationStartDate:     day(4),
		ValidationEndDate:       day(10),
		UnderRevisionStartDate:  day(10),
		UnderRevisionEndDate:    day(20),
		CompletedDate:           day(20),
		OabiCertificate:         "REG-OABI-7781",
		OabiCertificateUrl:      "https://example.com/oabi.pdf",
		Comment:                 comment,
		Verdict:                 verdict,
	}
}

func renderClosingEmail(t *testing.T, notify *request.ShipmentControlNotify) (string, string) {
	t.Helper()
	msg, subject := GetShipmentControlNotificationMessageAndSubject(
		shipmentControlEmailComplete, "XIAOMI", "Redmi Note 15 Pro+ 5G", notify.SoftwareVersion, notify.Verdict)
	data := BuildShipmentControlPhaseEmailData(notify, "XIAOMI", "2510ERA8BG", "Redmi Note 15 Pro+ 5G",
		"Android", "Camilo Espinoza", msg, shipmentControlEmailComplete)
	data.LogoDataURI = resolveBrandedEmailLogo(utils.LBOneTrackLogoPNG)
	html, err := RenderShipmentControlPhaseEmailHTML(data, templatePath(t, utils.TEMPLATE_SHIPMENT_CONTROL_PHASE_PATH))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return subject, string(html)
}

func TestClosingEmailApprovedComplete(t *testing.T) {
	subject, body := renderClosingEmail(t, closingNotify(500, "approved", ""))
	if subject != "Subject: Control Shipment Approved – XIAOMI Redmi Note 15 Pro+ 5G SW version OS2.0.210.0.VMRMIXM" {
		t.Errorf("subject = %q", subject)
	}
	for _, want := range []string{"APPROVED — 500 of 500 IMEI registered", "Download OABI Certificate", "Excel file", "#10b981"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "This request is closed") {
		t.Error("approved email shows the closed notice")
	}
}

func TestClosingEmailApprovedPartial(t *testing.T) {
	_, body := renderClosingEmail(t, closingNotify(480, "approved", "20 IMEI were not accepted by OABI"))
	for _, want := range []string{
		"APPROVED — partial registration: 480 of 500 IMEI registered",
		"20 IMEI were not accepted by OABI", "Download OABI Certificate", "#f59e0b",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestClosingEmailRejected(t *testing.T) {
	subject, body := renderClosingEmail(t, closingNotify(0, "rejected", "OABI rejected the whole batch"))
	if !strings.HasPrefix(subject, "Subject: Control Shipment Rejected – XIAOMI") {
		t.Errorf("subject = %q", subject)
	}
	for _, want := range []string{
		"REJECTED — 0 of 500 IMEI registered. No OABI certificate was issued.",
		"OABI rejected the whole batch",
		"This request is closed. To register these devices, please create a new Shipment Control request in the platform.",
		"#dc2626", "Excel file",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "Download OABI Certificate") {
		t.Error("rejected email links an OABI certificate")
	}
}

// TestWriteShipmentControlClosingEmailSamples writes the three closing emails
// for visual review when SHIPMENT_CONTROL_EMAIL_SAMPLES_DIR is set.
func TestWriteShipmentControlClosingEmailSamples(t *testing.T) {
	dir := os.Getenv("SHIPMENT_CONTROL_EMAIL_SAMPLES_DIR")
	if dir == "" {
		t.Skip("set SHIPMENT_CONTROL_EMAIL_SAMPLES_DIR to write the samples")
	}
	samples := map[string]*request.ShipmentControlNotify{
		"shipment_control_approved.html":         closingNotify(500, "approved", ""),
		"shipment_control_approved_partial.html": closingNotify(480, "approved", "20 IMEI were not accepted by OABI (duplicated in the national registry)."),
		"shipment_control_rejected.html":         closingNotify(0, "rejected", "OABI rejected the whole batch: the IMEI file does not match the declared model."),
	}
	for name, notify := range samples {
		subject, body := renderClosingEmail(t, notify)
		html := "<!-- " + subject + " -->\r\n" + body
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(html), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", path)
	}
}
