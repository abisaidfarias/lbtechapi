package imeivalidator

import (
	"strconv"
	"strings"

	imeiutil "github.com/abisaidfarias/lbtechapi/utils/imei"
)

// Invalid reasons.
const (
	ReasonFormat = "format" // not 15 digits / unexpected characters
	ReasonLuhn   = "luhn"   // wrong check digit
	ReasonSerial = "serial" // empty S/N
	// ReasonQuantity: quantity empty or not an integer 1..MaxImeisPerSerial.
	ReasonQuantity = "quantity"
	// ReasonSnCount: the rows of the S/N do not match its declared quantity,
	// or its rows declare different quantities. Set on every row of the S/N.
	ReasonSnCount = "sn_count"
	// ReasonDeviceIncomplete: a valid, not yet stored IMEI whose device (S/N)
	// cannot be saved whole, because another of its IMEIs is invalid or
	// already stored. OABI needs every IMEI of a S/N in the same file.
	ReasonDeviceIncomplete = "device_incomplete"
)

// MaxImeisPerSerial bounds the declared quantity (OABI: up to 4 IMEIs per
// device).
const MaxImeisPerSerial = 4

// Existing is a stored IMEI found for an uploaded one.
type Existing struct {
	Status string // ready | registered
}

// InvalidRow is a row that will not be saved, with the reason.
type InvalidRow struct {
	Row    int
	Value  string
	Serial string
	Reason string
	Detail string
}

// Duplicate is an IMEI repeated in the file: Rows lists every row it
// appears on; only the first one is classified, the others are ignored.
type Duplicate struct {
	IMEI string
	Rows []int
}

// NewRecord is a row to store: IMEI normalized, quantity parsed.
type NewRecord struct {
	Row       int
	IMEI      string
	Serial    string
	ImeiCount int
}

// Classification splits the upload. New keeps the IMEIs of complete devices
// whose IMEIs are all new, in file order. AlreadyReady / AlreadyRegistered
// keep the stored IMEIs found (normalized), in file order.
type Classification struct {
	TotalRows         int
	New               []NewRecord
	AlreadyReady      []string
	AlreadyRegistered []string
	Duplicates        []Duplicate
	Invalid           []InvalidRow
}

// NormalizeIMEI strips the separators people type (spaces, dashes, dots) and
// reports whether only digits remain. Letters or other symbols are a format
// error rather than being silently dropped.
func NormalizeIMEI(raw string) (string, bool) {
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '.' || r == '\t':
		default:
			return "", false
		}
	}
	return b.String(), true
}

// ValidateIMEI returns the normalized IMEI or the invalid reason.
func ValidateIMEI(raw string) (string, string) {
	imei, ok := NormalizeIMEI(raw)
	if !ok || len(imei) != 15 {
		return "", ReasonFormat
	}
	if !imeiutil.IsValidLuhn(imei) {
		return "", ReasonLuhn
	}
	return imei, ""
}

// ParseQuantity returns the declared quantity, or 0 when it is not an
// integer 1..MaxImeisPerSerial.
func ParseQuantity(raw string) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 1 || n > MaxImeisPerSerial {
		return 0
	}
	return n
}

// ValidIMEIs returns the distinct valid IMEIs of the upload, the input for
// the bulk existence lookup.
func ValidIMEIs(rows []UploadRow) []string {
	seen := make(map[string]struct{}, len(rows))
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		imei, reason := ValidateIMEI(r.IMEI)
		if reason != "" {
			continue
		}
		if _, ok := seen[imei]; ok {
			continue
		}
		seen[imei] = struct{}{}
		out = append(out, imei)
	}
	return out
}

// row is the working state of one non-duplicate row.
type row struct {
	UploadRow
	imei     string
	quantity int
	reason   string
	detail   string
}

