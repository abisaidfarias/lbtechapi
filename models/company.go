package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Company struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	// Email and LogoUrl are no longer edited (email lives on each user). They
	// are kept with omitempty so the $set of an update never wipes stored data.
	Email       string `bson:"email,omitempty"`
	Name        string `bson:"name,omitempty"`
	Address     string `bson:"address"`
	LogoUrl     string `bson:"logo_url,omitempty"`
	ClientID    string `bson:"client_id,omitempty"`
	Rut         string `bson:"rut,omitempty"`
	RazonSocial string `bson:"razon_social,omitempty"`
	// AllowSmrWithoutTesting lets the company request SMR without testing.
	AllowSmrWithoutTesting bool `bson:"allow_smr_without_testing"`
}
