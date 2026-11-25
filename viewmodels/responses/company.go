package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company is the stored company. Email and LogoUrl are still read for
// internal use (they are legacy data) but no longer exposed in JSON.
type Company struct {
	ID                     primitive.ObjectID `bson:"_id" json:"_id"`
	Email                  string             `bson:"email" json:"-"`
	Name                   string             `bson:"name" json:"name"`
	Address                string             `bson:"address" json:"address"`
	LogoUrl                string             `bson:"logo_url" json:"-"`
	ClientID               string             `bson:"client_id" json:"client_id"`
	Rut                    string             `bson:"rut" json:"rut"`
	RazonSocial            string             `bson:"razon_social" json:"razon_social"`
	AllowSmrWithoutTesting bool               `bson:"allow_smr_without_testing" json:"allow_smr_without_testing"`
}

// CompanyListItem is one row of GET /company: the company plus the
// administration data of the Companies screen.
type CompanyListItem struct {
	Company
	// UsersCount counts users whose company or clients include this company.
	UsersCount int `json:"users_count"`
	// OabiCertificatesIssued counts this company's shipment controls with an
	// OABI certificate issued.
	OabiCertificatesIssued int `json:"oabi_certificates_issued"`
	// ClientIDLocked: client_id can no longer change (certificates issued).
	ClientIDLocked bool `json:"client_id_locked"`
	// OabiMissing lists the missing fields among client_id, rut,
	// razon_social, address. Empty means complete.
	OabiMissing []string `json:"oabi_missing"`
}

type CompanyClientIDAvailability struct {
	Available bool   `json:"available"`
	UsedBy    string `json:"used_by"`
}

type CompanyUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	ProfileName string `json:"profile_name"`
	IsInternal  bool   `json:"is_internal"`
}
