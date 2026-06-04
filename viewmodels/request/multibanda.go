package request

import "time"

type Multibanda struct {
	Company           string     `json:"company" binding:"required"`
	Device            string     `json:"device" binding:"required"`
	Brand             string     `json:"brand" binding:"required"`
	SoftwareVersion   string     `json:"software_version" binding:"required"`
	HardwareVersion   string     `json:"hardware_version"`
	OsVersion         string     `json:"os_version" binding:"required"`
	OsVersionView     string     `json:"os_version_view"`
	Type              string     `json:"type" binding:"required"`
	EvaluationType    []string   `json:"evaluation_type" binding:"required"`
	CurrentPhase      int        `json:"current_phase"`
	PlanningDate      time.Time  `json:"planning_date" binding:"required"`
	SampleStartDate         *time.Time `json:"sample_start_date"`
	SampleEndDate           *time.Time `json:"sample_end_date"`
	TestReportUrl           string     `json:"test_report_url"`
	MultibandCertificateUrl string     `json:"multiband_certificate_url"`
	Comment                 string     `json:"comment"`
	NeedReflash             bool       `json:"need_reflash"`
	CommentsReflash         string     `json:"comments_reflash"`
	IsInternalProject       bool       `json:"is_internal_project"`
}
