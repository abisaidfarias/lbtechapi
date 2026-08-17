package request

import (
	"encoding/json"
	"fmt"
)

// FlexibleString unmarshals from either a JSON string or a JSON number,
// storing the value as plain text. Some frontend inputs (e.g. reference
// codes that happen to look numeric) serialize as a JSON number even though
// the backend treats the field as free text, which a plain `string` field
// rejects with "cannot unmarshal number into Go struct field".
type FlexibleString string

func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*f = ""
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleString(s)
		return nil
	}

	// json.Number preserves the literal digits (no float rounding) for any
	// numeric input, including large integers.
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleString(n.String())
		return nil
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*f = FlexibleString(fmt.Sprintf("%v", raw))
	return nil
}

func (f FlexibleString) String() string {
	return string(f)
}
