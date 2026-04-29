package enums

import "fmt"

// Process type labels (stored as-is for display; Spanish copy matches the product UI).
const (
	ProcessTypeTesting                  = "Testing"
	ProcessTypeCertificadoMultibandaSAE = "Certificado Multibanda SAE"
	ProcessTypeATP                      = "ATP"
	ProcessTypeDevolucionDispositivos   = "Devolucion de dispositivo(s)"
)

var allowedProcessTypes = map[string]struct{}{
	ProcessTypeTesting:                  {},
	ProcessTypeCertificadoMultibandaSAE: {},
	ProcessTypeATP:                      {},
	ProcessTypeDevolucionDispositivos:   {},
}

// AllowedProcessTypes is the ordered list for clients (dropdowns, docs).
var AllowedProcessTypes = []string{
	ProcessTypeTesting,
	ProcessTypeCertificadoMultibandaSAE,
	ProcessTypeATP,
	ProcessTypeDevolucionDispositivos,
}

// ValidateProcessTypes returns an error if any value is not one of the allowed labels.
func ValidateProcessTypes(types []string) error {
	for _, t := range types {
		if _, ok := allowedProcessTypes[t]; !ok {
			return fmt.Errorf("invalid process_types value %q: must be one of %v", t, AllowedProcessTypes)
		}
	}
	return nil
}
