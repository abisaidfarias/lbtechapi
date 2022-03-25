package responses

import "go.mongodb.org/mongo-driver/bson/primitive"

// Company model
type Tracking struct {
	ID              primitive.ObjectID       `bson:"_id"`
	Brand           string                   `bson:"brand" json:"brand"`
	Model           string                   `bson:"model" json:"model"`
	TecnicalModel   string                   `bson:"tecnical_model" json:"tecnical_model"`
	ImageUrl        string                   `bson:"image_url" json:"image_url"`
	DeviceTrackings []DeviceTrackingExpanded `bson:"device_trackings" json:"device_trackings"`
}
