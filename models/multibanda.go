package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Multibanda struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	Company           primitive.ObjectID `bson:"company,omitempty"`
	Device            primitive.ObjectID `bson:"device,omitempty"`
	Brand             primitive.ObjectID `bson:"brand,omitempty"`
	SoftwareVersion   string             `bson:"software_version,omitempty"`
	HardwareVersion   string             `bson:"hardware_version,omitempty"`
	OsVersion         string             `bson:"os_version,omitempty"`
	OsVersionView     string             `bson:"os_version_view,omitempty"`
	Type              string             `bson:"type,omitempty"`
	EvaluationType    []string           `bson:"evaluation_type,omitempty"`
	CurrentPhase      int                `bson:"current_phase"`
	PlanningDate      time.Time          `bson:"planning_date"`
	SampleStartDate   *time.Time         `bson:"sample_start_date,omitempty"`
	SampleEndDate     *time.Time         `bson:"sample_end_date,omitempty"`
	TestStartDate     time.Time          `bson:"test_start_date"`
	TestEndDate       time.Time          `bson:"test_end_date"`
	UnderStartDate    time.Time          `bson:"under_start_date"`
	UnderEndDate      time.Time          `bson:"under_end_date"`
	CompletedDate     time.Time          `bson:"completed_date"`
	Status                  int                `bson:"status"`
	DashBoardPhase        int                `bson:"dashboard_phase"`
	TestReportUrl         string             `bson:"test_report_url,omitempty"`
	MultibandCertificateUrl string           `bson:"multiband_certificate_url,omitempty"`
	CertificateNumber     string             `bson:"certificate_number,omitempty"`
	Comment               string             `bson:"comment,omitempty"`
	IsInternalProject     bool               `bson:"is_internal_project"`
	CreatedDate       time.Time          `bson:"created_date"`
}
