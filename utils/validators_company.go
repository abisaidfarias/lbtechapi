package utils

import (
	"fmt"
	"strings"

	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

// Fields a company needs for OABI certificates, in the order reported by
// oabi_missing.
var CompanyOabiFields = []string{"client_id", "rut", "razon_social", "address"}

// NormalizeChileanRUT validates a Chilean RUT (12.345.678-9 or 12345678-9,
// check digit modulo 11, K allowed in either case) and returns it as
// "12.345.678-9". ok is false for any other format or a wrong check digit.
func NormalizeChileanRUT(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	dash := strings.LastIndex(raw, "-")
	if dash <= 0 || dash != len(raw)-2 {
		return "", false
	}
	bodyRaw, dv := raw[:dash], strings.ToUpper(raw[dash+1:])

	// The body is digits, optionally grouped with dots in threes.
	body := strings.ReplaceAll(bodyRaw, ".", "")
	if body == "" || len(body) > 8 || strings.HasPrefix(body, "0") {
		return "", false
	}
	for _, r := range body {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	if strings.Contains(bodyRaw, ".") && bodyRaw != groupThousands(body) {
		return "", false
	}

	if expectedRUTCheckDigit(body) != dv {
		return "", false
	}
	return groupThousands(body) + "-" + dv, true
}

func expectedRUTCheckDigit(body string) string {
	sum, factor := 0, 2
	for i := len(body) - 1; i >= 0; i-- {
		sum += int(body[i]-'0') * factor
		factor++
		if factor > 7 {
			factor = 2
		}
	}
	switch r := 11 - sum%11; r {
	case 11:
		return "0"
	case 10:
		return "K"
	default:
		return fmt.Sprintf("%d", r)
	}
}

func groupThousands(digits string) string {
	var parts []string
	for len(digits) > 3 {
		parts = append([]string{digits[len(digits)-3:]}, parts...)
		digits = digits[:len(digits)-3]
	}
	parts = append([]string{digits}, parts...)
	return strings.Join(parts, ".")
}

// ValidateCompanyRequest checks the required fields of a company create or
// update and normalizes the RUT in place. Uniqueness and the client_id lock
// need the database and are checked by the service.
func ValidateCompanyRequest(req *request.Company) error {
	if req == nil {
		return NewValidationError("company payload is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return NewValidationError("Name is required")
	}
	if strings.TrimSpace(req.ClientID) == "" {
		return NewValidationError("Client ID is required")
	}
	if strings.TrimSpace(req.RazonSocial) == "" {
		return NewValidationError("Razón social is required")
	}
	if strings.TrimSpace(req.Address) == "" {
		return NewValidationError("Address is required")
	}
	if strings.TrimSpace(req.Rut) == "" {
		return NewValidationError("RUT is required")
	}
	rut, ok := NormalizeChileanRUT(req.Rut)
	if !ok {
		return NewValidationError("RUT is not valid: use the format 12.345.678-9 with a correct check digit")
	}
	req.Rut = rut
	return nil
}

// CompanyOabiMissing lists the OABI fields a stored company lacks.
func CompanyOabiMissing(clientID, rut, razonSocial, address string) []string {
	values := map[string]string{
		"client_id":    clientID,
		"rut":          rut,
		"razon_social": razonSocial,
		"address":      address,
	}
	missing := []string{}
	for _, f := range CompanyOabiFields {
		if strings.TrimSpace(values[f]) == "" {
			missing = append(missing, f)
		}
	}
	return missing
}

// CompanyClientIDInUseMessage is the duplicate client_id error.
func CompanyClientIDInUseMessage(clientID, companyName string) string {
	return fmt.Sprintf("Client ID %s is already used by %q", clientID, companyName)
}

const CompanyClientIDLockedMessage = "Client ID is locked: OABI certificates were already issued with it"
