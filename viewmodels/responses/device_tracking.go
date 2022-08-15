package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type DeviceTracking struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Company      primitive.ObjectID `bson:"company,omitempty"`
	Device       primitive.ObjectID `bson:"device,omitempty"`
	Imei         string             `bson:"imei,omitempty"`
	TrackingLogs []TrackingLog      `bson:"tracking_logs"`
}
