package responses

type ShipmentControlBulkValidateSummary struct {
	Total   int `json:"total"`
	Valid   int `json:"valid"`
	Invalid int `json:"invalid"`
}

type ShipmentControlBulkValidateDevicePreview struct {
	CommercialModel string `json:"commercial_model"`
	HardwareVersion string `json:"hardware_version"`
}

type ShipmentControlBulkValidateRow struct {
	RowNumber               int                                       `json:"row_number"`
	Status                  string                                    `json:"status"`
	Client                  string                                    `json:"client,omitempty"`
	ReferenceID             string                                    `json:"reference_id,omitempty"`
	Validation              string                                    `json:"validation,omitempty"`
	ReworkNumber            string                                    `json:"rework_number,omitempty"`
	TechnicalModel          string                                    `json:"technical_model,omitempty"`
	SoftwareVersion         string                                    `json:"software_version,omitempty"`
	ImeiQuantity            int                                       `json:"imei_quantity,omitempty"`
	MultibandaID            string                                    `json:"multibanda_id,omitempty"`
	SubtelCertificateNumber string                                    `json:"subtel_certificate_number,omitempty"`
	Device                  *ShipmentControlBulkValidateDevicePreview `json:"device,omitempty"`
	Errors                  []string                                  `json:"errors,omitempty"`
}

type ShipmentControlBulkValidateResponse struct {
	Summary ShipmentControlBulkValidateSummary `json:"summary"`
	Rows    []ShipmentControlBulkValidateRow    `json:"rows"`
}

type ShipmentControlBulkConfirmSummary struct {
	Created int `json:"created"`
	Failed  int `json:"failed"`
}

type ShipmentControlBulkConfirmResult struct {
	RowNumber         int      `json:"row_number"`
	Status            string   `json:"status"`
	TechnicalModel    string   `json:"technical_model,omitempty"`
	SoftwareVersion   string   `json:"software_version,omitempty"`
	ShipmentControlID string   `json:"shipment_control_id,omitempty"`
	Errors            []string `json:"errors,omitempty"`
}

type ShipmentControlBulkConfirmResponse struct {
	Summary ShipmentControlBulkConfirmSummary  `json:"summary"`
	Results []ShipmentControlBulkConfirmResult `json:"results"`
}
