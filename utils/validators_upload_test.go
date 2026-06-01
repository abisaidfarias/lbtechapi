package utils_test

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/utils"
)

func TestValidateUploadFileSizeAcceptsWithinLimit(t *testing.T) {
	if err := utils.ValidateUploadFileSize(10 << 20); err != nil {
		t.Fatalf("expected valid size, got %v", err)
	}
}

func TestValidateUploadFileSizeRejectsOverLimit(t *testing.T) {
	if err := utils.ValidateUploadFileSize(utils.MaxUploadFileSize + 1); !utils.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestValidateUploadFileSizeRejectsEmpty(t *testing.T) {
	if err := utils.ValidateUploadFileSize(0); !utils.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
