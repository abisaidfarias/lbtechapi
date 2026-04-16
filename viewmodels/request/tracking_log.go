package request

import (
	"time"
)

// Company model
type TrackingLog struct {
	Country             Country    `bson:"country" json:"country" binding:"required"`
	Location            Location   `bson:"location" json:"location" binding:"required"`
	InternalResponsible UserResume `bson:"internal_responsible" json:"internal_responsible" binding:"required"`
	Person              Person     `bson:"person" json:"person" binding:"required"`
	Comment             string     `bson:"comment" json:"comment"`
	DocumentUrl         string     `bson:"document_url" json:"document_url"`
	TrackingDate        time.Time  `bson:"tracking_date" json:"tracking_date"`
	ProcessTypes        []string   `bson:"process_types" json:"process_types"`
}
