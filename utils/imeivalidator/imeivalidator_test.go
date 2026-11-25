package imeivalidator

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// imei builds a fake IMEI that passes Luhn from a short number.
func imei(n int) string { return withCheckDigit(fmt.Sprintf("35%012d", n)) }

func badLuhn(valid string) string {
	return fmt.Sprintf("%s%d", valid[:14], (int(valid[14]-'0')+1)%10)
}

func TestValidateIMEI(t *testing.T) {
	good := imei(1)
	cases := []struct{ raw, want, reason string }{
		{good, good, ""},
		{good[:2] + "-" + good[2:8] + " " + good[8:], good, ""},
		{badLuhn(good), "", ReasonLuhn},
		{good[:14], "", ReasonFormat},
		{good + "1", "", ReasonFormat},
		{good[:14] + "A", "", ReasonFormat},
		{"", "", ReasonFormat},
	}
	for _, tc := range cases {
		got, reason := ValidateIMEI(tc.raw)
		if got != tc.want || reason != tc.reason {
			t.Errorf("ValidateIMEI(%q) = %q,%q want %q,%q", tc.raw, got, reason, tc.want, tc.reason)
		}
	}
}

func TestParseQuantity(t *testing.T) {
	for raw, want := range map[string]int{"1": 1, " 2 ": 2, "4": 4, "0": 0, "5": 0, "": 0, "dos": 0, "-1": 0} {
		if got := ParseQuantity(raw); got != want {
			t.Errorf("ParseQuantity(%q) = %d want %d", raw, got, want)
		}
	}
}

// csvRows renders IMEI;S/N;quantity lines (CRLF) like OABI's real file.
func csvRows(sep string, rows ...[3]string) string {
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(r[0] + sep + r[1] + sep + r[2] + "\r\n")
	}
	return b.String()
}

func csvRow(n int, serial, qty string) [3]string { return [3]string{imei(n), serial, qty} }

// Real layout: Dual SIM devices whose rows are not contiguous (all IMEI1s
// first, then the IMEI2s), plus a Single SIM device. S/N contains "/".
func TestClassifyRealLayoutDualAndSingle(t *testing.T) {
	data := csvRows(";",
		csvRow(1, "83904/Y6UQ01392", "2"),
		csvRow(2, "83904/Y6UQ01449", "2"),
		csvRow(3, "83904/Y6UQ01392", "2"),
		csvRow(4, "83904/Y6UQ01449", "2"),
		csvRow(5, "83904/Y6UQ09999", "1"),
	)
	rows, err := ParseUpload(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	c := Classify(rows, nil)
	if len(c.New) != 5 || len(c.Invalid) != 0 {
		t.Fatalf("new=%d invalid=%+v", len(c.New), c.Invalid)
	}
	if c.New[0].Serial != "83904/Y6UQ01392" || c.New[0].ImeiCount != 2 || c.New[4].ImeiCount != 1 {
		t.Errorf("new = %+v", c.New)
	}
}

func TestClassifyRowReasonsAndSnCount(t *testing.T) {
	rows := []UploadRow{
		{Row: 1, IMEI: imei(1), Serial: "A", Quantity: "2"},          // A ok
		{Row: 2, IMEI: imei(2), Serial: "A", Quantity: "2"},          // A ok
		{Row: 3, IMEI: imei(3), Serial: "B", Quantity: "2"},          // B: only 1 row → sn_count
		{Row: 4, IMEI: imei(4), Serial: "", Quantity: "1"},           // serial
		{Row: 5, IMEI: imei(5), Serial: "C", Quantity: "x"},          // quantity
		{Row: 6, IMEI: imei(6), Serial: "D", Quantity: "1"},          // D: different quantities
		{Row: 7, IMEI: imei(7), Serial: "D", Quantity: "2"},          //
		{Row: 8, IMEI: badLuhn(imei(8)), Serial: "E", Quantity: "1"}, // luhn
		{Row: 9, IMEI: "123", Serial: "F", Quantity: "1"},            // format
	}
	c := Classify(rows, nil)
	if len(c.New) != 2 {
		t.Fatalf("new = %+v", c.New)
	}
	got := map[int]string{}
	for _, inv := range c.Invalid {
		got[inv.Row] = inv.Reason
	}
	want := map[int]string{3: ReasonSnCount, 4: ReasonSerial, 5: ReasonQuantity, 6: ReasonSnCount,
		7: ReasonSnCount, 8: ReasonLuhn, 9: ReasonFormat}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reasons = %v, want %v", got, want)
	}
	for _, inv := range c.Invalid {
		if inv.Reason == ReasonSnCount && inv.Detail == "" {
			t.Errorf("sn_count without detail: %+v", inv)
		}
	}
}

