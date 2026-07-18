package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Company struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Email       string             `bson:"email"`
	Name        string             `bson:"name,omitempty"`
	Address     string             `bson:"address"`
	LogoUrl     string             `bson:"logo_url"`
	ClientID    string             `bson:"client_id,omitempty"`
	Rut         string             `bson:"rut,omitempty"`
	RazonSocial string             `bson:"razon_social,omitempty"`
}
