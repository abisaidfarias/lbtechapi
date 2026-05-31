package request

import "time"

type ShipmentControlNotify struct {
	CompanyName             string
	CountryName             string
	Client                  string
	ReworkNumber            string
	ImeiQuantity            int
	ImeiFileUrl             string
	RegisteredImeiCount     int
	CertificateNumber       string
	MultibandCertificateUrl string
	SoftwareVersion         string
	HardwareVersion         string
	OsVersion               string
	OsVersionView           string
	CurrentPhase            int
	PlanningDate            time.Time
	ValidationStartDate     time.Time
	ValidationEndDate       time.Time
	UnderRevisionStartDate  time.Time
	UnderRevisionEndDate    time.Time
	CompletedDate           time.Time
	OabiCertificate         string
	Comment                 string
	SubtelCertificateUrl    string
	OabiCertificateUrl      string
}
