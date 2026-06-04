package request

// Company model
type TrackingLogMultiple struct {
	DeviceTrackings []string    `bson:"device_trackings" json:"device_trackings" binding:"required"`
	TrackingLog     TrackingLog `bson:"tracking_log" json:"tracking_log" binding:"required"`
	// WithDelivery when true: perform move but do not send notification email.
	WithDelivery bool `bson:"with_delivery" json:"with_delivery"`
}
