package responses

import (
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MultibandaExpanded struct {
	ID              primitive.ObjectID `bson:"_id" json:"_id"`
	Company         Company            `bson:"company" json:"company"`
	Device          Device             `bson:"device" json:"device"`
	Brand           Brand              `bson:"brand" json:"brand"`
	SoftwareVersion string             `bson:"software_version" json:"software_version"`
	HardwareVersion string             `bson:"hardware_version" json:"hardware_version"`
	Carrier         string             `bson:"carrier" json:"carrier"`
	BaseBandVersion string             `bson:"base_band_version" json:"base_band_version"`
	OsVersion       string             `bson:"os_version" json:"os_version"`
	OsVersionView   string             `bson:"os_version_view" json:"os_version_view"`
	Type            string             `bson:"type" json:"type"`
	EvaluationType  []string           `bson:"evaluation_type" json:"evaluation_type"`
	CurrentPhase    int                `bson:"current_phase" json:"current_phase"`
	PlanningDate    *time.Time         `bson:"planning_date" json:"planning_date"`
	SampleStartDate *time.Time         `bson:"sample_start_date" json:"sample_start_date"`
	SampleEndDate   *time.Time         `bson:"sample_end_date" json:"sample_end_date"`
	TestStartDate   *time.Time         `bson:"test_start_date" json:"test_start_date"`
	TestEndDate     *time.Time         `bson:"test_end_date" json:"test_end_date"`
	UnderStartDate  *time.Time         `bson:"under_start_date" json:"under_start_date"`
	UnderEndDate    *time.Time         `bson:"under_end_date" json:"under_end_date"`
	CompletedDate   *time.Time         `bson:"completed_date" json:"completed_date"`
	Status          int                `bson:"status" json:"status"`
	DashBoardPhase  int                `bson:"dashboard_phase" json:"dashboard_phase"`
	TestReportUrl   string             `bson:"test_report_url" json:"test_report_url"`
	// ReportStatus is the automatic report's state: "draft", "generated" or
	// empty when the report screen was never opened. Independent of
	// TestReportUrl, which holds the manually uploaded report of the old flow.
	ReportStatus            string     `bson:"report_status" json:"report_status"`
	// ReportDate is when the automatic report was first generated (the moment
	// the "report available" email goes out); null until then.
	ReportDate              *time.Time `bson:"report_date" json:"report_date"`
	MultibandCertificateUrl string     `bson:"multiband_certificate_url" json:"multiband_certificate_url"`
	AdbInfoUrl              string     `bson:"adb_info_url" json:"adb_info_url"`
	AdbInfoOriginalFileName string     `bson:"adb_info_original_file_name" json:"adb_info_original_file_name"`
	BandValidationUrl       string     `bson:"band_validation_url" json:"band_validation_url"`
	BandValidationOriginalFileName string `bson:"band_validation_original_file_name" json:"band_validation_original_file_name"`
	SubtelCertificateNumber string     `bson:"subtel_certificate_number" json:"subtel_certificate_number"`
	Comment                 string     `bson:"comment" json:"comment"`
	NeedReflash             bool       `bson:"need_reflash" json:"need_reflash"`
	CommentsReflash         string     `bson:"comments_reflash" json:"comments_reflash"`
	RequestDelete           bool       `bson:"request_delete" json:"request_delete"`
	IsInternalProject       bool       `bson:"is_internal_project" json:"is_internal_project"`
	CreatedDate             *time.Time `bson:"created_date" json:"created_date"`
	RequesterName           string     `bson:"requester_name" json:"requester_name"`

	// RequiresTesting: null unless SMR without testing applies to the request.
	RequiresTesting *bool `bson:"requires_testing" json:"requires_testing"`
	// PhaseStates: pending | done | skipped per phase.
	PhaseStates *models.MultibandaPhaseStates `bson:"phase_states" json:"phase_states"`
	// SmrValidation: only for SMR without testing; null otherwise.
	SmrValidation *MultibandaSmrValidation `bson:"smr_validation" json:"smr_validation"`
}

// MultibandaSmrValidation is the stored SMR validation. The list exposes
// status and the decision; comment and log come from GET smr-validation.
type MultibandaSmrValidation struct {
	Status        string                      `bson:"status" json:"status"`
	DecidedAt     *time.Time                  `bson:"decided_at" json:"decided_at"`
	DecidedByName string                      `bson:"decided_by_name" json:"decided_by_name"`
	Comment       string                      `bson:"comment" json:"-"`
	Log           []models.SmrValidationEntry `bson:"log" json:"-"`
}
