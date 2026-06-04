package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Multibanda struct {
	ID                primitive.ObjectID `bson:"_id" json:"_id"`
	Company           primitive.ObjectID `bson:"company" json:"company"`
	Device            primitive.ObjectID `bson:"device" json:"device"`
	Brand             primitive.ObjectID `bson:"brand" json:"brand"`
	SoftwareVersion   string             `bson:"software_version" json:"software_version"`
	HardwareVersion   string             `bson:"hardware_version" json:"hardware_version"`
	OsVersion         string             `bson:"os_version" json:"os_version"`
	OsVersionView     string             `bson:"os_version_view" json:"os_version_view"`
	Type              string             `bson:"type" json:"type"`
	EvaluationType    []string           `bson:"evaluation_type" json:"evaluation_type"`
	CurrentPhase      int                `bson:"current_phase" json:"current_phase"`
	PlanningDate      time.Time          `bson:"planning_date" json:"planning_date"`
	SampleStartDate         *time.Time         `bson:"sample_start_date,omitempty" json:"sample_start_date"`
	SampleEndDate           *time.Time         `bson:"sample_end_date,omitempty" json:"sample_end_date"`
	TestReportUrl           string             `bson:"test_report_url" json:"test_report_url"`
	MultibandCertificateUrl string             `bson:"multiband_certificate_url" json:"multiband_certificate_url"`
	IsInternalProject       bool               `bson:"is_internal_project" json:"is_internal_project"`
	CreatedDate       time.Time          `bson:"created_date" json:"created_date"`
}
