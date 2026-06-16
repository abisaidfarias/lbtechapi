package utils_test

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
)

func TestNormalizeCalendarDateUTCKeepsCalendarDay(t *testing.T) {
	input := time.Date(2026, 6, 19, 6, 0, 0, 0, time.UTC)
	got := utils.NormalizeCalendarDateUTC(input)
	if got.Year() != 2026 || got.Month() != time.June || got.Day() != 19 {
		t.Fatalf("expected 2026-06-19, got %v", got)
	}
	if got.Hour() != 12 || got.Location() != time.UTC {
		t.Fatalf("expected noon UTC, got %v", got)
	}
}

func TestNormalizeOptionalCalendarDateUTCNilWhenEmpty(t *testing.T) {
	if got := utils.NormalizeOptionalCalendarDateUTC(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}
