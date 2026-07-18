package models



import (

	"time"



	"go.mongodb.org/mongo-driver/bson/primitive"

)



type ShipmentControl struct {

	ID                      primitive.ObjectID `bson:"_id,omitempty"`

	Multibanda              primitive.ObjectID `bson:"multibanda"`

	Company                 primitive.ObjectID `bson:"company"`

	Country                 primitive.ObjectID `bson:"country"`

	CurrentPhase            int                `bson:"current_phase"`

	Status                  int                `bson:"status"`

	PlanningDate            time.Time          `bson:"planning_date"`

	ValidationStartDate     time.Time          `bson:"validation_start_date"`

	ValidationEndDate       time.Time          `bson:"validation_end_date"`

	UnderRevisionStartDate  time.Time          `bson:"under_revision_start_date"`

	UnderRevisionEndDate    time.Time          `bson:"under_revision_end_date"`

	CompletedDate           time.Time          `bson:"completed_date"`

	ImeiQuantity            int                `bson:"imei_quantity"`

	ImeiFileUrl             string             `bson:"imei_file_url"`

	RegisteredImeiCount     int                `bson:"registered_imei_count"`

	ReworkNumber            string             `bson:"rework_number"`

	OabiCertificate         string             `bson:"oabi_certificate"`

	Client                  string             `bson:"client"`

	SubtelCertificateUrl    string             `bson:"subtel_certificate_url"`

	SubtelCertificateNumber string             `bson:"subtel_certificate_number"`

	OabiCertificateUrl      string             `bson:"oabi_certificate_url"`

	OabiCertificateNumber   string             `bson:"oabi_certificate_number"`

	OabiCertificateState    *OabiCertificateState `bson:"oabi_certificate_state,omitempty"`

	Comment                 string             `bson:"comment"`

	RequestDelete           bool               `bson:"request_delete"`

	CreatedDate             time.Time          `bson:"created_date"`

}


