package functions

import "strings"

func EvaluateBulkMultibandaMatch(deviceCount, multibandaCount int, subtelCertificateNumber string) ([]string, bool) {
	if deviceCount == 0 {
		return []string{BulkErrorDeviceNotFound}, false
	}
	if deviceCount > 1 {
		return []string{BulkErrorAmbiguousDevice}, false
	}
	if multibandaCount == 0 {
		return []string{BulkErrorMultibandaNotFound}, false
	}
	if multibandaCount > 1 {
		return []string{BulkErrorAmbiguousMultibanda}, false
	}
	if strings.TrimSpace(subtelCertificateNumber) == "" {
		return []string{BulkErrorCertificateMissing}, false
	}
	return nil, true
}
