package request

import "time"

type ShipmentControl struct {
	MultibandaID            string         `json:"multibanda_id" binding:"required"`
	CurrentPhase            int            `json:"current_phase"`
	Status                  int            `json:"status"`
	PlanningDate            time.Time      `json:"planning_date"`
	ValidationStartDate     time.Time      `json:"validation_start_date"`
	ValidationEndDate       time.Time      `json:"validation_end_date"`
	UnderRevisionStartDate  time.Time      `json:"under_revision_start_date"`
	UnderRevisionEndDate    time.Time      `json:"under_revision_end_date"`
	CompletedDate           time.Time      `json:"completed_date"`
	ImeiQuantity            int            `json:"imei_quantity" binding:"required"`
	ImeiFileUrl             string         `json:"imei_file_url" binding:"required"`
	RegisteredImeiCount     int            `json:"registered_imei_count"`
	ReferenceID             FlexibleString `json:"reference_id"`
	Validation              FlexibleString `json:"validation"`
	ReworkNumber            string         `json:"rework_number"`
	OabiCertificate         string         `json:"oabi_certificate"`
	Client                  string         `json:"client"`
	Country                 string         `json:"country"`
	SubtelCertificateUrl    string         `json:"subtel_certificate_url"`
	SubtelCertificateNumber string         `json:"subtel_certificate_number"`
	OabiCertificateUrl      string         `json:"oabi_certificate_url"`
	OabiCertificateNumber   string         `json:"oabi_certificate_number"`
	Comment                 string         `json:"comment"`
}
