package request

import (
	"time"
)

// Company model
type Homologation struct {
	Company          string    `bson:"company" json:"company" binding:"required"`
	Device           string    `bson:"device" json:"device" binding:"required"`
	Country          string    `bson:"country" json:"country" binding:"required"`
	SoftwareVersion  string    `json:"software_version"`
	HardwareVersion  string    `json:"hardware_version"`
	Type             int       `json:"type"`
	TestCategories   []string  `bson:"test_categories" json:"test_categories" binding:"required"`
	CurrentPhase     int       `json:"current_phase"`
	PlanningDate     time.Time `json:"planning_date"`
	SampleStartDate  time.Time `json:"sample_start_date"`
	SampleEndDate    time.Time `json:"sample_end_date"`
	TestStartDate    time.Time `json:"test_start_date"`
	TestEndDate      time.Time `json:"test_end_date"`
	UnderStartDate   time.Time `json:"under_start_date"`
	UnderEndDate     time.Time `json:"under_end_date"`
	CompletedDate    time.Time `json:"completed_date"`
	IsCustomTestPlan bool      `json:"is_custom_test_plan"`
}
