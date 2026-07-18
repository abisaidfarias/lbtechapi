package functions

import (
	"strings"
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func TestFormatShipmentControlEmailDateUsesDashForEmpty(t *testing.T) {
	if got := FormatShipmentControlEmailDate(time.Time{}); got != "-" {
		t.Fatalf("expected dash for zero date, got %q", got)
	}
}

func TestFormatMultibandaEmailDateUsesDashForEmpty(t *testing.T) {
	if got := FormatMultibandaEmailDate(time.Time{}); got != "-" {
		t.Fatalf("expected dash for zero date, got %q", got)
	}
}

func TestResolveShipmentControlPhaseEmailKindCreate(t *testing.T) {
	kind := ResolveShipmentControlPhaseEmailKind(utils.CREATE, nil, &request.ShipmentControlNotify{})
	if kind != shipmentControlEmailCreate {
		t.Fatalf("expected create, got %s", kind)
	}
}

func TestResolveShipmentControlPhaseEmailKindValidationStart(t *testing.T) {
	updated := &request.ShipmentControlNotify{
		CurrentPhase:        enums.ShipmentControlPhaseValidation,
		ValidationStartDate: time.Now(),
	}
	kind := ResolveShipmentControlPhaseEmailKind(utils.PHASE, nil, updated)
	if kind != shipmentControlEmailValidationStart {
		t.Fatalf("expected validation start, got %s", kind)
	}
}

func TestResolveShipmentControlPhaseEmailKindCompleteOnUnderRevisionEnd(t *testing.T) {
	updated := &request.ShipmentControlNotify{
		UnderRevisionEndDate: time.Now(),
	}
	kind := ResolveShipmentControlPhaseEmailKind(utils.PHASE, nil, updated)
	if kind != shipmentControlEmailComplete {
		t.Fatalf("expected complete, got %s", kind)
	}
}

func TestGetShipmentControlNotificationSubjectIncludesSoftwareVersion(t *testing.T) {
	_, subject := GetShipmentControlNotificationMessageAndSubject(
		shipmentControlEmailCreate,
		"XIAOMI",
		"Redmi Note 15 5G",
		"OS2.0.210.0.VMRMIXM",
	)
	if subject == "" {
		t.Fatal("expected subject")
	}
	if !strings.Contains(subject, "SW version OS2.0.210.0.VMRMIXM") {
		t.Fatalf("expected SW version in subject, got %q", subject)
	}
}

func TestBuildShipmentControlPhaseEmailDataIncludesCertificateNumbers(t *testing.T) {
	notify := &request.ShipmentControlNotify{
		SubtelCertificateNumber: "SUB-100",
		OabiCertificate:         "OABI-200",
	}
	data := BuildShipmentControlPhaseEmailData(
		notify, "XIAOMI", "2510ERA8BG", "Redmi Note 15 Pro+ 5G", "Android", "Camilo Espinoza", "completed", shipmentControlEmailComplete,
	)
	if data.MultibandaCertificateNumber != "SUB-100" {
		t.Fatalf("multibanda cert: got %q", data.MultibandaCertificateNumber)
	}
	if data.OabiCertificate != "OABI-200" {
		t.Fatalf("oabi cert: got %q", data.OabiCertificate)
	}
}

func TestBuildShipmentControlPhaseEmailDataIncludesImeiFileURLOnlyOnCreate(t *testing.T) {
	notify := &request.ShipmentControlNotify{
		ImeiQuantity: 10,
		ImeiFileUrl:  "https://example.com/imei.xlsx",
	}

	createData := BuildShipmentControlPhaseEmailData(
		notify, "XIAOMI", "TM", "CM", "Android", "User", "msg", shipmentControlEmailCreate,
	)
	if createData.ImeiFileURL != notify.ImeiFileUrl {
		t.Fatalf("expected imei file url on create email, got %q", createData.ImeiFileURL)
	}

	phaseData := BuildShipmentControlPhaseEmailData(
		notify, "XIAOMI", "TM", "CM", "Android", "User", "msg", shipmentControlEmailValidationStart,
	)
	if phaseData.ImeiFileURL != "" {
		t.Fatalf("expected empty imei file url on phase email, got %q", phaseData.ImeiFileURL)
	}
}

func TestGetShipmentControlDeleteNotificationSubjectRequestDeleteInternal(t *testing.T) {
	_, subject := GetShipmentControlDeleteNotificationMessageAndSubject(
		ShipmentControlEmailRequestDeleteInternal,
		"XIAOMI",
		"Redmi Note",
		"OS2.0.210.0.VMRMIXM",
		"Xiaomi Chile",
	)
	if subject == "" {
		t.Fatal("expected subject")
	}
}

func TestGetShipmentControlDeleteNotificationSubjectDeleted(t *testing.T) {
	msg, subject := GetShipmentControlDeleteNotificationMessageAndSubject(
		ShipmentControlEmailDeleted,
		"XIAOMI",
		"Redmi Note",
		"OS2.0.210.0.VMRMIXM",
		"Xiaomi Chile",
	)
	if msg == "" || subject == "" {
		t.Fatalf("expected message and subject, got %q / %q", msg, subject)
	}
}

func TestFormatEmailCommentMultilinePreservesExplicitNewlines(t *testing.T) {
	got := formatEmailCommentMultiline("line one\nline two")
	if got != "line one\nline two" {
		t.Fatalf("expected preserved newlines, got %q", got)
	}
}

func TestFormatEmailCommentMultilineSplitsSpaceSeparatedNumbers(t *testing.T) {
	got := formatEmailCommentMultiline("10 IMEI no subidos 32442 23234 242342")
	want := "10 IMEI no subidos\n32442\n23234\n242342"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFormatEmailCommentMultilineLeavesPlainTextUntouched(t *testing.T) {
	got := formatEmailCommentMultiline("Waiting for client feedback")
	if got != "Waiting for client feedback" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderShipmentControlPhaseEmailHTML(t *testing.T) {
	html, err := RenderShipmentControlPhaseEmailHTML(ShipmentControlPhaseEmailData{
		MainMessage: "Test notification",
		Year:        2026,
	}, utils.TEMPLATE_SHIPMENT_CONTROL_PHASE_PATH)
	if err != nil {
		t.Fatalf("render shipment control email: %v", err)
	}
	if len(html) == 0 {
		t.Fatal("expected non-empty html")
	}
}

func TestShipmentControlNotifiesExternalRecipients(t *testing.T) {
	if !ShipmentControlNotifiesExternalRecipients(utils.CREATE, "") {
		t.Fatal("expected external recipients on create")
	}
	if !ShipmentControlNotifiesExternalRecipients(utils.PHASE, shipmentControlEmailComplete) {
		t.Fatal("expected external recipients on complete")
	}
	if ShipmentControlNotifiesExternalRecipients(utils.PHASE, shipmentControlEmailValidationStart) {
		t.Fatal("did not expect external recipients on validation start")
	}
}

func TestMultibandaNotifiesExternalRecipients(t *testing.T) {
	if !MultibandaNotifiesExternalRecipients(utils.CREATE, 0) {
		t.Fatal("expected external recipients on create")
	}
	if !MultibandaNotifiesExternalRecipients(utils.PHASE, enums.HomologationPhase_value["COMPLETE"]) {
		t.Fatal("expected external recipients on complete phase")
	}
	if MultibandaNotifiesExternalRecipients(utils.PHASE, enums.HomologationPhase_value["TEST"]) {
		t.Fatal("did not expect external recipients on intermediate phase")
	}
}