// Classify sorts every row.
//
//  1. Rows repeating an IMEI already seen in the file go to Duplicates and
//     are otherwise ignored.
//  2. Each row is checked on its own: IMEI (format, luhn), S/N (serial),
//     quantity (quantity).
//  3. Rows are grouped by S/N. A S/N whose row count differs from its
//     declared quantity, or whose rows declare different quantities, marks
//     every one of its rows not already invalid as sn_count.
//  4. A device (S/N) is saved only when all its rows are valid and all its
//     IMEIs are new. Otherwise its stored IMEIs are reported as already
//     ready / registered, and its new valid IMEIs as device_incomplete.
func Classify(rows []UploadRow, existing map[string]Existing) Classification {
	c := Classification{TotalRows: len(rows)}

	// 1-2. duplicates and per-row checks.
	var work []*row
	firstRow := map[string]int{}
	dupIndex := map[string]int{}
	for _, r := range rows {
		w := &row{UploadRow: r}
		imei, reason := ValidateIMEI(r.IMEI)
		if reason == "" {
			if first, seen := firstRow[imei]; seen {
				if i, ok := dupIndex[imei]; ok {
					c.Duplicates[i].Rows = append(c.Duplicates[i].Rows, r.Row)
				} else {
					dupIndex[imei] = len(c.Duplicates)
					c.Duplicates = append(c.Duplicates, Duplicate{IMEI: imei, Rows: []int{first, r.Row}})
				}
				continue
			}
			firstRow[imei] = r.Row
			w.imei = imei
		}
		w.quantity = ParseQuantity(r.Quantity)
		switch {
		case reason != "":
			w.reason = reason
		case strings.TrimSpace(r.Serial) == "":
			w.reason = ReasonSerial
		case w.quantity == 0:
			w.reason = ReasonQuantity
		}
		work = append(work, w)
	}

	// 3. group by S/N and check the declared quantity.
	bySerial := map[string][]*row{}
	var serialOrder []string
	for _, w := range work {
		s := strings.TrimSpace(w.Serial)
		if s == "" {
			continue
		}
		if _, ok := bySerial[s]; !ok {
			serialOrder = append(serialOrder, s)
		}
		bySerial[s] = append(bySerial[s], w)
	}
	for _, s := range serialOrder {
		group := bySerial[s]
		declared, mixed := 0, false
		for _, w := range group {
			if w.quantity == 0 {
				continue
			}
			if declared == 0 {
				declared = w.quantity
			} else if w.quantity != declared {
				mixed = true
			}
		}
		if mixed || (declared != 0 && len(group) != declared) {
			detail := snCountDetail(len(group), declared, mixed)
			for _, w := range group {
				if w.reason == "" {
					w.reason, w.detail = ReasonSnCount, detail
				}
			}
		}
	}

	// 4. device decision.
	for _, s := range serialOrder {
		group := bySerial[s]
		complete := true
		for _, w := range group {
			if w.reason != "" {
				complete = false
				continue
			}
			if _, stored := existing[w.imei]; stored {
				complete = false
			}
		}
		if complete {
			continue // saved as new below
		}
		for _, w := range group {
			if w.reason != "" {
				continue
			}
			if _, stored := existing[w.imei]; !stored {
				w.reason = ReasonDeviceIncomplete
				w.detail = "S/N " + s + ": not saved because another IMEI of this device is invalid or already stored; fix the whole device and upload it again"
			}
		}
	}

	// Emit in file order.
	for _, w := range work {
		switch {
		case w.reason != "":
			c.Invalid = append(c.Invalid, InvalidRow{
				Row: w.Row, Value: w.IMEI, Serial: strings.TrimSpace(w.Serial), Reason: w.reason, Detail: w.detail,
			})
		default:
			e, stored := existing[w.imei]
			switch {
			case !stored:
				c.New = append(c.New, NewRecord{Row: w.Row, IMEI: w.imei, Serial: strings.TrimSpace(w.Serial), ImeiCount: w.quantity})
			case e.Status == "registered":
				c.AlreadyRegistered = append(c.AlreadyRegistered, w.imei)
			default:
				c.AlreadyReady = append(c.AlreadyReady, w.imei)
			}
		}
	}
	return c
}

func snCountDetail(rows, declared int, mixed bool) string {
	if mixed {
		return "the rows of this S/N declare different quantities"
	}
	return "this S/N has " + strconv.Itoa(rows) + " row(s) but declares quantity " + strconv.Itoa(declared)
}
