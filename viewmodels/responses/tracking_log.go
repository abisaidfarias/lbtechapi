package responses

import (
	"time"
)

// Company model
type TrackingLog struct {
	TrackingID          string    `bson:"tracking_id" json:"tracking_id"`
	Country             Country   `bson:"country" json:"country"`
	Location            Location  `bson:"location" json:"location"`
	InternalResponsible User      `bson:"internal_responsible" json:"internal_responsible"`
	Person              Person    `bson:"person" json:"person"`
	Comment             string    `bson:"comment" json:"comment"`
	DocumentUrl         string    `bson:"document_url" json:"document_url"`
	TrackingDate        time.Time `bson:"tracking_date" json:"tracking_date"`
	ExternalResponsible string    `bson:"external_responsible" json:"external_responsible"`
	ProcessTypes        []string  `bson:"process_types" json:"process_types"`
}
