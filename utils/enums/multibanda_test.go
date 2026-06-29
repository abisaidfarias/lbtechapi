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

func TestValidateMultibandaEvaluationTypesRejectsMultipleValues(t *testing.T) {
	err := enums.ValidateMultibandaEvaluationTypes([]string{
		enums.MultibandaEvalSAEMultibandaCertificate,
		enums.MultibandaEvalSismatePeru,
	})
	if err == nil {
		t.Fatal("expected error for multiple evaluation types")
	}
}

func TestValidateMultibandaEvaluationTypesAllowsSingleValue(t *testing.T) {
	cases := []string{
		enums.MultibandaEvalSAEMultibandaCertificate,
		enums.MultibandaEvalSAEOnlyCMASTest,
		enums.MultibandaEvalSismatePeru,
		enums.MultibandaEvalArcotelEcuador,
	}
	for _, code := range cases {
		if err := enums.ValidateMultibandaEvaluationTypes([]string{code}); err != nil {
			t.Fatalf("expected valid single value %q, got %v", code, err)
		}
	}
}

func TestValidateMultibandaTypeRejectsInvalidValue(t *testing.T) {
	err := enums.ValidateMultibandaType("invalid_type")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}
