package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type DeviceTracking struct {
	ID           primitive.ObjectID `bson:"_id" json:"_id"`
	Company      Company            `bson:"company" json:"company"`
	Device       DeviceResume       `bson:"device" json:"device"`
	Imei         string             `bson:"imei" json:"imei"`
	TrackingLogs []TrackingLog      `bson:"tracking_logs" json:"tracking_logs"`
}
