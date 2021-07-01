package responses

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company model
type DeviceTracking struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Company     Company            `bson:"company,omitempty"`
	Device      Device             `bson:"device,omitempty"`
	Imei        string             `bson:"imei,omitempty"`
	TrackingLog TrackingLog        `bson:"tracking_log"`
}
