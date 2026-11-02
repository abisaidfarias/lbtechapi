package imei_test

import (
	"testing"

	imeiutil "github.com/abisaidfarias/lbtechapi/utils/imei"
)

func TestIsValidLuhnKnownIMEI(t *testing.T) {
	if !imeiutil.IsValidLuhn("350105094917765") {
		t.Fatal("expected valid IMEI")
	}
	if imeiutil.IsValidLuhn("350105094917764") {
		t.Fatal("expected invalid check digit")
	}
}
