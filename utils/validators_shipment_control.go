package utils



import (

	"strings"



	"github.com/abisaidfarias/lbtechapi/utils/enums"

	"github.com/abisaidfarias/lbtechapi/viewmodels/request"

)



func ValidateShipmentControlCreateRequest(req *request.ShipmentControl) error {

	if strings.TrimSpace(req.MultibandaID) == "" {

		return NewValidationError("multibanda_id is required")

	}

	if _, err := ValidateObjectIDField("multibanda_id", req.MultibandaID); err != nil {

		return err

	}

	if req.ImeiQuantity <= 0 {

		return NewValidationError("imei_quantity must be greater than 0")

	}

	if strings.TrimSpace(req.ImeiFileUrl) == "" {

		return NewValidationError("imei_file_url is required")

	}

	if strings.TrimSpace(req.Country) != "" {

		if _, err := ValidateObjectIDField("country", req.Country); err != nil {

			return err

		}

	}

	return nil

}



func ValidateShipmentControlUpdateRequest(req *request.ShipmentControl) error {

	if strings.TrimSpace(req.MultibandaID) == "" {

		return NewValidationError("multibanda_id is required")

	}

	if _, err := ValidateObjectIDField("multibanda_id", req.MultibandaID); err != nil {

		return err

	}

	if !enums.ValidateShipmentControlPhase(req.CurrentPhase) {

		return NewValidationError("invalid current_phase")

	}

	if req.ImeiQuantity <= 0 {

		return NewValidationError("imei_quantity must be greater than 0")

	}

	if strings.TrimSpace(req.ImeiFileUrl) == "" {

		return NewValidationError("imei_file_url is required")

	}

	if strings.TrimSpace(req.Country) == "" {

		return NewValidationError("country is required")

	}

	if _, err := ValidateObjectIDField("country", req.Country); err != nil {

		return err

	}

	return nil

}



func ValidateShipmentControlPhaseRequest(req *request.ShipmentControlResume) error {

	if !enums.ValidateShipmentControlPhase(req.CurrentPhase) {

		return NewValidationError("invalid current_phase")

	}

	if req.ImeiQuantity <= 0 {

		return NewValidationError("imei_quantity must be greater than 0")

	}

	if req.RegisteredImeiCount < 0 {

		return NewValidationError("registered_imei_count must be greater than or equal to 0")

	}

	if strings.TrimSpace(req.Country) != "" {

		if _, err := ValidateObjectIDField("country", req.Country); err != nil {

			return err

		}

	}

	return nil

}


