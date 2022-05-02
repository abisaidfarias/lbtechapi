package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type Notification struct {
	ID                 primitive.ObjectID  `bson:"_id,omitempty"`
	Company            primitive.ObjectID  `bson:"company" json:"company"`
	NotificationEmails []NotificationEmail `bson:"notification_emails,omitempty" json:"notification_emails"`
	Name               string              `bson:"name,omitempty"`
	LastName           string              `bson:"lastName,omitempty"`
}
