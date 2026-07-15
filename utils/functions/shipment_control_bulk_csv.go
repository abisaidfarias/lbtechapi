package functions

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	BulkErrorMissingClient           = "missing_client"
	BulkErrorMissingQty              = "missing_qty"
	BulkErrorInvalidQty              = "invalid_qty"
	BulkErrorMissingTechnicalModel   = "missing_technical_model"
	BulkErrorMissingSoftwareVersion  = "missing_software_version"
	BulkErrorDeviceNotFound          = "device_not_found"
	BulkErrorAmbiguousDevice         = "ambiguous_device"
	BulkErrorMultibandaNotFound      = "multibanda_not_found"
	BulkErrorAmbiguousMultibanda     = "ambiguous_multibanda"
	BulkErrorCertificateMissing      = "certificate_missing"
	BulkErrorMissingImeiFileURL      = "missing_imei_file_url"
	BulkErrorInvalidMultibandaID     = "invalid_multibanda_id"
	BulkErrorMultibandaNotApproved   = "multibanda_not_approved"
	BulkErrorMultibandaDeletePending = "multibanda_delete_pending"
	BulkErrorForbidden               = "forbidden"
)

var stripAccents = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

type ShipmentControlBulkCSVRow struct {
	RowNumber       int
	Client          string
	QtyRaw          string
	Qty             int
	TechnicalModel  string
	SoftwareVersion string
	ReworkNumber    string
	ImeiQuantity    int
}

// Keys are accent-insensitive, lowercased header names.
var shipmentControlBulkCSVHeaders = map[string]string{
	"cliente":        "client",
	"qty":            "qty",
	"modelo tecnico": "technical_model",
	"version sw":     "software_version",
	"# rework":       "rework_number",
	"rework":         "rework_number",
}

var bulkCSVRequiredColumnLabels = map[string]string{
	"client":           "Cliente",
	"qty":              "Qty",
	"technical_model":  "Modelo Tecnico",
	"software_version": "Version SW",
	"rework_number":    "# Rework",
}

func DecodeBulkCSVContent(raw []byte) (string, error) {
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	if len(raw) == 0 {
		return "", nil
	}
	if utf8.Valid(raw) {
		return string(raw), nil
	}
	decoded, err := charmap.Windows1252.NewDecoder().Bytes(raw)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func ParseShipmentControlBulkCSVFromBytes(raw []byte) ([]ShipmentControlBulkCSVRow, error) {
	decoded, err := DecodeBulkCSVContent(raw)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(decoded) == "" {
		return nil, fmt.Errorf("csv header: empty file")
	}
	return ParseShipmentControlBulkCSV(strings.NewReader(decoded))
}

func ParseShipmentControlBulkCSV(reader io.Reader) ([]ShipmentControlBulkCSVRow, error) {
	csvReader, err := newShipmentControlBulkCSVReader(reader)
	if err != nil {
		return nil, err
	}

	headerRecord, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}

	columnIndex, err := mapShipmentControlBulkCSVHeaders(headerRecord)
	if err != nil {
		return nil, err
	}

	rows := []ShipmentControlBulkCSVRow{}
	rowNumber := 1

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv row %d: %w", rowNumber+1, err)
		}

		rowNumber++
		parsed, ok := parseShipmentControlBulkCSVRecord(record, columnIndex, rowNumber)
		if !ok {
			continue
		}
		rows = append(rows, parsed)
	}

	return rows, nil
}

func newShipmentControlBulkCSVReader(reader io.Reader) (*csv.Reader, error) {
	bufReader := bufio.NewReader(reader)
	firstLine, err := bufReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("csv header: %w", err)
	}
	if strings.TrimSpace(firstLine) == "" {
		return nil, fmt.Errorf("csv header: empty file")
	}

	firstLine = strings.TrimSuffix(firstLine, "\r")
	firstLine = strings.TrimSuffix(firstLine, "\n")

	csvReader := csv.NewReader(io.MultiReader(strings.NewReader(firstLine+"\n"), bufReader))
	csvReader.Comma = detectShipmentControlBulkCSVDelimiter(firstLine)
	csvReader.TrimLeadingSpace = true
	csvReader.LazyQuotes = true
	return csvReader, nil
}

