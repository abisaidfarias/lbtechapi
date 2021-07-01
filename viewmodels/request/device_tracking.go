package request

// Company model
type DeviceTracking struct {
	Company     string      `bson:"company" json:"company" binding:"required"`
	Device      string      `bson:"device" json:"device" binding:"required"`
	Imeis       []string    `bson:"imeis" json:"imeis" binding:"required"`
	TrackingLog TrackingLog `bson:"tracking_log" json:"tracking_log" binding:"required"`
}

