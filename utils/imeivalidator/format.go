// Package imeivalidator holds the file-format and classification logic of the
// IMEI Validator, kept free of HTTP and Mongo so it can be tested directly.
//
// FORMAT-DEPENDENT CODE LIVES ONLY IN THIS FILE. The format is OABI's real
// one (Nov 2026): a CSV with no header row and three columns
//
//	IMEI ; S/N ; quantity of IMEIs for that S/N
//
// A Dual SIM device is two rows with the same S/N and quantity 2 (not
// necessarily contiguous); a Single SIM device is one row with quantity 1.
// OABI's real files use ';' (Spanish Excel) and CRLF.
package imeivalidator

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Errors returned for unusable files; the service maps them to 400.
var (
	ErrEmptyFile  = errors.New("the file is empty: it has no IMEI rows")
	ErrUnreadable = errors.New("the file could not be read: upload a .csv (IMEI;S/N;quantity) or .xlsx file")
	ErrLegacyXls  = errors.New("old .xls files are not supported: save it as .xlsx or .csv and upload it again")
)

// File names and content type of the generated files.
const (
	TemplateFileName = "IMEI-template.csv"
	CSVContentType   = "text/csv; charset=utf-8"
)

// UploadRow is one data row of the uploaded file. Row is the line/row number
// in the file (1-based, counting a skipped header), so users can find it.
type UploadRow struct {
	Row      int
	IMEI     string // raw value, normalized later by Classify
	Serial   string // S/N, trimmed
	Quantity string // raw declared quantity, validated by Classify
}

// OabiRecord is what the OABI file needs from each stored IMEI.
type OabiRecord struct {
	IMEI      string
	Serial    string
	ImeiCount int
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// ParseUpload reads a .csv (';' or ',', auto-detected) or .xlsx upload with
// the columns IMEI, S/N, quantity and no header. A first row whose IMEI is
// not numeric (e.g. "IMEI") is taken as a header and skipped.
func ParseUpload(r io.Reader) ([]UploadRow, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, ErrUnreadable
	}
	data = bytes.TrimPrefix(data, utf8BOM)

	var records [][]string
	switch {
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		records, err = readXlsx(data)
	case bytes.HasPrefix(data, []byte{0xD0, 0xCF, 0x11, 0xE0}):
		return nil, ErrLegacyXls
	default:
		records, err = readCSV(data)
	}
	if err != nil {
		return nil, err
	}

	var out []UploadRow
	for i, cells := range records {
		row := UploadRow{
			Row:      i + 1,
			IMEI:     cellNumberAsText(cell(cells, 0)),
			Serial:   cell(cells, 1),
			Quantity: cellNumberAsText(cell(cells, 2)),
		}
		if row.IMEI == "" && row.Serial == "" && row.Quantity == "" {
			continue // blank line
		}
		if i == 0 && looksLikeHeader(row.IMEI) {
			continue
		}
		out = append(out, row)
	}
	if len(out) == 0 {
		return nil, ErrEmptyFile
	}
	return out, nil
}

func readCSV(data []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = detectDelimiter(data)
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	records, err := r.ReadAll() // handles CRLF, LF and a mix of both
	if err != nil {
		return nil, ErrUnreadable
	}
	return records, nil
}

// detectDelimiter picks ';' or ',' from the first non-empty line (';' wins
// a tie, as OABI's real files use it).
func detectDelimiter(data []byte) rune {
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.Count(line, ",") > strings.Count(line, ";") {
			return ','
		}
		return ';'
	}
	return ';'
}

func readXlsx(data []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, ErrUnreadable
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrEmptyFile
	}
	rows, err := f.Rows(sheets[0])
	if err != nil {
		return nil, ErrUnreadable
	}
	defer rows.Close()
	var out [][]string
	for rows.Next() {
		cells, err := rows.Columns(excelize.Options{RawCellValue: true})
		if err != nil {
			return nil, ErrUnreadable
		}
		out = append(out, cells)
	}
	return out, nil
}

func cell(cells []string, i int) string {
	if i >= len(cells) {
		return ""
	}
	return strings.TrimSpace(cells[i])
}

// looksLikeHeader: a first-row IMEI without any digit ("IMEI", "imei 1").
func looksLikeHeader(imei string) bool {
	return !strings.ContainsAny(imei, "0123456789")
}

// cellNumberAsText undoes Excel's scientific notation and float rendering
// for numbers ("3.56938035643809E+14" → "356938035643809", "2.0" → "2").
func cellNumberAsText(v string) string {
	if strings.ContainsAny(v, "eE") {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil || f != float64(int64(f)) {
			return v
		}
		return strconv.FormatFloat(f, 'f', 0, 64)
	}
	// "2.0" / "356938035643809.00" → drop the zero decimals. Dots used as
	// separators ("35.693803.564380.9") are left for NormalizeIMEI.
	if i := strings.Index(v, "."); i > 0 && strings.Count(v, ".") == 1 &&
		strings.Trim(v[i+1:], "0") == "" && strings.Trim(v[:i], "0123456789") == "" {
		return v[:i]
	}
	return v
}

// BuildOabiFile returns the file to load in OABI: CSV with ';', CRLF, no
// header, one row per IMEI as IMEI;S/N;quantity. The IMEIs of a S/N are kept
// together, devices in the order they first appear.
func BuildOabiFile(records []OabiRecord) ([]byte, error) {
	var order []string
	bySerial := map[string][]OabiRecord{}
	for _, r := range records {
		if _, ok := bySerial[r.Serial]; !ok {
			order = append(order, r.Serial)
		}
		bySerial[r.Serial] = append(bySerial[r.Serial], r)
	}
	var b bytes.Buffer
	for _, serial := range order {
		for _, r := range bySerial[serial] {
			fmt.Fprintf(&b, "%s;%s;%d\r\n", r.IMEI, r.Serial, r.ImeiCount)
		}
	}
	return b.Bytes(), nil
}

// OabiFileName is "<batch code>-OABI.csv".
func OabiFileName(batchCode string) string {
	return fmt.Sprintf("%s-OABI.csv", batchCode)
}

// BuildTemplate returns an example upload in the OABI format: one Dual SIM
// device (two rows, quantity 2) and one Single SIM device (one row,
// quantity 1). The IMEIs are fake but pass the Luhn check.
func BuildTemplate() ([]byte, error) {
	rows := []OabiRecord{
		{IMEI: withCheckDigit("35000000000001"), Serial: "00000/TEST00001", ImeiCount: 2},
		{IMEI: withCheckDigit("35000000000002"), Serial: "00000/TEST00001", ImeiCount: 2},
		{IMEI: withCheckDigit("35000000000003"), Serial: "00000/TEST00002", ImeiCount: 1},
	}
	return BuildOabiFile(rows)
}

// withCheckDigit appends the Luhn check digit to a 14-digit IMEI body.
func withCheckDigit(body string) string {
	sum := 0
	for i := 0; i < 14; i++ {
		n := int(body[i] - '0')
		if i%2 == 1 {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}
	return fmt.Sprintf("%s%d", body, (10-sum%10)%10)
}
