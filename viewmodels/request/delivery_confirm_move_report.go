package request

import "time"

// MoveDeliveryReceiver is the person who physically receives devices at the local (signature from canvas as data URL).
type MoveDeliveryReceiver struct {
	FullName            string    `json:"full_name" binding:"required"`
	RUT                 string    `json:"rut" binding:"required"`
	DeliveredAt         time.Time `json:"delivered_at" binding:"required"`
	SignaturePNGDataURL string    `json:"signature_png_data_url" binding:"required"`
}

// DeliveryConfirmMoveReportItem groups IMEIs that share the same movement (same tracking_id on each device_tracking).
type DeliveryConfirmMoveReportItem struct {
	TrackingID string   `json:"tracking_id" binding:"required"`
	Imeis      []string `json:"imeis" binding:"required,min=1,dive,required"`
}

// DeliveryConfirmMoveReportRequest confirms in-store delivery and triggers move email + PDF (same as move with with_delivery false).
type DeliveryConfirmMoveReportRequest struct {
	Receiver MoveDeliveryReceiver             `json:"receiver" binding:"required"`
	Items    []DeliveryConfirmMoveReportItem `json:"items" binding:"required,min=1,dive"`
}
