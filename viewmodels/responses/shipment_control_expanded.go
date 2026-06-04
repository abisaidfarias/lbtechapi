package responses



import (

	"time"



	"go.mongodb.org/mongo-driver/bson/primitive"

)



type ShipmentControlCompanySummary struct {

	ID   primitive.ObjectID `bson:"_id" json:"_id"`

	Name string             `bson:"name" json:"name"`

}



type ShipmentControlCountrySummary struct {

	ID   primitive.ObjectID `bson:"_id" json:"_id"`

	Name string             `bson:"name" json:"name"`

}



type ShipmentControlMultibandaSummary struct {

	ID                primitive.ObjectID `json:"_id"`

	CertificateNumber string             `json:"certificate_number"`

	SoftwareVersion   string             `json:"software_version"`

	HardwareVersion   string             `json:"hardware_version"`

	OsVersion         string             `json:"os_version"`

	OsVersionView     string             `json:"os_version_view"`

}



type ShipmentControlDeviceSummary struct {

	CommercialModel string `json:"commercial_model"`

	TechnicalModel  string `json:"technical_model"`

	Brand           string `json:"brand"`

	PlatformOs      string `json:"platform_os"`

}



type ShipmentControlExpanded struct {

	ID                      primitive.ObjectID              `bson:"_id" json:"_id"`

	Multibanda              ShipmentControlMultibandaSummary `json:"multibanda"`

	Device                  ShipmentControlDeviceSummary    `json:"device"`

	Company                 ShipmentControlCompanySummary   `bson:"company" json:"company"`

	Country                 ShipmentControlCountrySummary   `bson:"country" json:"country"`

	CurrentPhase            int                             `bson:"current_phase" json:"current_phase"`

	Status                  int                             `bson:"status" json:"status"`

	PlanningDate            *time.Time                      `bson:"planning_date" json:"planning_date"`

	ValidationStartDate     *time.Time                      `bson:"validation_start_date" json:"validation_start_date"`

	ValidationEndDate       *time.Time                      `bson:"validation_end_date" json:"validation_end_date"`

	UnderRevisionStartDate  *time.Time                      `bson:"under_revision_start_date" json:"under_revision_start_date"`

	UnderRevisionEndDate    *time.Time                      `bson:"under_revision_end_date" json:"under_revision_end_date"`

	CompletedDate           *time.Time                      `bson:"completed_date" json:"completed_date"`

	ImeiQuantity            int                             `bson:"imei_quantity" json:"imei_quantity"`

	ImeiFileUrl             string                          `bson:"imei_file_url" json:"imei_file_url"`

	RegisteredImeiCount     int                             `bson:"registered_imei_count" json:"registered_imei_count"`

	ReworkNumber            string                          `bson:"rework_number" json:"rework_number"`

	OabiCertificate         string                          `bson:"oabi_certificate" json:"oabi_certificate"`

	Client                  string                          `bson:"client" json:"client"`

	SubtelCertificateUrl    string                          `bson:"subtel_certificate_url" json:"subtel_certificate_url"`

	SubtelCertificateNumber string                          `bson:"subtel_certificate_number" json:"subtel_certificate_number"`

	OabiCertificateUrl      string                          `bson:"oabi_certificate_url" json:"oabi_certificate_url"`

	OabiCertificateNumber   string                          `bson:"oabi_certificate_number" json:"oabi_certificate_number"`

	Comment                 string                          `bson:"comment" json:"comment"`

	RequestDelete           bool                            `bson:"request_delete" json:"request_delete"`

	CreatedDate             *time.Time                      `bson:"created_date" json:"created_date"`

}


