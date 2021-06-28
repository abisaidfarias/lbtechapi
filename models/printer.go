package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type Printer struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	Modelo           string             `bson:"model"`
	Serial           string             `bson:"serial"`
	Pages            string             `bson:"pages"`
	Location         string             `bson:"location"`
	MaxTonner        string             `bson:"maxtoner,omitempty"`
	RemToner         string             `bson:"remtoner,omitempty"`
	PercentageTonner string             `bson:"percentage_tonner,omitempty"`
	SNconsumible     string             `bson:"SNconsumible,omitempty"`
	PNconsumible     string             `bson:"PNconsumible,omitempty"`
	Details          []Detail           `bson:"details" json:"details"`
	CreatedDate      time.Time          `bson:"created_date"`
}
