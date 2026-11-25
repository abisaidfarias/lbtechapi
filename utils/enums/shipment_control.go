package enums

const (
	ShipmentControlExternalCountryName = "Chile"

	ShipmentControlPhasePlanning      = 0
	ShipmentControlPhaseValidation    = 1
	ShipmentControlPhaseUnderRevision = 2
	ShipmentControlPhaseCompleted     = 3
)

const (
	ShipmentControlStatusInProgress = 0
	ShipmentControlStatusCompleted  = 1
)

var ShipmentControlPhaseLabels = map[int]string{
	ShipmentControlPhasePlanning:      "Planning",
	ShipmentControlPhaseValidation:    "Validation",
	ShipmentControlPhaseUnderRevision: "Under Revision",
	// Phase 3 closes the control with a verdict (approved / rejected).
	ShipmentControlPhaseCompleted: "Decision",
}

// Verdict of a closed (phase 3) shipment control. It is computed on the
// server from registered_imei_count vs imei_quantity; "partial" is not stored
// but derived as approved && registered < requested.
const (
	ShipmentControlVerdictApproved = "approved"
	ShipmentControlVerdictRejected = "rejected"
)

const (
	ShipmentControlStatusViewApproved        = "Approved"
	ShipmentControlStatusViewApprovedPartial = "Approved (partial)"
	ShipmentControlStatusViewRejected        = "Rejected"
)

var ShipmentControlStatusLabels = map[int]string{
	ShipmentControlStatusInProgress: "Ongoing",
	ShipmentControlStatusCompleted:  "Completed",
}

func ValidateShipmentControlPhase(phase int) bool {
	_, ok := ShipmentControlPhaseLabels[phase]
	return ok
}

func ValidateShipmentControlStatus(status int) bool {
	_, ok := ShipmentControlStatusLabels[status]
	return ok
}
