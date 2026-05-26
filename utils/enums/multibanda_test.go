package enums_test

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
)

func TestValidateMultibandaEvaluationTypesRejectsEmpty(t *testing.T) {
	err := enums.ValidateMultibandaEvaluationTypes(nil)
	if err == nil {
		t.Fatal("expected error for empty evaluation_type")
	}
}

func TestValidateMultibandaEvaluationTypesRejectsMutuallyExclusiveValues(t *testing.T) {
	err := enums.ValidateMultibandaEvaluationTypes([]string{
		enums.MultibandaEvalSAEMultibandaCertificate,
		enums.MultibandaEvalSAEOnlyCMASTest,
	})
	if err == nil {
		t.Fatal("expected error for mutually exclusive evaluation types")
	}
}

func TestValidateMultibandaEvaluationTypesAllowsSismateCombination(t *testing.T) {
	err := enums.ValidateMultibandaEvaluationTypes([]string{
		enums.MultibandaEvalSAEMultibandaCertificate,
		enums.MultibandaEvalSismatePeru,
	})
	if err != nil {
		t.Fatalf("expected valid combination, got %v", err)
	}
}

func TestValidateMultibandaTypeRejectsInvalidValue(t *testing.T) {
	err := enums.ValidateMultibandaType("invalid_type")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}
