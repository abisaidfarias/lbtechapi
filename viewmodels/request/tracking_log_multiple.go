package request

// Company model
type TrackingLogMultiple struct {
	DeviceTrackingIds []string    `bson:"device_trakings" json:"device_trakings" binding:"required"`
	TrackingLog       TrackingLog `bson:"tracking_log" json:"tracking_log" binding:"required"`
}