func detectShipmentControlBulkCSVDelimiter(headerLine string) rune {
	tabs := strings.Count(headerLine, "\t")
	semicolons := strings.Count(headerLine, ";")
	commas := strings.Count(headerLine, ",")
	if tabs > semicolons && tabs > commas {
		return '\t'
	}
	if semicolons > commas {
		return ';'
	}
	return ','
}

func mapShipmentControlBulkCSVHeaders(headerRecord []string) (map[string]int, error) {
	columnIndex := map[string]int{}
	foundHeaders := make([]string, 0, len(headerRecord))

	for index, header := range headerRecord {
		trimmed := strings.TrimSpace(header)
		if trimmed != "" {
			foundHeaders = append(foundHeaders, trimmed)
		}
		field, ok := matchBulkCSVHeader(trimmed)
		if !ok {
			continue
		}
		columnIndex[field] = index
	}

	required := []string{"client", "qty", "technical_model", "software_version", "rework_number"}
	missing := []string{}
	for _, field := range required {
		if _, ok := columnIndex[field]; !ok {
			missing = append(missing, bulkCSVRequiredColumnLabels[field])
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf(
			"csv missing required columns: %s (headers found: %s)",
			strings.Join(missing, ", "),
			strings.Join(foundHeaders, " | "),
		)
	}

	return columnIndex, nil
}

func matchBulkCSVHeader(header string) (string, bool) {
	normalized := foldBulkCSVHeader(header)
	if field, ok := shipmentControlBulkCSVHeaders[normalized]; ok {
		return field, true
	}

	switch {
	case strings.Contains(normalized, "cliente"):
		return "client", true
	case normalized == "qty" || strings.HasSuffix(normalized, " qty"):
		return "qty", true
	case strings.Contains(normalized, "modelo") && strings.Contains(normalized, "tecnic"):
		return "technical_model", true
	case strings.Contains(normalized, "version") && strings.Contains(normalized, "sw"):
		return "software_version", true
	case strings.Contains(normalized, "rework"):
		return "rework_number", true
	}

	return "", false
}

func foldBulkCSVHeader(header string) string {
	header = strings.TrimSpace(header)
	header = strings.TrimPrefix(header, "\ufeff")
	header = strings.ReplaceAll(header, "\u00a0", " ")
	header = strings.ReplaceAll(header, "\u200b", "")

	folded, _, err := transform.String(stripAccents, header)
	if err != nil {
		folded = header
	}
	return strings.ToLower(folded)
}

func parseShipmentControlBulkCSVRecord(record []string, columnIndex map[string]int, rowNumber int) (ShipmentControlBulkCSVRow, bool) {
	valueAt := func(field string) string {
		index, ok := columnIndex[field]
		if !ok || index >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[index])
	}

	client := valueAt("client")
	qtyRaw := valueAt("qty")
	technicalModel := valueAt("technical_model")
	softwareVersion := valueAt("software_version")
	reworkNumber := valueAt("rework_number")

	if client == "" && qtyRaw == "" && technicalModel == "" && softwareVersion == "" && reworkNumber == "" {
		return ShipmentControlBulkCSVRow{}, false
	}

	qty := 0
	if qtyRaw != "" {
		parsedQty, err := strconv.Atoi(qtyRaw)
		if err == nil {
			qty = parsedQty
		}
	}

	return ShipmentControlBulkCSVRow{
		RowNumber:       rowNumber,
		Client:          client,
		QtyRaw:          qtyRaw,
		Qty:             qty,
		TechnicalModel:  technicalModel,
		SoftwareVersion: softwareVersion,
		ReworkNumber:    reworkNumber,
		ImeiQuantity:    qty * 2,
	}, true
}

func ValidateShipmentControlBulkCSVFields(row ShipmentControlBulkCSVRow) []string {
	errors := []string{}
	if row.Client == "" {
		errors = append(errors, BulkErrorMissingClient)
	}
	if strings.TrimSpace(row.QtyRaw) == "" {
		errors = append(errors, BulkErrorMissingQty)
	} else if row.Qty <= 0 {
		errors = append(errors, BulkErrorInvalidQty)
	}
	if row.TechnicalModel == "" {
		errors = append(errors, BulkErrorMissingTechnicalModel)
	}
	if row.SoftwareVersion == "" {
		errors = append(errors, BulkErrorMissingSoftwareVersion)
	}
	return errors
}
