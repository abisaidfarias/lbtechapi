package responses

import (
	"time"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
)

// SmrLastReport is the last generated report of another request with the
// same device technical model.
type SmrLastReport struct {
	ID           string     `json:"id"`
	MultibandaID string     `json:"multibanda_id"`
	Type         string     `json:"type"`
	ReportDate   *time.Time `json:"report_date"`
	StampCode    string     `json:"stamp_code"`
}

// SmrValidationLogEntry is one audited action.
type SmrValidationLogEntry struct {
	Action          string    `json:"action"` // approve | reject
	At              time.Time `json:"at"`
	ByName          string    `json:"by_name"`
	Comment         string    `json:"comment"`
	FMRadioOriginal *bool     `json:"fm_radio_original"`
	FMRadioFinal    *bool     `json:"fm_radio_final"`
	StampOriginal   string    `json:"stamp_original"`
	StampFinal      string    `json:"stamp_final"`
	SourceReportID  string    `json:"source_report_id"`
}

// SmrValidation is GET /multibanda/:id/smr-validation.
type SmrValidation struct {
	Status     string         `json:"status"` // pending | approved | rejected
	LastReport *SmrLastReport `json:"last_report"`
	// FMRadio is the device's communication_radio (null when not set).
	FMRadio *bool `json:"fm_radio"`
	// StampCode comes from last_report ("" when there is none).
	StampCode     string                  `json:"stamp_code"`
	Stamps        []enums.ReportStamp     `json:"stamps"`
	DecidedAt     *time.Time              `json:"decided_at"`
	DecidedByName string                  `json:"decided_by_name"`
	Comment       string                  `json:"comment"`
	Log           []SmrValidationLogEntry `json:"log"`
}
