package request

// Company model
type Notification struct {
	Company            string              `bson:"company" json:"company" binding:"required"`
	NotificationEmails []NotificationEmail `bson:"notification_emails,omitempty" json:"notification_emails"`
	Name               string              `bson:"name,omitempty"`
	LastName           string              `bson:"lastName,omitempty"`
}
