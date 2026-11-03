package queries

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// The filter must only match a process without a report date, so a later
// generation can never overwrite the first one.
func TestSetMultibandaReportDateOnlyMatchesUnsetDate(t *testing.T) {
	oid := primitive.NewObjectID()
	date := time.Date(2026, 11, 3, 15, 4, 5, 0, time.FixedZone("CLT", -3*3600))

	filter, update := SetMultibandaReportDate(oid, date)

	if filter["_id"] != oid {
		t.Fatalf("filter does not target the process: %v", filter)
	}
	if value, ok := filter["report_date"]; !ok || value != nil {
		t.Fatalf("filter must require an unset report_date, got %v", filter)
	}

	set, ok := update[0].Value.(primitive.D)
	if update[0].Key != "$set" || !ok || len(set) != 1 || set[0].Key != "report_date" {
		t.Fatalf("unexpected update %v", update)
	}
	stored, ok := set[0].Value.(time.Time)
	if !ok || !stored.Equal(date) || stored.Location() != time.UTC {
		t.Errorf("expected %v stored in UTC, got %v", date, set[0].Value)
	}
}
