package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Homologation struct {
	ID               primitive.ObjectID `bson:"_id" json:"_id"`
	Company          primitive.ObjectID `bson:"company" json:"company"`
	Device           primitive.ObjectID `bson:"device" json:"device"`
	Country          primitive.ObjectID `bson:"country" json:"country"`
	TestPlan         primitive.ObjectID `bson:"test_plan" json:"test_plan"`
	SoftwareVersion  string             `json:"software_version"`
	HardwareVersion  string             `json:"hardware_version"`
	Type             int                `json:"type"`
	TestCategories   []TestCategory     `bson:"test_categories" json:"test_categories"`
	CurrentPhase     int                `json:"phase"`
	PlanningDate     time.Time          `json:"planning_date"`
	SampleStartDate  time.Time          `json:"sample_start_date"`
	SampleEndDate    time.Time          `json:"sample_end_date"`
	TestStartDate    time.Time          `json:"test_start_date"`
	TestEndDate      time.Time          `json:"test_end_date"`
	UnderStartDate   time.Time          `json:"under_start_date"`
	UnderEndDate     time.Time          `json:"under_end_date"`
	CompletedDate    time.Time          `json:"completed_date"`
	IsCustomTestPlan bool               `json:"is_custom_test_plan"`
	Status           int                `bson:"status"`
}
