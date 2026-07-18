package models

import "time"

const (
	OabiCertificateStatusGenerating = "generating"
	OabiCertificateStatusReady      = "ready"
	OabiCertificateStatusFailed     = "failed"
)

type OabiCertificateInputSnapshot struct {
	RegistroOABI        string `bson:"registro_oabi" json:"registro_oabi"`
	RegisteredImeiCount int    `bson:"registered_imei_count" json:"registered_imei_count"`
}

type OabiCertificateState struct {
	Status        string                       `bson:"status" json:"status"`
	URL           string                       `bson:"url" json:"url"`
	ControlNumber string                       `bson:"control_number" json:"control_number"`
	InputSnapshot OabiCertificateInputSnapshot `bson:"input_snapshot" json:"input_snapshot"`
	StartedAt     time.Time                    `bson:"started_at" json:"started_at"`
	GeneratedAt   time.Time                    `bson:"generated_at" json:"generated_at"`
	Attempts      int                          `bson:"attempts" json:"attempts"`
	Error         string                       `bson:"error" json:"error"`
}
