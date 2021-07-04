package responses

// Company model
type Tracking struct {
	Brand           string           `bson:"brand" json:"brand"`
	Model           string           `bson:"model" json:"model"`
	ImageUrl        string           `bson:"image_url" json:"image_url"`
	DeviceTrackings []DeviceTracking `bson:"device_trackings" json:"device_trackings"`
}
