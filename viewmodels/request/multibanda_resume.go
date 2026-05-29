package request

import "time"

type MultibandaResume struct {
	SoftwareVersion   string    `json:"software_version"`
	HardwareVersion   string    `json:"hardware_version"`
	OsVersion         string    `json:"os_version"`
	OsVersionView     string    `json:"os_version_view"`
	CurrentPhase      int       `json:"current_phase"`
	PlanningDate      time.Time `json:"planning_date"`
	SampleStartDate   time.Time `json:"sample_start_date"`
	SampleEndDate     time.Time `json:"sample_end_date"`
	TestStartDate     time.Time `json:"test_start_date"`
	TestEndDate       time.Time `json:"test_end_date"`
	UnderStartDate    time.Time `json:"under_start_date"`
	UnderEndDate      time.Time `json:"under_end_date"`
	CompletedDate           time.Time `json:"completed_date"`
	Status                  int       `json:"status"`
	CertificateNumber       string    `json:"certificate_number"`
	TestReportUrl           string    `json:"test_report_url"`
	MultibandCertificateUrl string    `json:"multiband_certificate_url"`
	IsInternalProject       bool      `json:"is_internal_project,omitempty"`
}

type MultibandaNotify struct {
	Company                 string
	CompanyName             string
	Device                  string
	Brand                   string
	SoftwareVersion         string
	HardwareVersion         string
	OsVersion               string
	OsVersionView           string
	Type                    string
	EvaluationType          []string
	CurrentPhase            int
	PlanningDate            time.Time
	SampleStartDate         time.Time
	SampleEndDate           time.Time
	TestStartDate           time.Time
	TestEndDate             time.Time
	UnderStartDate          time.Time
	UnderEndDate            time.Time
	CompletedDate           time.Time
	Status                  int
	CertificateNumber       string
	TestReportUrl           string
	MultibandCertificateUrl string
	IsInternalProject       bool
}
