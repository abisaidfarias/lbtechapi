package functions

import (
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
)

// Error messages of the SMR without testing flow (all 400).
const (
	SmrErrNotEnabled            = "SMR without testing is not enabled for this request"
	SmrErrRequiresTestingNeeded = "requires_testing is required (true or false) for this SMR"
	SmrErrRequiresTestingLocked = "requires_testing cannot be changed after the SMR validation was decided"
	SmrErrPendingValidation     = "this SMR without testing is pending validation: it cannot advance until it is approved"
	SmrErrRejectedClosed        = "this SMR without testing was rejected and is closed; it cannot be modified or reopened"
	SmrErrNoTestingPhases       = "this SMR does not require testing: it cannot go to Sample reception or Testing"
	SmrErrNotWithoutTesting     = "this request is not an SMR without testing"
	SmrErrNotPending            = "the SMR validation was already decided"
	SmrErrStampRequired         = "the stamp is required to approve"
	SmrErrStampInvalid          = "invalid stamp_code"
	SmrErrFMRadioRequired       = "fm_radio is required (true or false) to approve"
	SmrErrCommentOnCorrection   = "a comment is required when FM Radio or the stamp is corrected"
	SmrErrCommentOnReject       = "a comment is required to reject"
	SmrErrReportPending         = "the report cannot be generated while the SMR without testing is pending validation"
	SmrErrOnlyInPlanning        = "SMR without testing can only be chosen while the request is in Planning"
)

// phase order: planning(0) sample_reception(1) testing(2) under_evaluation(3) complete(4).
var phaseSlots = []func(*models.MultibandaPhaseStates) *string{
	func(p *models.MultibandaPhaseStates) *string { return &p.Planning },
	func(p *models.MultibandaPhaseStates) *string { return &p.SampleReception },
	func(p *models.MultibandaPhaseStates) *string { return &p.Testing },
	func(p *models.MultibandaPhaseStates) *string { return &p.UnderEvaluation },
	func(p *models.MultibandaPhaseStates) *string { return &p.Complete },
}

// ComputePhaseStates derives each phase state from current_phase: before it
// done, at it and after it pending; complete is done once the process is
// complete. A phase already "skipped" in prev stays skipped.
func ComputePhaseStates(currentPhase int, prev *models.MultibandaPhaseStates) *models.MultibandaPhaseStates {
	out := &models.MultibandaPhaseStates{}
	complete := enums.HomologationPhase_value["COMPLETE"]
	for phase, slot := range phaseSlots {
		state := models.PhaseStatePending
		if phase < currentPhase || (phase == complete && currentPhase == complete) {
			state = models.PhaseStateDone
		}
		if prev != nil && *slot(prev) == models.PhaseStateSkipped {
			state = models.PhaseStateSkipped
		}
		*slot(out) = state
	}
	return out
}

// SmrWithoutTestingInitialStates: sample reception and testing are skipped
// from creation, so the front shows N/A from the first moment.
func SmrWithoutTestingInitialStates() *models.MultibandaPhaseStates {
	return &models.MultibandaPhaseStates{
		Planning:        models.PhaseStatePending,
		SampleReception: models.PhaseStateSkipped,
		Testing:         models.PhaseStateSkipped,
		UnderEvaluation: models.PhaseStatePending,
		Complete:        models.PhaseStatePending,
	}
}

// SmrApprovedStates: approved goes straight to Under evaluation.
func SmrApprovedStates() *models.MultibandaPhaseStates {
	return &models.MultibandaPhaseStates{
		Planning:        models.PhaseStateDone,
		SampleReception: models.PhaseStateSkipped,
		Testing:         models.PhaseStateSkipped,
		UnderEvaluation: models.PhaseStatePending,
		Complete:        models.PhaseStatePending,
	}
}

// SmrRejectedStates: a rejection closes the request.
func SmrRejectedStates() *models.MultibandaPhaseStates {
	return &models.MultibandaPhaseStates{
		Planning:        models.PhaseStateDone,
		SampleReception: models.PhaseStateSkipped,
		Testing:         models.PhaseStateSkipped,
		UnderEvaluation: models.PhaseStateSkipped,
		Complete:        models.PhaseStateDone,
	}
}

// RequiresTestingApplies: the question is asked only for SMR + SAE
// multibanda certificate when the company allows SMR without testing.
func RequiresTestingApplies(multibandaType string, evaluationTypes []string, companyAllows bool) bool {
	if !companyAllows || strings.TrimSpace(multibandaType) != enums.MultibandaTypeSMR {
		return false
	}
	for _, e := range evaluationTypes {
		if strings.TrimSpace(e) == enums.MultibandaEvalSAEMultibandaCertificate {
			return true
		}
	}
	return false
}

