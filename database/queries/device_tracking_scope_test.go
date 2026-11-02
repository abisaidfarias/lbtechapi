package queries

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

// matchBody returns the conditions of the pipeline's $match stage, or nil when
// the pipeline has none and therefore returns every document.
func matchBody(t *testing.T, pipeline []bson.D) bson.D {
	t.Helper()
	for _, stage := range pipeline {
		if stageName(stage) != "$match" {
			continue
		}
		body, ok := stage[0].Value.(bson.D)
		if !ok {
			t.Fatalf("unexpected $match body %T", stage[0].Value)
		}
		return body
	}
	return nil
}

// companyEquals reports whether the conditions pin company._id to the company.
func companyEquals(conditions bson.D, companyID primitive.ObjectID) bool {
	for _, condition := range conditions {
		if condition.Key == "company._id" && condition.Value == companyID {
			return true
		}
	}
	return false
}

// Each case is run against both tracking queries: the list and advanced search
// have to apply the same scope.
func trackingQueries() map[string]func(bool, primitive.ObjectID, []primitive.ObjectID, []primitive.ObjectID) []bson.D {
	return map[string]func(bool, primitive.ObjectID, []primitive.ObjectID, []primitive.ObjectID) []bson.D{
		"list": func(isInternal bool, companyID primitive.ObjectID, brands, companies []primitive.ObjectID) []bson.D {
			return GetDeviceTrackingExpanded(isInternal, companyID, brands, nil, companies)
		},
		"advanced search": func(isInternal bool, companyID primitive.ObjectID, brands, companies []primitive.ObjectID) []bson.D {
			return GetDeviceTrackingAdvancedSearch(companyID, isInternal, &request.SearchOption{}, brands, nil, companies)
		},
	}
}

// An external user with no brands, countries or clients must still be limited
// to their company: without any other filter the pipeline used to have no
// $match at all and returned every company's IMEIs.
func TestTrackingQueriesLimitExternalWithoutFiltersToCompany(t *testing.T) {
	companyID := primitive.NewObjectID()

	for name, query := range trackingQueries() {
		t.Run(name, func(t *testing.T) {
			conditions := matchBody(t, query(false, companyID, nil, nil))
			if !companyEquals(conditions, companyID) {
				t.Fatalf("expected a company filter, got %v", conditions)
			}
		})
	}
}

func TestTrackingQueriesKeepCompanyFilterForExternalWithFilters(t *testing.T) {
	companyID := primitive.NewObjectID()
	brands := []primitive.ObjectID{primitive.NewObjectID()}

	for name, query := range trackingQueries() {
		t.Run(name, func(t *testing.T) {
			conditions := matchBody(t, query(false, companyID, brands, nil))
			if !companyEquals(conditions, companyID) {
				t.Fatalf("expected a company filter, got %v", conditions)
			}
		})
	}
}

// A user with no company assigned must match nothing rather than fall back to
// an unfiltered pipeline.
func TestTrackingQueriesFilterExternalWithoutCompany(t *testing.T) {
	for name, query := range trackingQueries() {
		t.Run(name, func(t *testing.T) {
			conditions := matchBody(t, query(false, primitive.NilObjectID, nil, nil))
			if !companyEquals(conditions, primitive.NilObjectID) {
				t.Fatalf("expected a company filter, got %v", conditions)
			}
		})
	}
}

func TestTrackingQueriesLeaveInternalWithoutFiltersUnchanged(t *testing.T) {
	for name, query := range trackingQueries() {
		t.Run(name, func(t *testing.T) {
			if conditions := matchBody(t, query(true, primitive.NewObjectID(), nil, nil)); conditions != nil {
				t.Fatalf("expected no filter for an internal user, got %v", conditions)
			}
		})
	}
}

func TestTrackingQueriesDoNotPinInternalToOwnCompany(t *testing.T) {
	companyID := primitive.NewObjectID()
	clients := []primitive.ObjectID{primitive.NewObjectID()}

	for name, query := range trackingQueries() {
		t.Run(name, func(t *testing.T) {
			conditions := matchBody(t, query(true, companyID, nil, clients))
			if len(conditions) != 1 {
				t.Fatalf("expected only the clients filter, got %v", conditions)
			}
			if companyEquals(conditions, companyID) {
				t.Fatalf("internal user was pinned to their own company: %v", conditions)
			}
		})
	}
}
