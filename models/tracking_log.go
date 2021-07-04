package models

import (
	"time"
)

// Company model
type TrackingLog struct {
	Country             Country   `bson:"country,omitempty"`
	Location            Location  `bson:"location,omitempty"`
	InternalResponsible User      `bson:"internal_responsible,omitempty"`
	ExternalResponsible string    `bson:"external_responsible,omitempty"`
	Comment             string    `bson:"comment"`
	DocumentUrl         string    `bson:"document_url"`
	TrackingDate        time.Time `bson:"tracking_date"`
}
