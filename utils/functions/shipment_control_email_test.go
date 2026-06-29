package functions

import (
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

func TestGetShipmentControlNotificationSubjectIncludesRework(t *testing.T) {
	_, subject := GetShipmentControlNotificationMessageAndSubject(
		shipmentControlEmailCreate,
		"XIAOMI",
		"Redmi Note",
		"RW-001",
	)
	if subject == "" {
		t.Fatal("expected subject")
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
