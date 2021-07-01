package request

import (
	"time"
)

// Company model
type TrackingLog struct {
	Country             string    `bson:"country" json:"country" binding:"required"`
	Location            string    `bson:"location" json:"location" binding:"required"`
	InternalResponsible string    `bson:"internal_responsible" json:"internal_responsible" binding:"required"`
	ExternalResponsible string    `bson:"external_responsible" json:"external_responsible" binding:"required"`
	Comment             string    `bson:"comment" json:"comment"`
	DocumentUrl         string    `bson:"document_url" json:"document_url"`
	TrackingDate        time.Time `bson:"tracking_date" json:"tracking_date"`
}
