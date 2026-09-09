package functions_test

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/utils/functions"
)

func TestBuildShipmentCertificateFileName(t *testing.T) {
	tests := []struct {
		name          string
		controlNumber string
		referenceID   string
		reworkNumber  string
		registroOABI  string
		want          string
	}{
		{
			name:          "all fields present",
			controlNumber: "001-260729007",
			referenceID:   "75962",
			reworkNumber:  "DZ202607210023",
			registroOABI:  "1989206",
			want:          "001-260729007-75962-DZ202607210023-1989206.pdf",
		},
		{
			name:          "missing reference id",
			controlNumber: "002-260717001",
			referenceID:   "",
			reworkNumber:  "DZ2026CLB0627",
			registroOABI:  "43443",
			want:          "002-260717001-DZ2026CLB0627-43443.pdf",
		},
		{
			name:          "missing rework number",
			controlNumber: "123-260724001",
			referenceID:   "75966",
			reworkNumber:  "",
			registroOABI:  "1234566",
			want:          "123-260724001-75966-1234566.pdf",
		},
		{
			name:          "missing reference id and rework number",
			controlNumber: "002-260717001",
			referenceID:   "",
			reworkNumber:  "",
			registroOABI:  "43443",
			want:          "002-260717001-43443.pdf",
		},
		{
			name:          "trims whitespace around each segment",
			controlNumber: "  001-260729007  ",
			referenceID:   " 75962 ",
			reworkNumber:  " DZ202607210023 ",
			registroOABI:  " 1989206 ",
			want:          "001-260729007-75962-DZ202607210023-1989206.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := functions.BuildShipmentCertificateFileName(
				tt.controlNumber, tt.referenceID, tt.reworkNumber, tt.registroOABI,
			)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
