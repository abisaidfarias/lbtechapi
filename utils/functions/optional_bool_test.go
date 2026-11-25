package functions

import "testing"

func TestOptionalBoolHelpers(t *testing.T) {
	yes, no := true, false
	if FormatOptionalYesNo(nil) != "—" || FormatOptionalYesNo(&yes) != "Yes" || FormatOptionalYesNo(&no) != "No" {
		t.Fatalf("FormatOptionalYesNo: got %q/%q/%q", FormatOptionalYesNo(nil), FormatOptionalYesNo(&yes), FormatOptionalYesNo(&no))
	}
	if IsTrue(nil) || IsTrue(&no) || !IsTrue(&yes) {
		t.Fatal("IsTrue must be true only for a declared true")
	}
}
