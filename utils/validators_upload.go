package utils

import "fmt"

func ValidateUploadFileSize(size int64) error {
	if size <= 0 {
		return NewValidationError("file is empty")
	}
	if size > MaxUploadFileSize {
		return NewValidationError(fmt.Sprintf("file exceeds maximum size of %dMB", MaxUploadFileSize/(1<<20)))
	}
	return nil
}
