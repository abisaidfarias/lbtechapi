package functions

import (
	"fmt"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
)

// ExcelDate formats a timestamp for spreadsheet export (dd/mm/yyyy), matching homologation exports.
func ExcelDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	y, m, d := t.Date()
	return fmt.Sprintf("%d/%d/%d", d, int(m), y)
}

// ExcelYesNo renders booleans for spreadsheet export.
func ExcelYesNo(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

// MultibandaEvaluationTypeLabels joins evaluation type codes with human-readable labels.
func MultibandaEvaluationTypeLabels(codes []string) string {
	if len(codes) == 0 {
		return ""
	}
	parts := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if label, ok := enums.MultibandaEvaluationTypeLabels[code]; ok {
			parts = append(parts, label)
		} else {
			parts = append(parts, code)
		}
	}
	return strings.Join(parts, ", ")
}
