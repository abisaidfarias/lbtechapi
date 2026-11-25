package main

import (
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestPhaseStatesFor(t *testing.T) {
	cases := map[int]map[string]string{
		0: {"planning": "pending", "sample_reception": "pending", "testing": "pending", "under_evaluation": "pending", "complete": "pending"},
		3: {"planning": "done", "sample_reception": "done", "testing": "done", "under_evaluation": "pending", "complete": "pending"},
		// Closed requests (approved or rejected) are done everywhere.
		4: {"planning": "done", "sample_reception": "done", "testing": "done", "under_evaluation": "done", "complete": "done"},
	}
	for phase, want := range cases {
		if got := PhaseStatesFor(phase); !reflect.DeepEqual(got, want) {
			t.Errorf("phase %d: %v, want %v", phase, got, want)
		}
	}
}

// The pipeline expression must follow the same rule as PhaseStatesFor.
func TestPhaseStatesExprMatchesRule(t *testing.T) {
	expr := phaseStatesExpr()
	for i, name := range phaseNames {
		cond := expr[name].(bson.M)["$cond"].(bson.A)
		test := cond[0].(bson.M)
		if i == completePhase {
			if !reflect.DeepEqual(test, bson.M{"$gte": bson.A{"$current_phase", completePhase}}) {
				t.Errorf("complete test = %v", test)
			}
		} else if !reflect.DeepEqual(test, bson.M{"$lt": bson.A{i, "$current_phase"}}) {
			t.Errorf("%s test = %v", name, test)
		}
		if cond[1] != "done" || cond[2] != "pending" {
			t.Errorf("%s branches = %v", name, cond)
		}
	}
}

// Idempotency: only zero (pre-1971) dates and documents without
// phase_states are matched, so a migrated database matches nothing.
func TestFiltersOnlyMatchUnmigratedData(t *testing.T) {
	f := zeroDateFilter("completed_date")["completed_date"].(bson.M)
	if f["$type"] != "date" || !f["$lt"].(time.Time).Equal(zeroCutoff) || zeroCutoff.Year() != 1971 {
		t.Errorf("zero filter = %v", f)
	}
	if !reflect.DeepEqual(missingPhaseStatesFilter(), bson.M{"phase_states": bson.M{"$exists": false}}) {
		t.Errorf("phase_states filter = %v", missingPhaseStatesFilter())
	}
}
