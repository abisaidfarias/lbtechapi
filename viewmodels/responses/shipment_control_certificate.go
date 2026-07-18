package responses

type ShipmentControlCertificateAccepted struct {
	Status     string `json:"status"`
	ShipmentID string `json:"shipment_id"`
}

type ShipmentControlCertificateStatus struct {
	Status        string `json:"status"`
	URL           string `json:"url,omitempty"`
	ControlNumber string `json:"control_number,omitempty"`
	GeneratedAt   string `json:"generated_at,omitempty"`
	Error         string `json:"error,omitempty"`
}
