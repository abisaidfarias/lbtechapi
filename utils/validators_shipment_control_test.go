package utils_test



import (

	"testing"



	"github.com/abisaidfarias/lbtechapi/utils"

	"github.com/abisaidfarias/lbtechapi/viewmodels/request"

)



func TestValidateShipmentControlCreateRequestSuccess(t *testing.T) {

	req := &request.ShipmentControl{

		MultibandaID: "507f1f77bcf86cd799439011",

		ImeiQuantity: 10,

		Country:      "507f1f77bcf86cd799439012",

	}



	if err := utils.ValidateShipmentControlCreateRequest(req); err != nil {

		t.Fatalf("expected valid request, got %v", err)

	}

}



func TestValidateShipmentControlCreateRequestRejectsInvalidImeiQuantity(t *testing.T) {

	req := &request.ShipmentControl{

		MultibandaID: "507f1f77bcf86cd799439011",

		ImeiQuantity: 0,

		Country:      "507f1f77bcf86cd799439012",

	}



	if err := utils.ValidateShipmentControlCreateRequest(req); !utils.IsValidationError(err) {

		t.Fatalf("expected validation error, got %v", err)

	}

}



func TestValidateShipmentControlCreateRequestAllowsMissingCountry(t *testing.T) {

	req := &request.ShipmentControl{

		MultibandaID: "507f1f77bcf86cd799439011",

		ImeiQuantity: 10,

	}



	if err := utils.ValidateShipmentControlCreateRequest(req); err != nil {

		t.Fatalf("expected valid request without country, got %v", err)

	}

}



func TestValidateMultibandaCreateRequestRequiresSoftwareVersion(t *testing.T) {

	req := validMultibandaRequest()

	req.SoftwareVersion = ""



	err := utils.ValidateMultibandaCreateRequest(req)

	if !utils.IsValidationError(err) {

		t.Fatalf("expected validation error, got %v", err)

	}

}


