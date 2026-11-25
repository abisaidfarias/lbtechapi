package functions

import (
	"strings"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

// Error messages of the verdict rules (all returned as 400 validation errors).
const (
	ShipmentControlErrRegisteredAboveRequested = "registered_imei_count cannot be greater than imei_quantity"
	ShipmentControlErrCommentRequired          = "comment is required when the shipment control is rejected or partially registered"
	ShipmentControlErrRejectedClosed           = "this shipment control was rejected and is closed; it cannot be modified or reopened. Create a new Shipment Control request instead"
)

// ShipmentControlVerdictFor computes the verdict from the registered and
// requested IMEI counts: 0 registered is rejected, anything else approved
// (partial when below the requested quantity).
func ShipmentControlVerdictFor(registered, requested int) (string, error) {
	if registered > requested {
		return "", utils.NewValidationError(ShipmentControlErrRegisteredAboveRequested)
	}
	if registered <= 0 {
		return enums.ShipmentControlVerdictRejected, nil
	}
	return enums.ShipmentControlVerdictApproved, nil
}

// IsShipmentControlPartial reports an approved control that registered fewer
// IMEI than requested.
func IsShipmentControlPartial(verdict string, registered, requested int) bool {
	return verdict == enums.ShipmentControlVerdictApproved && registered < requested
}

// EnsureShipmentControlNotRejected blocks any change to a rejected control:
// rejection is final.
func EnsureShipmentControlNotRejected(existing *models.ShipmentControl) error {
	if existing != nil && existing.Verdict == enums.ShipmentControlVerdictRejected {
		return utils.NewValidationError(ShipmentControlErrRejectedClosed)
	}
	return nil
}

// ApplyShipmentControlVerdict sets the verdict once the phase rules have run.
// The client never sends it: it is derived from the counts. Open controls
// carry no verdict. A rejected control gets no OABI certificate, so any OABI
// fields sent are cleared.
func ApplyShipmentControlVerdict(shipmentControl *models.ShipmentControl) error {
	if shipmentControl == nil {
		return nil
	}
	if shipmentControl.CurrentPhase != enums.ShipmentControlPhaseCompleted {
		shipmentControl.Verdict = ""
		return nil
	}

	verdict, err := ShipmentControlVerdictFor(shipmentControl.RegisteredImeiCount, shipmentControl.ImeiQuantity)
	if err != nil {
		return err
	}
	partial := IsShipmentControlPartial(verdict, shipmentControl.RegisteredImeiCount, shipmentControl.ImeiQuantity)
	if (verdict == enums.ShipmentControlVerdictRejected || partial) &&
		strings.TrimSpace(shipmentControl.Comment) == "" {
		return utils.NewValidationError(ShipmentControlErrCommentRequired)
	}

	if verdict == enums.ShipmentControlVerdictRejected {
		shipmentControl.OabiCertificate = ""
		shipmentControl.OabiCertificateUrl = ""
		shipmentControl.OabiCertificateNumber = ""
	}
	shipmentControl.Verdict = verdict
	return nil
}

// EffectiveShipmentControlVerdict is the verdict exposed on read: closed
// controls from before verdicts existed count as approved.
func EffectiveShipmentControlVerdict(verdict string, currentPhase int) string {
	if verdict == "" && currentPhase == enums.ShipmentControlPhaseCompleted {
		return enums.ShipmentControlVerdictApproved
	}
	return verdict
}

// ShipmentControlStatusView is the status shown to users: the verdict for
// closed controls, the status label otherwise.
func ShipmentControlStatusView(verdict string, status int) string {
	switch verdict {
	case enums.ShipmentControlVerdictApproved:
		return enums.ShipmentControlStatusViewApproved
	case enums.ShipmentControlVerdictRejected:
		return enums.ShipmentControlStatusViewRejected
	}
	return enums.ShipmentControlStatusLabels[status]
}

// ShipmentControlExportStatus is the Excel status: like the status view, but
// partial approvals are called out.
func ShipmentControlExportStatus(verdict string, status, registered, requested int) string {
	if IsShipmentControlPartial(verdict, registered, requested) {
		return enums.ShipmentControlStatusViewApprovedPartial
	}
	return ShipmentControlStatusView(verdict, status)
}

// EnrichShipmentControlVerdict fills verdict and status_view on a list item.
func EnrichShipmentControlVerdict(item *responses.ShipmentControlExpanded) {
	if item == nil {
		return
	}
	item.Verdict = EffectiveShipmentControlVerdict(item.Verdict, item.CurrentPhase)
	item.StatusView = ShipmentControlStatusView(item.Verdict, item.Status)
}
