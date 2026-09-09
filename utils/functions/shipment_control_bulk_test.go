package functions_test

import (
	"strings"
	"testing"

	"github.com/abisaidfarias/lbtechapi/utils/functions"
)

func TestParseShipmentControlBulkCSV_Windows1252(t *testing.T) {
	// "Modelo Técnico" encoded as Windows-1252 (byte 0xE9 for é) — common in plain CSV on Windows.
	raw := []byte("Cliente,Qty,Modelo T\xe9cnico,Version SW,# Rework\n")
	raw = append(raw, []byte("Brightcell,350,2602DPT53G,3.0.302.0.WPTMIXM,DZ2026CLB0627\n")...)

	rows, err := functions.ParseShipmentControlBulkCSVFromBytes(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].TechnicalModel != "2602DPT53G" {
		t.Fatalf("unexpected technical_model: %s", rows[0].TechnicalModel)
	}
}

func TestParseShipmentControlBulkCSV_SemicolonDelimiter(t *testing.T) {
	csv := `Cliente;ID;Nombre Comercial;Qty;Modelo Técnico;Valor SAR;Version SW;# Rework;Fec_Sol;Fec_Rec;Validación;OABI;Certificado;Cargador;Fábrica;FF;QR
Brightcell;75966;Xiaomi 17T;350;2602DPT53G;1.08;3.0.302.0.WPTMIXM;DZ2026CLB0627;6/26/2026;;4;;;;;;
`

	rows, err := functions.ParseShipmentControlBulkCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].TechnicalModel != "2602DPT53G" {
		t.Fatalf("unexpected technical_model: %s", rows[0].TechnicalModel)
	}
	if rows[0].ReferenceID != "75966" {
		t.Fatalf("unexpected reference_id: %s", rows[0].ReferenceID)
	}
	if rows[0].Validation != "4" {
		t.Fatalf("unexpected validation: %s", rows[0].Validation)
	}
}

func TestParseShipmentControlBulkCSV(t *testing.T) {
	csv := `Cliente,Qty,Modelo Técnico,Version SW,# Rework
Brightcell,350,2602DPT53G,3.0.302.0.WPTMIXM,DZ2026CLB0627
Brightcell,150,2602EPTC0G,3.0.303.0.WPSMIXM,DZ2026CLB0627
,,,,
`

	rows, err := functions.ParseShipmentControlBulkCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	first := rows[0]
	if first.RowNumber != 2 {
		t.Fatalf("expected row_number 2, got %d", first.RowNumber)
	}
	if first.Client != "Brightcell" {
		t.Fatalf("unexpected client: %s", first.Client)
	}
	if first.Qty != 350 {
		t.Fatalf("expected qty 350, got %d", first.Qty)
	}
	if first.ImeiQuantity != 700 {
		t.Fatalf("expected imei_quantity 700, got %d", first.ImeiQuantity)
	}
	if first.TechnicalModel != "2602DPT53G" {
		t.Fatalf("unexpected technical_model: %s", first.TechnicalModel)
	}
	if first.SoftwareVersion != "3.0.302.0.WPTMIXM" {
		t.Fatalf("unexpected software_version: %s", first.SoftwareVersion)
	}
	if first.ReworkNumber != "DZ2026CLB0627" {
		t.Fatalf("unexpected rework_number: %s", first.ReworkNumber)
	}
}

func TestParseShipmentControlBulkCSV_MissingColumns(t *testing.T) {
	csv := `Cliente,Qty,Modelo Técnico
Brightcell,350,2602DPT53G
`
	_, err := functions.ParseShipmentControlBulkCSV(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected missing columns error")
	}
}

func TestValidateShipmentControlBulkCSVFields(t *testing.T) {
	row := functions.ShipmentControlBulkCSVRow{
		Client:          "Brightcell",
		QtyRaw:          "150",
		Qty:             150,
		TechnicalModel:  "2602EPTC0G",
		SoftwareVersion: "3.0.303.0.WPSMIXM",
		ImeiQuantity:    300,
	}
	if errors := functions.ValidateShipmentControlBulkCSVFields(row); len(errors) != 0 {
		t.Fatalf("expected no errors, got %v", errors)
	}
}

func TestEvaluateBulkMultibandaMatch(t *testing.T) {
	tests := []struct {
		name       string
		devices    int
		multibandas int
		cert       string
		wantError  string
		wantOK     bool
	}{
		{name: "device not found", devices: 0, multibandas: 0, wantError: functions.BulkErrorDeviceNotFound},
		{name: "ambiguous device", devices: 2, multibandas: 0, wantError: functions.BulkErrorAmbiguousDevice},
		{name: "multibanda not found", devices: 1, multibandas: 0, wantError: functions.BulkErrorMultibandaNotFound},
		{name: "ambiguous multibanda", devices: 1, multibandas: 2, wantError: functions.BulkErrorAmbiguousMultibanda},
		{name: "certificate missing", devices: 1, multibandas: 1, cert: "  ", wantError: functions.BulkErrorCertificateMissing},
		{name: "valid match", devices: 1, multibandas: 1, cert: "CERT-123", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, ok := functions.EvaluateBulkMultibandaMatch(tt.devices, tt.multibandas, tt.cert)
			if tt.wantOK {
				if !ok || len(errors) != 0 {
					t.Fatalf("expected valid match, got errors=%v ok=%v", errors, ok)
				}
				return
			}
			if ok || len(errors) != 1 || errors[0] != tt.wantError {
				t.Fatalf("expected error %s, got errors=%v ok=%v", tt.wantError, errors, ok)
			}
		})
	}
}
