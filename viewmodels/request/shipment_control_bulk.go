package request

type ShipmentControlBulkConfirmRow struct {
	RowNumber    int    `json:"row_number" binding:"required"`
	MultibandaID string `json:"multibanda_id" binding:"required"`
	Client       string `json:"client"`
	ReworkNumber string `json:"rework_number"`
	ImeiQuantity int    `json:"imei_quantity" binding:"required"`
	ImeiFileUrl  string `json:"imei_file_url" binding:"required"`
}

type ShipmentControlBulkConfirm struct {
	Rows []ShipmentControlBulkConfirmRow `json:"rows" binding:"required,min=1,dive"`
}
