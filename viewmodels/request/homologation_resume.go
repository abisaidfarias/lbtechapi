package request

import (
	"time"
)

// Company model
type HomologationResume struct {
	SoftwareVersion   string    `json:"software_version"`
	HardwareVersion   string    `json:"hardware_version"`
	CurrentPhase      int       `json:"current_phase"`
	PlanningDate      time.Time `json:"planning_date"`
	SampleStartDate   time.Time `json:"sample_start_date"`
	SampleEndDate     time.Time `json:"sample_end_date"`
	TestStartDate     time.Time `json:"test_start_date"`
	TestEndDate       time.Time `json:"test_end_date"`
	UnderStartDate    time.Time `json:"under_start_date"`
	UnderEndDate      time.Time `json:"under_end_date"`
	CompletedDate     time.Time `json:"completed_date"`
	Status            int       `bson:"status" json:"status"`
	IsInternalProject bool      `bson:"is_internal_project,omitempty"`
	OsVersion         string    `bson:"os_version" json:"os_version"`
	DocumentUrl       string    `bson:"document_url" json:"document_url"`
	ResultUrl         string    `bson:"result_url" json:"result_url"`
	Carrier           string    `bson:"carrier" json:"carrier"`
	TestingType       string    `bson:"testing_type" json:"testing_type"`
}
