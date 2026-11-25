package main

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// The filter must only touch closed controls with no verdict, so a rerun is a
// no-op and verdicts set by the new flow (including rejected) are kept.
func TestPendingFilterOnlyMatchesClosedWithoutVerdict(t *testing.T) {
	f := pendingFilter()
	if f["current_phase"] != phaseCompleted {
		t.Fatalf("current_phase filter = %v, want %d", f["current_phase"], phaseCompleted)
	}
	or, ok := f["$or"].(bson.A)
	if !ok || len(or) != 3 {
		t.Fatalf("$or = %v", f["$or"])
	}
	want := []bson.M{
		{"verdict": bson.M{"$exists": false}},
		{"verdict": nil},
		{"verdict": ""},
	}
	for i, w := range want {
		gotRaw, _ := bson.Marshal(or[i])
		wantRaw, _ := bson.Marshal(w)
		if string(gotRaw) != string(wantRaw) {
			t.Errorf("$or[%d] = %v, want %v", i, or[i], w)
		}
	}
}

func TestMigrationSetsApproved(t *testing.T) {
	set, ok := migrationUpdate()["$set"].(bson.M)
	if !ok || set["verdict"] != "approved" || len(set) != 1 {
		t.Fatalf("update = %v", migrationUpdate())
	}
}