// NormalizeRequiresTesting validates the sent value and returns what to
// store: the value when the question applies, nil otherwise (true is
// tolerated and dropped there; false is rejected).
func NormalizeRequiresTesting(applies bool, value *bool) (*bool, error) {
	if applies {
		if value == nil {
			return nil, utils.NewValidationError(SmrErrRequiresTestingNeeded)
		}
		v := *value
		return &v, nil
	}
	if value != nil && !*value {
		return nil, utils.NewValidationError(SmrErrNotEnabled)
	}
	return nil, nil
}

// IsSmrWithoutTesting reports requires_testing == false.
func IsSmrWithoutTesting(requiresTesting *bool) bool {
	return requiresTesting != nil && !*requiresTesting
}

// GuardSmrChange blocks Update/PhaseChange on an SMR without testing:
// rejected is closed; pending cannot change phase or status; approved cannot
// go back to Sample reception or Testing.
func GuardSmrChange(validationStatus string, currentPhase, currentStatus, newPhase, newStatus int) error {
	switch validationStatus {
	case models.SmrValidationRejected:
		return utils.NewValidationError(SmrErrRejectedClosed)
	case models.SmrValidationPending:
		if newPhase != currentPhase || newStatus != currentStatus {
			return utils.NewValidationError(SmrErrPendingValidation)
		}
	case models.SmrValidationApproved:
		if newPhase == enums.HomologationPhase_value["SAMPLE_RECEPTION"] ||
			newPhase == enums.HomologationPhase_value["TEST"] {
			return utils.NewValidationError(SmrErrNoTestingPhases)
		}
	}
	return nil
}

// SmrApproveCheck validates an approval. A correction is changing a value
// that already existed (FM Radio not null, stamp not empty); filling an empty
// value is not. Corrections need a comment.
func SmrApproveCheck(originalFM *bool, originalStamp string, fmRadio *bool, stampCode, comment string) error {
	if fmRadio == nil {
		return utils.NewValidationError(SmrErrFMRadioRequired)
	}
	stampCode = strings.TrimSpace(stampCode)
	if stampCode == "" {
		return utils.NewValidationError(SmrErrStampRequired)
	}
	if !enums.IsValidStampCode(stampCode) {
		return utils.NewValidationError(SmrErrStampInvalid)
	}
	fmCorrected := originalFM != nil && *originalFM != *fmRadio
	stampCorrected := strings.TrimSpace(originalStamp) != "" && strings.TrimSpace(originalStamp) != stampCode
	if (fmCorrected || stampCorrected) && strings.TrimSpace(comment) == "" {
		return utils.NewValidationError(SmrErrCommentOnCorrection)
	}
	return nil
}

// SmrKpi counts SMR requests by testing mode and validation outcome, and
// averages the days from creation to the decision of those without
// testing. It needs only type, requires_testing, smr_validation and
// created_date (never the sample/testing dates).
type SmrKpi struct {
	WithTesting       int
	WithoutTesting    int
	Approved          int
	Rejected          int
	Pending           int
	AvgDaysToDecision float64
}

// ComputeSmrKpi applies SmrKpi to a list of requests.
func ComputeSmrKpi(items []SmrKpiInput) SmrKpi {
	var k SmrKpi
	var totalDays float64
	decided := 0
	for _, it := range items {
		if it.Type != enums.MultibandaTypeSMR {
			continue
		}
		if !IsSmrWithoutTesting(it.RequiresTesting) {
			k.WithTesting++
			continue
		}
		k.WithoutTesting++
		switch it.ValidationStatus {
		case models.SmrValidationApproved:
			k.Approved++
		case models.SmrValidationRejected:
			k.Rejected++
		default:
			k.Pending++
			continue
		}
		if it.DecidedAt != nil && !it.CreatedDate.IsZero() {
			totalDays += it.DecidedAt.Sub(it.CreatedDate).Hours() / 24
			decided++
		}
	}
	if decided > 0 {
		k.AvgDaysToDecision = totalDays / float64(decided)
	}
	return k
}

// SmrKpiInput is the data the KPI needs from each request.
type SmrKpiInput struct {
	Type             string
	RequiresTesting  *bool
	ValidationStatus string
	CreatedDate      time.Time
	DecidedAt        *time.Time
}
