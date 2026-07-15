package utils

import (
	"fmt"
	"strings"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/gabriel-vasile/mimetype"
)

func ValidateUploadFileSize(size int64) error {
	if size <= 0 {
		return NewValidationError("file is empty")
	}
	if size > MaxUploadFileSize {
		return NewValidationError(fmt.Sprintf("file exceeds maximum size of %dMB", MaxUploadFileSize/(1<<20)))
	}
	return nil
}

// ValidateUploadContentType ensures detected MIME type is allowed for user uploads.
func ValidateUploadContentType(content []byte) error {
	mtype := mimetype.Detect(content)
	if mtype == nil {
		return NewValidationError("unsupported file type")
	}
	ext := strings.ToLower(mtype.Extension())
	for _, allowedExt := range enums.MimeTypes_value {
		if strings.EqualFold(ext, allowedExt) {
			return nil
		}
	}
	return NewValidationError(fmt.Sprintf("unsupported file type: %s", mtype.String()))
}
