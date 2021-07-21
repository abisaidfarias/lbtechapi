package request

// Company model
type TrackingLogMultiple struct {
	DeviceTrackings []string    `bson:"device_trackings" json:"device_trackings" binding:"required"`
	TrackingLog     TrackingLog `bson:"tracking_log" json:"tracking_log" binding:"required"`
}
