package request

type ShipmentControlCertificate struct {
	RegistroOABI        string `json:"registro_oabi" binding:"required"`
	RegisteredImeiCount int    `json:"registered_imei_count" binding:"required_without=ImeiQuantity,omitempty,min=1"`
	ImeiQuantity        int    `json:"imei_quantity" binding:"required_without=RegisteredImeiCount,omitempty,min=1"`
}

func (r *ShipmentControlCertificate) RegisteredCount() int {
	if r.RegisteredImeiCount > 0 {
		return r.RegisteredImeiCount
	}
	return r.ImeiQuantity
}
