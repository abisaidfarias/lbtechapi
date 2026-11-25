package request

// SmrValidationApprove is POST /multibanda/:id/smr-validation/approve.
type SmrValidationApprove struct {
	FMRadio   *bool  `json:"fm_radio"`
	StampCode string `json:"stamp_code"`
	Comment   string `json:"comment"`
}

// SmrValidationReject is POST /multibanda/:id/smr-validation/reject.
type SmrValidationReject struct {
	Comment string `json:"comment"`
}
