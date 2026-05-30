package functions

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

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
