package mapping

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/models"
)

// Reproduces the bug reported after shipping reference_id/validation: the
// phase-change request payload does not carry these fields (it only carries
// dates/status/etc.), so without preservation a phase transition would $set
// them back to empty and silently erase what was entered at creation.
func TestPreserveShipmentControlReferenceAndValidationKeepsExistingOnPhaseChange(t *testing.T) {
	shipment := &models.ShipmentControl{} // built from a phase-change request: no reference_id/validation
	existing := &models.ShipmentControl{ReferenceID: "75962", Validation: "11"}

	PreserveShipmentControlReferenceAndValidation(shipment, existing)

	if shipment.ReferenceID != "75962" {
		t.Fatalf("reference_id: got %q, want %q", shipment.ReferenceID, "75962")
	}
	if shipment.Validation != "11" {
		t.Fatalf("validation: got %q, want %q", shipment.Validation, "11")
	}
}

func TestPreserveShipmentControlReferenceAndValidationKeepsIncomingWhenPresent(t *testing.T) {
	shipment := &models.ShipmentControl{ReferenceID: "NEW-1", Validation: "9"}
	existing := &models.ShipmentControl{ReferenceID: "OLD-1", Validation: "1"}

	PreserveShipmentControlReferenceAndValidation(shipment, existing)

	if shipment.ReferenceID != "NEW-1" {
		t.Fatalf("reference_id: got %q, want incoming value preserved", shipment.ReferenceID)
	}
	if shipment.Validation != "9" {
		t.Fatalf("validation: got %q, want incoming value preserved", shipment.Validation)
	}
}

func TestPreserveShipmentControlReferenceAndValidationNilExisting(t *testing.T) {
	shipment := &models.ShipmentControl{}

	// Must not panic when there is no existing document (e.g. defensive call).
	PreserveShipmentControlReferenceAndValidation(shipment, nil)

	if shipment.ReferenceID != "" || shipment.Validation != "" {
		t.Fatalf("expected fields to stay empty, got reference_id=%q validation=%q", shipment.ReferenceID, shipment.Validation)
	}
}
