package queries

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// stageName returns the operator of a pipeline stage, e.g. "$lookup".
func stageName(stage bson.D) string {
	if len(stage) == 0 {
		return ""
	}
	return stage[0].Key
}

// The list must join the automatic report so the UI can colour its icon.
func TestGetMultibandasJoinsAutomaticReport(t *testing.T) {
	pipeline := GetMultibandas(nil, nil, true, primitive.NilObjectID)

	var lookup, addFields, project bson.D
	for _, stage := range pipeline {
		switch stageName(stage) {
		case "$lookup":
			if from := lookupFrom(stage); from == "multibanda_reports" {
				lookup = stage
			}
		case "$addFields":
			addFields = stage
		case "$project":
			project = stage
		}
	}

	if lookup == nil {
		t.Fatal("expected a $lookup against multibanda_reports")
	}
	if addFields == nil {
		t.Fatal("expected an $addFields projecting report_status")
	}
	if project == nil {
		t.Fatal("expected a $project dropping the joined array")
	}

	// The join must be on multibanda, not _id, or every row would match nothing.
	spec := lookup[0].Value.(bson.D)
	assertField(t, spec, "localField", "_id")
	assertField(t, spec, "foreignField", "multibanda")
	assertField(t, spec, "as", "automatic_report")

	fields := addFields[0].Value.(bson.D)
	if fields[0].Key != "report_status" {
		t.Errorf("added field: got %q, want report_status", fields[0].Key)
	}

	// The temporary join array must not leak into the response.
	dropped := project[0].Value.(bson.D)
	assertField(t, dropped, "automatic_report", int32(0))
}

// The status join must run before the sort so ordering is unaffected.
func TestReportStatusStagesRunBeforeSort(t *testing.T) {
	pipeline := GetMultibandas(nil, nil, true, primitive.NilObjectID)

	sortIdx, statusIdx := -1, -1
	for i, stage := range pipeline {
		switch stageName(stage) {
		case "$sort":
			sortIdx = i
		case "$addFields":
			statusIdx = i
		}
	}
	if statusIdx == -1 || sortIdx == -1 {
		t.Fatal("expected both the status stage and the sort stage")
	}
	if statusIdx > sortIdx {
		t.Errorf("report_status is added at %d, after the sort at %d", statusIdx, sortIdx)
	}
}

func lookupFrom(stage bson.D) string {
	spec, ok := stage[0].Value.(bson.D)
	if !ok {
		return ""
	}
	for _, e := range spec {
		if e.Key == "from" {
			if s, ok := e.Value.(string); ok {
				return s
			}
		}
	}
	return ""
}

func assertField(t *testing.T, doc bson.D, key string, want interface{}) {
	t.Helper()
	for _, e := range doc {
		if e.Key != key {
			continue
		}
		switch w := want.(type) {
		case string:
			if got, _ := e.Value.(string); got != w {
				t.Errorf("%s: got %q, want %q", key, got, w)
			}
		case int32:
			if got, _ := e.Value.(int); int32(got) != w {
				t.Errorf("%s: got %v, want %v", key, e.Value, w)
			}
		}
		return
	}
	t.Errorf("missing field %q", key)
}
