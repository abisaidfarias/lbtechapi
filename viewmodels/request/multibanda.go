package request

import "time"

type Multibanda struct {
	Company                 string     `json:"company" binding:"required"`
	Device                  string     `json:"device" binding:"required"`
	Brand                   string     `json:"brand" binding:"required"`
	SoftwareVersion         string     `json:"software_version" binding:"required"`
	HardwareVersion         string     `json:"hardware_version"`
	OsVersion               string     `json:"os_version" binding:"required"`
	OsVersionView           string     `json:"os_version_view"`
	Type                    string     `json:"type" binding:"required"`
	EvaluationType          []string   `json:"evaluation_type" binding:"required"`
	CurrentPhase            int        `json:"current_phase"`
	Status                  int        `json:"status"`
	PlanningDate            time.Time  `json:"planning_date" binding:"required"`
	SampleStartDate         *time.Time `json:"sample_start_date"`
	SampleEndDate           *time.Time `json:"sample_end_date"`
	TestStartDate           time.Time  `json:"test_start_date"`
	TestEndDate             time.Time  `json:"test_end_date"`
	UnderStartDate          time.Time  `json:"under_start_date"`
	UnderEndDate            time.Time  `json:"under_end_date"`
	CompletedDate           time.Time  `json:"completed_date"`
	TestReportUrl           string     `json:"test_report_url"`
	MultibandCertificateUrl string     `json:"multiband_certificate_url"`
	SubtelCertificateNumber string     `json:"subtel_certificate_number"`
	Comment                 string     `json:"comment"`
	NeedReflash             bool       `json:"need_reflash"`
	CommentsReflash         string     `json:"comments_reflash"`
	IsInternalProject       bool       `json:"is_internal_project"`
}
