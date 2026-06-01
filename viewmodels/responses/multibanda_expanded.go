package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MultibandaExpanded struct {
	ID                primitive.ObjectID `bson:"_id" json:"_id"`
	Company           Company            `bson:"company" json:"company"`
	Device            Device             `bson:"device" json:"device"`
	Brand             Brand              `bson:"brand" json:"brand"`
	SoftwareVersion   string             `bson:"software_version" json:"software_version"`
	HardwareVersion   string             `bson:"hardware_version" json:"hardware_version"`
	OsVersion         string             `bson:"os_version" json:"os_version"`
	OsVersionView     string             `bson:"os_version_view" json:"os_version_view"`
	Type              string             `bson:"type" json:"type"`
	EvaluationType    []string           `bson:"evaluation_type" json:"evaluation_type"`
	CurrentPhase      int                `bson:"current_phase" json:"current_phase"`
	PlanningDate      *time.Time         `bson:"planning_date" json:"planning_date"`
	SampleStartDate   *time.Time         `bson:"sample_start_date" json:"sample_start_date"`
	SampleEndDate     *time.Time         `bson:"sample_end_date" json:"sample_end_date"`
	TestStartDate     *time.Time         `bson:"test_start_date" json:"test_start_date"`
	TestEndDate       *time.Time         `bson:"test_end_date" json:"test_end_date"`
	UnderStartDate    *time.Time         `bson:"under_start_date" json:"under_start_date"`
	UnderEndDate      *time.Time         `bson:"under_end_date" json:"under_end_date"`
	CompletedDate     *time.Time         `bson:"completed_date" json:"completed_date"`
	Status                  int                `bson:"status" json:"status"`
	DashBoardPhase          int                `bson:"dashboard_phase" json:"dashboard_phase"`
	TestReportUrl           string             `bson:"test_report_url" json:"test_report_url"`
	MultibandCertificateUrl string             `bson:"multiband_certificate_url" json:"multiband_certificate_url"`
	CertificateNumber       string             `bson:"certificate_number" json:"certificate_number"`
	Comment                 string             `bson:"comment" json:"comment"`
	NeedReflash             bool               `bson:"need_reflash" json:"need_reflash"`
	CommentsReflash         string             `bson:"comments_reflash" json:"comments_reflash"`
	RequestDelete           bool               `bson:"request_delete" json:"request_delete"`
	IsInternalProject       bool               `bson:"is_internal_project" json:"is_internal_project"`
	CreatedDate       *time.Time         `bson:"created_date" json:"created_date"`
}
