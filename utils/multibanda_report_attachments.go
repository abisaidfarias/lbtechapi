package utils

import (
	"path/filepath"
	"strings"

	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

const (
	MaxMultibandaAdbInfoBytes          = 2 * 1024 * 1024
	MaxMultibandaBandValidationBytes   = 5 * 1024 * 1024
	MultibandaReportAttachmentAdbInfo  = "adb_info"
	MultibandaReportAttachmentBandValidation = "band_validation"
)

var multibandaAdbInfoExtensions = map[string]struct{}{
	".xlsx": {}, ".xls": {}, ".xlsm": {}, ".xlsb": {}, ".csv": {},
}

var multibandaBandValidationExtensions = map[string]struct{}{
	".zip": {}, ".rar": {}, ".7z": {}, ".tar": {}, ".gz": {},
}

// ValidateMultibandaReportAttachmentMeta checks extension and size for TEST
// ENTRY attachments when a file name is present (upload or saved metadata).
func ValidateMultibandaReportAttachmentMeta(kind, originalFileName string, size int64) error {
	kind = strings.TrimSpace(strings.ToLower(kind))
	switch kind {
	case MultibandaReportAttachmentAdbInfo:
		return validateAttachmentFile(originalFileName, size, multibandaAdbInfoExtensions, MaxMultibandaAdbInfoBytes, "2 MB")
	case MultibandaReportAttachmentBandValidation:
		return validateAttachmentFile(originalFileName, size, multibandaBandValidationExtensions, MaxMultibandaBandValidationBytes, "5 MB")
	default:
		return NewValidationError("invalid attachment kind")
	}
}

func validateAttachmentFile(
	originalFileName string,
	size int64,
	allowed map[string]struct{},
	maxBytes int64,
	maxLabel string,
) error {
	if err := ValidateUploadOriginalFileName(originalFileName); err != nil {
		return err
	}
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(originalFileName)))
	if _, ok := allowed[ext]; !ok {
		return NewValidationError("Formato no permitido")
	}
	if size > maxBytes {
		return NewValidationError("El archivo supera el máximo de " + maxLabel)
	}
	return nil
}

// ValidateMultibandaReportAttachmentsDraft validates attachment metadata on a
// save payload when URLs are present (optional for PDF generation).
func ValidateMultibandaReportAttachmentsDraft(req *request.MultibandaReportSave) error {
	if req == nil {
		return nil
	}
	if url := strings.TrimSpace(req.AdbInfoURL); url != "" {
		name := strings.TrimSpace(req.AdbInfoOriginalFileName)
		if name == "" {
			return NewValidationError("adb_info_original_file_name is required when adb_info_url is set")
		}
		if err := ValidateMultibandaReportAttachmentMeta(MultibandaReportAttachmentAdbInfo, name, 0); err != nil {
			return err
		}
	}
	if url := strings.TrimSpace(req.BandValidationURL); url != "" {
		name := strings.TrimSpace(req.BandValidationOriginalFileName)
		if name == "" {
			return NewValidationError("band_validation_original_file_name is required when band_validation_url is set")
		}
		if err := ValidateMultibandaReportAttachmentMeta(MultibandaReportAttachmentBandValidation, name, 0); err != nil {
			return err
		}
	}
	return nil
}

// ValidateMultibandaReportAttachmentsForGeneration validates attachment metadata
// when present. ADB INFO and Band Validation are not required to generate the PDF.
func ValidateMultibandaReportAttachmentsForGeneration(req *request.MultibandaReportSave, _ string) error {
	return ValidateMultibandaReportAttachmentsDraft(req)
}
