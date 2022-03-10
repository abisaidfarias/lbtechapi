package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type HomologationExpanded struct {
	ID                 primitive.ObjectID `bson:"_id" json:"_id"`
	Company            Company            `bson:"company" json:"company"`
	Device             Device             `bson:"device" json:"device"`
	Country            Country            `bson:"country" json:"country"`
	TestPlan           TestPlan           `bson:"test_plan" json:"test_plan"`
	Brand              Brand              `bson:"brand" json:"brand"`
	SoftwareVersion    string             `bson:"software_version" json:"software_version"`
	HardwareVersion    string             `bson:"hardware_version" json:"hardware_version"`
	Type               int                `bson:"type" json:"type"`
	TestCategories     []TestCategory     `bson:"test_categories" json:"test_categories"`
	CurrentPhase       int                `bson:"current_phase" json:"current_phase"`
	PlanningDate       *time.Time         `bson:"planning_date" json:"planning_date"`
	SampleStartDate    *time.Time         `bson:"sample_start_date" json:"sample_start_date"`
	SampleEndDate      *time.Time         `bson:"sample_end_date" json:"sample_end_date"`
	TestStartDate      *time.Time         `bson:"test_start_date" json:"test_start_date"`
	TestEndDate        *time.Time         `bson:"test_end_date" json:"test_end_date"`
	UnderStartDate     *time.Time         `bson:"under_start_date" json:"under_start_date"`
	UnderEndDate       *time.Time         `bson:"under_end_date" json:"under_end_date"`
	CompletedDate      *time.Time         `bson:"completed_date" json:"completed_date"`
	IsCustomTestPlan   bool               `bson:"is_custom_test_plan" json:"is_custom_test_plan"`
	Status             int                `bson:"status" json:"status"`
	DocumentUrl        string             `bson:"document_url" json:"document_url"`
	OsVersion          string             `bson:"os_version" json:"os_version"`
	IsInternalProject  bool               `bson:"is_internal_project" json:"is_internal_project"`
	ResultUrl          string             `bson:"result_url" json:"result_url"`
	Carrier            string             `bson:"carrier" json:"carrier"`
	TestingType        string             `bson:"testing_type" json:"testing_type"`
	Comment            string             `bson:"comment" json:"comment"`
	ApprovalTypeOption string             `bson:"approval_type_option" json:"approval_type_option"`
	ApprovalType       string             `bson:"approval_type" json:"approval_type"`
	ProjectType        string             `bson:"project_type" json:"project_type"`
	StatusView         string             `bson:"status_view" json:"status_view"`
	OsVersionView      string             `bson:"os_version_view" json:"os_version_view"`
}
