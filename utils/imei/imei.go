package imei

import (
	"strings"
	"unicode"
)

// Normalize strips spaces and dashes and keeps digits only.
func Normalize(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsValidLuhn checks the IMEI check digit (15-digit IMEI).
func IsValidLuhn(imei string) bool {
	imei = Normalize(imei)
	if len(imei) != 15 {
		return false
	}
	for _, r := range imei {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	sum := 0
	for i := 0; i < 14; i++ {
		n := int(imei[i] - '0')
		if i%2 == 1 {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}
	check := (10 - (sum % 10)) % 10
	return check == int(imei[14]-'0')
}

// TAC returns the first 8 digits when the IMEI is normalized.
func TAC(imei string) string {
	imei = Normalize(imei)
	if len(imei) < 8 {
		return ""
	}
	return imei[:8]
}
