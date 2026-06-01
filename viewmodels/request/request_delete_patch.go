package request

type RequestDeletePatch struct {
	RequestDelete bool `json:"request_delete" binding:"required"`
}
