package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type TrackingLog struct {
	Country             primitive.ObjectID `bson:"country,omitempty"`
	Location            primitive.ObjectID `bson:"location,omitempty"`
	InternalResponsible primitive.ObjectID `bson:"internal_responsible,omitempty"`
	ExternalResponsible string             `bson:"external_responsible,omitempty"`
	Comment             string             `bson:"comment"`
	DocumentUrl         string             `bson:"document_url"`
	TrackingDate        time.Time          `bson:"tracking_date"`
}
