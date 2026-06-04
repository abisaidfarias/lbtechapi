package request

type ShipmentControl struct {
	MultibandaID    string `json:"multibanda_id" binding:"required"`
	ImeiQuantity    int    `json:"imei_quantity" binding:"required"`
	ImeiFileUrl     string `json:"imei_file_url" binding:"required"`
	ReworkNumber    string `json:"rework_number"`
	OabiCertificate string `json:"oabi_certificate"`
	Client          string `json:"client"`
	Country         string `json:"country"`
}

