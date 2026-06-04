package utils_test

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/utils"
)

func TestValidateSARValueAcceptsValidValues(t *testing.T) {
	validValues := []float64{0.01, 0.5, 1, 1.25, 2, 2.00}
	for _, value := range validValues {
		if !utils.ValidateSARValue(value) {
			t.Fatalf("expected %v to be valid", value)
		}
	}
}

func TestValidateSARValueRejectsOutOfRange(t *testing.T) {
	invalidValues := []float64{0, 0.009, 2.01, 3}
	for _, value := range invalidValues {
		if utils.ValidateSARValue(value) {
			t.Fatalf("expected %v to be invalid", value)
		}
	}
}

func TestValidateSARValueRejectsMoreThanTwoDecimals(t *testing.T) {
	invalidValues := []float64{1.234, 0.999, 1.001}
	for _, value := range invalidValues {
		if utils.ValidateSARValue(value) {
			t.Fatalf("expected %v to be invalid", value)
		}
	}
}
