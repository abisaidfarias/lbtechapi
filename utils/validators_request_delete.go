package utils

import "github.com/abisaidfarias/lbtechapi/viewmodels/request"

func ValidateRequestDeletePatch(body *request.RequestDeletePatch) error {
	if body == nil || !body.RequestDelete {
		return NewValidationError("request_delete must be true")
	}
	return nil
}
