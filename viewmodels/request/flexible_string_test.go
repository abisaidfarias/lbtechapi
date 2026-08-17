package request_test

import (
	"encoding/json"
	"testing"

	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func TestShipmentControlUnmarshalsNumericValidationAndReferenceID(t *testing.T) {
	// Reproduces the exact frontend payload that failed with:
	// "json: cannot unmarshal number into Go struct field ShipmentControl.validation of type string"
	raw := []byte(`{
		"multibanda_id": "6a1df13adbeb0ab5c9f6fc5a",
		"imei_quantity": 55,
		"imei_file_url": "https://example.com/file.pdf",
		"rework_number": "32233",
		"reference_id": "123",
		"validation": 11
	}`)

	var sc request.ShipmentControl
	if err := json.Unmarshal(raw, &sc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.ReferenceID.String() != "123" {
		t.Fatalf("unexpected reference_id: %q", sc.ReferenceID.String())
	}
	if sc.Validation.String() != "11" {
		t.Fatalf("unexpected validation: %q", sc.Validation.String())
	}
}

func TestFlexibleStringUnmarshalVariants(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{"string", `"75962"`, "75962"},
		{"int", `14`, "14"},
		{"float", `4.5`, "4.5"},
		{"empty string", `""`, ""},
		{"null", `null`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f request.FlexibleString
			if err := json.Unmarshal([]byte(tt.json), &f); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f.String() != tt.want {
				t.Fatalf("got %q, want %q", f.String(), tt.want)
			}
		})
	}
}
