package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Homologation struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	Company           primitive.ObjectID `bson:"company,omitempty"`
	Device            primitive.ObjectID `bson:"device,omitempty"`
	Country           primitive.ObjectID `bson:"country,omitempty"`
	Brand             primitive.ObjectID `bson:"brand,omitempty"`
	SoftwareVersion   string             `bson:"software_version,omitempty"`
	HardwareVersion   string             `bson:"hardware_version,omitempty"`
	Type              int                `bson:"type"`
	TestResults       []TestResult       `bson:"test_results,omitempty"`
	CurrentPhase      int                `bson:"current_phase"`
	PlanningDate      time.Time          `bson:"planning_date"`
	SampleStartDate   time.Time          `bson:"sample_start_date"`
	SampleEndDate     time.Time          `bson:"sample_end_date"`
	TestStartDate     time.Time          `bson:"test_start_date"`
	TestEndDate       time.Time          `bson:"test_end_date"`
	UnderStartDate    time.Time          `bson:"under_start_date"`
	UnderEndDate      time.Time          `bson:"under_end_date"`
	CompletedDate     time.Time          `bson:"completed_date"`
	IsCustomTestPlan  bool               `bson:"is_custom_test_plan"`
	CreatedDate       time.Time          `bson:"created_date"`
	Status            int                `bson:"status"`
	TestPlan          primitive.ObjectID `bson:"test_plan,omitempty"`
	IsInternalProject bool               `bson:"is_internal_project"`
	OsVersion         string             `bson:"os_version,omitempty"`
	DocumentUrl       string             `bson:"document_url"`
	ResultUrl         string             `bson:"result_url"`
	Carrier           string             `bson:"carrier"`
	TestingType       string             `bson:"testing_type"`
	Comment           string             `bson:"comment"`
}
