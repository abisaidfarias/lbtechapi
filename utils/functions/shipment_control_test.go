package functions

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGroupAvailableMultibandasGroupsByDevice(t *testing.T) {
	deviceID := primitive.NewObjectID()
	company := responses.Company{ID: primitive.NewObjectID(), Name: "Xiaomi"}

	multibandas := []*responses.MultibandaExpanded{
		{
			ID:              primitive.NewObjectID(),
			SoftwareVersion: "SW 1.0",
			Device:          responses.Device{ID: deviceID, CommercialModel: "Note 50"},
			Brand:           responses.Brand{Name: "XIAOMI"},
		},
		{
			ID:              primitive.NewObjectID(),
			SoftwareVersion: "SW 2.0",
			Device:          responses.Device{ID: deviceID, CommercialModel: "Note 50"},
			Brand:           responses.Brand{Name: "XIAOMI"},
		},
	}

	result := GroupAvailableMultibandas(company, multibandas)
	if len(result.Devices) != 1 {
		t.Fatalf("expected 1 device group, got %d", len(result.Devices))
	}
	if len(result.Devices[0].Options) != 2 {
		t.Fatalf("expected 2 software options, got %d", len(result.Devices[0].Options))
	}
}

func TestApplyShipmentControlPhaseDateRulesSetsValidationStartFromPlanningDate(t *testing.T) {
	planningDate := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	existing := &models.ShipmentControl{
		PlanningDate: planningDate,
	}
	shipmentControl := &models.ShipmentControl{
		CurrentPhase: enums.ShipmentControlPhaseValidation,
	}

	ApplyShipmentControlPhaseDateRules(shipmentControl, existing)

	if !shipmentControl.ValidationStartDate.Equal(planningDate) {
		t.Fatalf("expected validation_start_date %v, got %v", planningDate, shipmentControl.ValidationStartDate)
	}
}

func TestApplyShipmentControlPhaseDateRulesSetsUnderRevisionStartFromValidationEnd(t *testing.T) {
	validationEnd := time.Date(2026, 5, 29, 6, 0, 0, 0, time.UTC)
	shipmentControl := &models.ShipmentControl{
		CurrentPhase:     enums.ShipmentControlPhaseUnderRevision,
		ValidationEndDate: validationEnd,
	}

	ApplyShipmentControlPhaseDateRules(shipmentControl, nil)

	if !shipmentControl.UnderRevisionStartDate.Equal(validationEnd) {
		t.Fatalf("expected under_revision_start_date %v, got %v", validationEnd, shipmentControl.UnderRevisionStartDate)
	}
}

func TestApplyShipmentControlPhaseDateRulesSetsCompletedDateFromUnderRevisionEnd(t *testing.T) {
	underRevisionEnd := time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)
	shipmentControl := &models.ShipmentControl{
		CurrentPhase:         enums.ShipmentControlPhaseCompleted,
		UnderRevisionEndDate: underRevisionEnd,
	}

	ApplyShipmentControlPhaseDateRules(shipmentControl, nil)

	if !shipmentControl.CompletedDate.Equal(underRevisionEnd) {
		t.Fatalf("expected completed_date %v, got %v", underRevisionEnd, shipmentControl.CompletedDate)
	}
}

func TestApplyShipmentControlStatusRulesSetsCompletedWhenUnderRevisionEndExists(t *testing.T) {
	underRevisionEnd := time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC)
	shipmentControl := &models.ShipmentControl{
		CurrentPhase:         enums.ShipmentControlPhaseUnderRevision,
		UnderRevisionEndDate: underRevisionEnd,
	}

	ApplyShipmentControlStatusRules(shipmentControl, nil)

	if shipmentControl.CurrentPhase != enums.ShipmentControlPhaseCompleted {
		t.Fatalf("expected completed phase, got %d", shipmentControl.CurrentPhase)
	}
	if shipmentControl.Status != enums.ShipmentControlStatusCompleted {
		t.Fatalf("expected completed status, got %d", shipmentControl.Status)
	}
}

func TestApplyShipmentControlStatusRulesKeepsOngoingWithoutUnderRevisionEnd(t *testing.T) {
	shipmentControl := &models.ShipmentControl{
		CurrentPhase: enums.ShipmentControlPhaseValidation,
	}

	ApplyShipmentControlStatusRules(shipmentControl, nil)

	if shipmentControl.Status != enums.ShipmentControlStatusInProgress {
		t.Fatalf("expected ongoing status, got %d", shipmentControl.Status)
	}
}

func TestUserHasClientAccessExternalOnlyOwnCompany(t *testing.T) {
	companyID := primitive.NewObjectID()
	user := &responses.User{
		IsInternal: false,
		Company:    companyID,
	}

	if !UserHasClientAccess(user, companyID) {
		t.Fatal("expected external user to access own company")
	}
	if UserHasClientAccess(user, primitive.NewObjectID()) {
		t.Fatal("expected external user to be denied for other company")
	}
}

func TestUserHasClientAccessInternalWithoutClientsCanAccessAnyCompany(t *testing.T) {
	user := &responses.User{
		IsInternal: true,
		Clients:    nil,
	}

	if !UserHasClientAccess(user, primitive.NewObjectID()) {
		t.Fatal("expected internal user without client restrictions to access any company")
	}
}
