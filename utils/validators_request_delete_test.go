package utils_test

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func TestValidateRequestDeletePatchRequiresTrue(t *testing.T) {
	if err := utils.ValidateRequestDeletePatch(&request.RequestDeletePatch{RequestDelete: true}); err != nil {
		t.Fatalf("expected valid body, got %v", err)
	}
	if err := utils.ValidateRequestDeletePatch(&request.RequestDeletePatch{RequestDelete: false}); !utils.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
