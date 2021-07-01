package responses

// Company model
type Tracking struct {
	Brand           string           `bson:"company,omitempty"`
	Model           string           `bson:"device,omitempty"`
	ImageUrl        string           `bson:"image_url" json:"image_url"`
	DeviceTrackings []DeviceTracking `bson:"device_trackings,omitempty"`
}