// A device is saved whole or not at all.
func TestClassifyMixedDevice(t *testing.T) {
	stored := imei(10)
	rows := []UploadRow{
		// G: one IMEI already stored, one new → new one is device_incomplete.
		{Row: 1, IMEI: stored, Serial: "G", Quantity: "2"},
		{Row: 2, IMEI: imei(11), Serial: "G", Quantity: "2"},
		// H: one IMEI invalid → the valid one is device_incomplete.
		{Row: 3, IMEI: badLuhn(imei(12)), Serial: "H", Quantity: "2"},
		{Row: 4, IMEI: imei(13), Serial: "H", Quantity: "2"},
		// I: every IMEI already stored → reported as stored, nothing invalid.
		{Row: 5, IMEI: imei(14), Serial: "I", Quantity: "2"},
		{Row: 6, IMEI: imei(15), Serial: "I", Quantity: "2"},
	}
	existing := map[string]Existing{
		stored:   {Status: "ready"},
		imei(14): {Status: "registered"},
		imei(15): {Status: "ready"},
	}
	c := Classify(rows, existing)
	if len(c.New) != 0 {
		t.Fatalf("nothing should be saved: %+v", c.New)
	}
	reasons := map[int]string{}
	for _, inv := range c.Invalid {
		reasons[inv.Row] = inv.Reason
	}
	want := map[int]string{2: ReasonDeviceIncomplete, 3: ReasonLuhn, 4: ReasonDeviceIncomplete}
	if !reflect.DeepEqual(reasons, want) {
		t.Fatalf("reasons = %v, want %v", reasons, want)
	}
	if !reflect.DeepEqual(c.AlreadyReady, []string{stored, imei(15)}) || !reflect.DeepEqual(c.AlreadyRegistered, []string{imei(14)}) {
		t.Fatalf("ready=%v registered=%v", c.AlreadyReady, c.AlreadyRegistered)
	}
}

// A repeated IMEI is reported once as a duplicate and does not count as a row
// of its device.
func TestClassifyDuplicateInFile(t *testing.T) {
	rows := []UploadRow{
		{Row: 1, IMEI: imei(1), Serial: "A", Quantity: "1"},
		{Row: 2, IMEI: imei(1), Serial: "A", Quantity: "1"},
		{Row: 3, IMEI: imei(1), Serial: "A", Quantity: "1"},
	}
	c := Classify(rows, nil)
	if len(c.New) != 1 || len(c.Invalid) != 0 {
		t.Fatalf("new=%+v invalid=%+v", c.New, c.Invalid)
	}
	if !reflect.DeepEqual(c.Duplicates, []Duplicate{{IMEI: imei(1), Rows: []int{1, 2, 3}}}) {
		t.Fatalf("duplicates = %+v", c.Duplicates)
	}
}

func TestParseUploadCSVVariants(t *testing.T) {
	base := csvRows(";", csvRow(1, "S/1", "1"))
	comma := csvRows(",", csvRow(1, "S/1", "1"))
	withHeader := "IMEI;S/N;Cantidad\r\n" + base
	withBOM := "\xef\xbb\xbf" + base
	mixedEOL := imei(1) + ";S/1;2\n" + imei(2) + ";S/1;2\r\n"
	for name, data := range map[string]string{
		"semicolon": base, "comma": comma, "header": withHeader, "bom": withBOM,
	} {
		rows, err := ParseUpload(strings.NewReader(data))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(rows) != 1 || rows[0].IMEI != imei(1) || rows[0].Serial != "S/1" || rows[0].Quantity != "1" {
			t.Errorf("%s: rows = %+v", name, rows)
		}
	}
	rows, err := ParseUpload(strings.NewReader(mixedEOL))
	if err != nil || len(rows) != 2 || rows[1].Row != 2 {
		t.Fatalf("mixed line endings: %+v %v", rows, err)
	}
	if rows, _ := ParseUpload(strings.NewReader(withHeader)); rows[0].Row != 2 {
		t.Errorf("row numbers must count the skipped header: %+v", rows)
	}
}

