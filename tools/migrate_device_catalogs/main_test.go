package main

import (
	"reflect"
	"testing"
)

// Seeding keeps the configuration order, trims, and drops empty values and
// case-insensitive duplicates (first occurrence wins).
func TestSeedValues(t *testing.T) {
	got := seedValues([]string{" USB-C ", "", "MicroUSB", "usb-c", "  ", "Lightning", "MICROUSB"})
	want := []string{"USB-C", "MicroUSB", "Lightning"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("seedValues = %v, want %v", got, want)
	}
}

func TestTestLikeReason(t *testing.T) {
	for _, v := range []string{"rrr", "2222", "yyy", "123", "cap", "Memory", "ROM", "ram", "usb", "device", "test 1", "x"} {
		if testLikeReason(v) == "" {
			t.Errorf("%q should look like test data", v)
		}
	}
	for _, v := range []string{"128GB", "8 GB", "Smartphone", "USB-C", "MicroUSB", "512GB", "IoT"} {
		if r := testLikeReason(v); r != "" {
			t.Errorf("%q flagged as test data (%s)", v, r)
		}
	}
}
