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
	ShipmentControlPhaseCompleted:     "Completed",
}

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
