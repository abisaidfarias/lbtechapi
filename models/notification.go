package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCase model
type Notification struct {
	ID                 primitive.ObjectID  `bson:"_id,omitempty"`
	Company            primitive.ObjectID  `bson:"company,omitempty"`
	NotificationEmails []NotificationEmail `bson:"notification_emails,omitempty" json:"notification_emails"`
}