func TestParseUploadXlsx(t *testing.T) {
	f := excelize.NewFile()
	_ = f.SetSheetRow("Sheet1", "A1", &[]interface{}{356938035643809, "83904/Y6UQ01392", 2})
	_ = f.SetSheetRow("Sheet1", "A2", &[]interface{}{imei(2), "83904/Y6UQ01392", "2"})
	var b bytes.Buffer
	if err := f.Write(&b); err != nil {
		t.Fatal(err)
	}
	rows, err := ParseUpload(bytes.NewReader(b.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].IMEI != "356938035643809" || rows[0].Quantity != "2" || rows[1].Serial != "83904/Y6UQ01392" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestParseUploadErrors(t *testing.T) {
	if _, err := ParseUpload(strings.NewReader("")); !errors.Is(err, ErrEmptyFile) {
		t.Errorf("empty: %v", err)
	}
	if _, err := ParseUpload(strings.NewReader("IMEI;S/N;Cantidad\r\n")); !errors.Is(err, ErrEmptyFile) {
		t.Errorf("header only: %v", err)
	}
	if _, err := ParseUpload(bytes.NewReader([]byte{0xD0, 0xCF, 0x11, 0xE0, 0, 0})); !errors.Is(err, ErrLegacyXls) {
		t.Errorf("xls: %v", err)
	}
}

func TestCellNumberAsText(t *testing.T) {
	for in, want := range map[string]string{
		"3.56938035643809E+14": "356938035643809", "2.0": "2", "356938035643809": "356938035643809",
		"35.693803.564380.9": "35.693803.564380.9",
	} {
		if got := cellNumberAsText(in); got != want {
			t.Errorf("cellNumberAsText(%q) = %q want %q", in, got, want)
		}
	}
}

// The OABI file groups each device's IMEIs together (in first-seen order).
func TestBuildOabiFile(t *testing.T) {
	file, err := BuildOabiFile([]OabiRecord{
		{IMEI: imei(1), Serial: "A", ImeiCount: 2},
		{IMEI: imei(2), Serial: "B", ImeiCount: 1},
		{IMEI: imei(3), Serial: "A", ImeiCount: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := imei(1) + ";A;2\r\n" + imei(3) + ";A;2\r\n" + imei(2) + ";B;1\r\n"
	if string(file) != want {
		t.Fatalf("file = %q\nwant  %q", file, want)
	}
	if got := OabiFileName("LOT-20261124-001"); got != "LOT-20261124-001-OABI.csv" {
		t.Errorf("file name = %q", got)
	}
}

// The template is a valid upload: one dual SIM and one single SIM device.
func TestTemplateIsAValidUpload(t *testing.T) {
	tpl, err := BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := ParseUpload(bytes.NewReader(tpl))
	if err != nil {
		t.Fatal(err)
	}
	c := Classify(rows, nil)
	if len(rows) != 3 || len(c.New) != 3 || len(c.Invalid) != 0 {
		t.Fatalf("rows=%d new=%d invalid=%+v", len(rows), len(c.New), c.Invalid)
	}
	if !strings.HasSuffix(string(tpl), "\r\n") || strings.Contains(strings.ToLower(string(tpl)), "imei;") {
		t.Errorf("template must be CRLF with no header: %q", tpl)
	}
}

// 100k rows (50k dual SIM devices) classify quickly.
func TestClassify100kRows(t *testing.T) {
	rows := make([]UploadRow, 0, 100000)
	for i := 0; i < 50000; i++ {
		serial := fmt.Sprintf("S/%d", i)
		rows = append(rows,
			UploadRow{Row: len(rows) + 1, IMEI: imei(2 * i), Serial: serial, Quantity: "2"},
			UploadRow{Row: len(rows) + 2, IMEI: imei(2*i + 1), Serial: serial, Quantity: "2"})
	}
	c := Classify(rows, nil)
	if len(c.New) != 100000 || len(c.Invalid) != 0 || len(c.Duplicates) != 0 {
		t.Fatalf("new=%d invalid=%d dup=%d", len(c.New), len(c.Invalid), len(c.Duplicates))
	}
}
