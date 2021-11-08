package request

import (
	"time"
)

// Company model
type Homologation struct {
	Company           string    `bson:"company" json:"company" binding:"required"`
	Device            string    `bson:"device" json:"device" binding:"required"`
	Country           string    `bson:"country" json:"country" binding:"required"`
	Brand             string    `bson:"brand" json:"brand"`
	SoftwareVersion   string    `json:"software_version"`
	HardwareVersion   string    `json:"hardware_version"`
	Type              int       `json:"type"`
	TestCategories    []string  `bson:"test_categories" json:"test_categories"`
	CurrentPhase      int       `json:"current_phase"`
	PlanningDate      time.Time `json:"planning_date"`
	SampleStartDate   time.Time `json:"sample_start_date"`
	SampleEndDate     time.Time `json:"sample_end_date"`
	TestStartDate     time.Time `json:"test_start_date"`
	TestEndDate       time.Time `json:"test_end_date"`
	UnderStartDate    time.Time `json:"under_start_date"`
	UnderEndDate      time.Time `json:"under_end_date"`
	CompletedDate     time.Time `json:"completed_date"`
	IsCustomTestPlan  bool      `json:"is_custom_test_plan"`
	TestPlan          string    `bson:"test_plan" json:"test_plan"`
	Status            int       `bson:"status" json:"status"`
	IsInternalProject bool      `bson:"is_internal_project" json:"is_internal_project"`
	OsVersion         string    `bson:"os_version" json:"os_version" binding:"required"`
	DocumentUrl       string    `bson:"document_url" json:"document_url"`
}
