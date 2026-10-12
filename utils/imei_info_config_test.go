package utils

import "testing"

func TestResolveImeiInfoServiceID_OnlyBasic(t *testing.T) {
	t.Setenv("IMEI_INFO_SERVICE", "11")
	if got := ResolveImeiInfoServiceID(); got != BasicImeiCheckServiceID {
		t.Fatalf("got %q, want %q", got, BasicImeiCheckServiceID)
	}
}

func TestResolveImeiInfoServiceID_Default(t *testing.T) {
	t.Setenv("IMEI_INFO_SERVICE", "")
	if got := ResolveImeiInfoServiceID(); got != BasicImeiCheckServiceID {
		t.Fatalf("got %q, want %q", got, BasicImeiCheckServiceID)
	}
}
