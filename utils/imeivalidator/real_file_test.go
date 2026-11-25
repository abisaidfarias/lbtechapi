package imeivalidator

import (
	"bytes"
	"os"
	"testing"
)

// TestRealOabiFile checks a real OABI file end to end (parse, classify,
// rebuild) without storing real IMEIs in the repo. It runs only when
// IMEI_VALIDATOR_REAL_FILE points at the file.
func TestRealOabiFile(t *testing.T) {
	path := os.Getenv("IMEI_VALIDATOR_REAL_FILE")
	if path == "" {
		t.Skip("set IMEI_VALIDATOR_REAL_FILE to a real OABI csv")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := ParseUpload(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	c := Classify(rows, nil)
	t.Logf("rows=%d new=%d invalid=%d duplicates=%d", len(rows), len(c.New), len(c.Invalid), len(c.Duplicates))
	if len(c.Invalid) != 0 || len(c.Duplicates) != 0 || len(c.New) != len(rows) {
		t.Fatalf("a file OABI accepted must be fully valid: invalid=%+v", c.Invalid)
	}
	records := make([]OabiRecord, len(c.New))
	for i, r := range c.New {
		records[i] = OabiRecord{IMEI: r.IMEI, Serial: r.Serial, ImeiCount: r.ImeiCount}
	}
	rebuilt, err := BuildOabiFile(records)
	if err != nil {
		t.Fatal(err)
	}
	again, err := ParseUpload(bytes.NewReader(rebuilt))
	if err != nil {
		t.Fatal(err)
	}
	if c2 := Classify(again, nil); len(c2.New) != len(rows) || len(c2.Invalid) != 0 {
		t.Fatalf("the rebuilt file must classify the same: %+v", c2.Invalid)
	}
}
