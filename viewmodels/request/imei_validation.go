package request

type ImeiLookup struct {
	IMEI string `json:"imei" binding:"required"`
}
