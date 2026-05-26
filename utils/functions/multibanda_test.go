package functions

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func TestFormatMultibandaOsVersionPrefersOsVersionView(t *testing.T) {
	got := FormatMultibandaOsVersion("Android 17", "Android", "17")
	if got != "Android 17" {
		t.Fatalf("expected Android 17, got %q", got)
	}
}

func TestFormatMultibandaOsVersionBuildsFromPlatformAndVersion(t *testing.T) {
	got := FormatMultibandaOsVersion("", "Android", "17")
	if got != "Android 17" {
		t.Fatalf("expected Android 17, got %q", got)
	}
}

func TestApplyMultibandaPhaseDateRulesSetsUnderStartFromTestEndOnUnderEvaluation(t *testing.T) {
	testEnd := time.Date(2026, 5, 29, 6, 0, 0, 0, time.UTC)

	multibanda := &models.Multibanda{
		CurrentPhase: enums.HomologationPhase_value["UNDER_EVALUATION"],
		TestEndDate:  testEnd,
	}

	ApplyMultibandaPhaseDateRules(multibanda, nil)

	if !multibanda.UnderStartDate.Equal(testEnd) {
		t.Fatalf("expected under_start_date %v, got %v", testEnd, multibanda.UnderStartDate)
	}
}

func TestApplyMultibandaPhaseDateRulesUsesExistingTestEndWhenRequestIsEmpty(t *testing.T) {
	testEnd := time.Date(2026, 5, 29, 6, 0, 0, 0, time.UTC)

	multibanda := &models.Multibanda{
		CurrentPhase: enums.HomologationPhase_value["UNDER_EVALUATION"],
	}
	existing := &responses.MultibandaExpanded{
		TestEndDate: &testEnd,
	}

	ApplyMultibandaPhaseDateRules(multibanda, existing)

	if !multibanda.UnderStartDate.Equal(testEnd) {
		t.Fatalf("expected under_start_date %v, got %v", testEnd, multibanda.UnderStartDate)
	}
	if !multibanda.TestEndDate.Equal(testEnd) {
		t.Fatalf("expected test_end_date %v, got %v", testEnd, multibanda.TestEndDate)
	}
}

func TestApplyMultibandaPhaseDateRulesSkipsOtherPhases(t *testing.T) {
	testEnd := time.Date(2026, 5, 29, 6, 0, 0, 0, time.UTC)

	multibanda := &models.Multibanda{
		CurrentPhase: enums.HomologationPhase_value["TEST"],
		TestEndDate:  testEnd,
	}

	ApplyMultibandaPhaseDateRules(multibanda, nil)

	if !multibanda.UnderStartDate.IsZero() {
		t.Fatalf("expected empty under_start_date, got %v", multibanda.UnderStartDate)
	}
}
