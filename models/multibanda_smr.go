package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Phase states: "N/A" is display only; a phase not performed is "skipped".
const (
	PhaseStatePending = "pending"
	PhaseStateDone    = "done"
	PhaseStateSkipped = "skipped"
)

// MultibandaPhaseStates holds the state of each multibanda phase.
type MultibandaPhaseStates struct {
	Planning        string `bson:"planning" json:"planning"`
	SampleReception string `bson:"sample_reception" json:"sample_reception"`
	Testing         string `bson:"testing" json:"testing"`
	UnderEvaluation string `bson:"under_evaluation" json:"under_evaluation"`
	Complete        string `bson:"complete" json:"complete"`
}

// SMR validation statuses.
const (
	SmrValidationPending  = "pending"
	SmrValidationApproved = "approved"
	SmrValidationRejected = "rejected"
)

// SmrValidation is the internal validation of an SMR without testing (FM
// Radio + stamp). Decision fields hold the last decision; Log keeps every
// action for audit.
type SmrValidation struct {
	Status        string               `bson:"status"`
	DecidedAt     *time.Time           `bson:"decided_at,omitempty"`
	DecidedBy     primitive.ObjectID   `bson:"decided_by,omitempty"`
	DecidedByName string               `bson:"decided_by_name,omitempty"`
	Comment       string               `bson:"comment,omitempty"`
	Log           []SmrValidationEntry `bson:"log,omitempty"`
}

// SmrValidationEntry is one audited approve/reject action.
type SmrValidationEntry struct {
	Action           string             `bson:"action"` // approve | reject
	At               time.Time          `bson:"at"`
	By               primitive.ObjectID `bson:"by"`
	ByName           string             `bson:"by_name"`
	Comment          string             `bson:"comment,omitempty"`
	FMRadioOriginal  *bool              `bson:"fm_radio_original"`
	FMRadioFinal     *bool              `bson:"fm_radio_final"`
	StampOriginal    string             `bson:"stamp_original"`
	StampFinal       string             `bson:"stamp_final"`
	SourceReportID   primitive.ObjectID `bson:"source_report_id,omitempty"`
	SourceMultibanda primitive.ObjectID `bson:"source_multibanda_id,omitempty"`
}
