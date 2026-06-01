package functions

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
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

func TestFormatMultibandaOsVersionPrefixesOsVersionViewWhenMissingPlatform(t *testing.T) {
	got := FormatMultibandaOsVersion("17", "Android", "17")
	if got != "Android 17" {
		t.Fatalf("expected Android 17, got %q", got)
	}
}

func TestMultibandaReflashURLForEmailOnlyOnCreate(t *testing.T) {
	notify := &request.MultibandaNotify{
		NeedReflash:     true,
		CommentsReflash: "https://example.com/reflash",
	}
	if got := multibandaReflashURLForEmail(utils.CREATE, notify.NeedReflash, notify.CommentsReflash); got != notify.CommentsReflash {
		t.Fatalf("expected reflash url on create, got %q", got)
	}
	if got := multibandaReflashURLForEmail(utils.PHASE, notify.NeedReflash, notify.CommentsReflash); got != "" {
		t.Fatalf("expected empty reflash url on phase email, got %q", got)
	}
}

func TestFormatMultibandaSarValue(t *testing.T) {
	if got := FormatMultibandaSarValue(0); got != "—" {
		t.Fatalf("expected dash for zero SAR, got %q", got)
	}
	if got := FormatMultibandaSarValue(1.25); got != "1.25 W/Kg" {
		t.Fatalf("expected 1.25 W/Kg, got %q", got)
	}
	if got := FormatMultibandaSarValue(1.09); got != "1.09 W/Kg" {
		t.Fatalf("expected 1.09 W/Kg, got %q", got)
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
