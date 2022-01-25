package request

// Company model
type DeviceTracking struct {
	Company      string        `bson:"company" json:"company" binding:"required"`
	Device       string        `bson:"device" json:"device" binding:"required"`
	Imeis        []string      `bson:"imeis" json:"imeis"`
	TrackingLog  TrackingLog   `bson:"tracking_log" json:"tracking_log"`
	Imei         string        `bson:"imei" json:"imei"`
	TrackingLogs []TrackingLog `bson:"tracking_logs" json:"tracking_logs"`
}
