package functions

import (
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func SetMultibandaDatesToNull(multibanda *responses.MultibandaExpanded) {
	if multibanda.PlanningDate != nil && multibanda.PlanningDate.IsZero() {
		multibanda.PlanningDate = nil
	}
	if multibanda.SampleStartDate != nil && multibanda.SampleStartDate.IsZero() {
		multibanda.SampleStartDate = nil
	}
	if multibanda.SampleEndDate != nil && multibanda.SampleEndDate.IsZero() {
		multibanda.SampleEndDate = nil
	}
	if multibanda.TestStartDate != nil && multibanda.TestStartDate.IsZero() {
		multibanda.TestStartDate = nil
	}
	if multibanda.TestEndDate != nil && multibanda.TestEndDate.IsZero() {
		multibanda.TestEndDate = nil
	}
	if multibanda.UnderStartDate != nil && multibanda.UnderStartDate.IsZero() {
		multibanda.UnderStartDate = nil
	}
	if multibanda.UnderEndDate != nil && multibanda.UnderEndDate.IsZero() {
		multibanda.UnderEndDate = nil
	}
	if multibanda.CompletedDate != nil && multibanda.CompletedDate.IsZero() {
		multibanda.CompletedDate = nil
	}
	if multibanda.CreatedDate != nil && multibanda.CreatedDate.IsZero() {
		multibanda.CreatedDate = nil
	}
}

func EnrichMultibandaExpanded(multibanda *responses.MultibandaExpanded) {
	if multibanda.OsVersionView == "" && multibanda.Device.PlatformOs != "" {
		multibanda.OsVersionView = multibanda.Device.PlatformOs + " " + multibanda.OsVersion
	}
	SetMultibandaDatesToNull(multibanda)
}

// ApplyMultibandaPhaseDateRules applies derived milestone dates during phase changes.
func ApplyMultibandaPhaseDateRules(multibanda *models.Multibanda, existing *responses.MultibandaExpanded) {
	if multibanda == nil || multibanda.CurrentPhase != enums.HomologationPhase_value["UNDER_EVALUATION"] {
		return
	}

	testEndDate := multibanda.TestEndDate
	if testEndDate.IsZero() && existing != nil && existing.TestEndDate != nil {
		testEndDate = *existing.TestEndDate
	}

	if testEndDate.IsZero() {
		return
	}

	if multibanda.TestEndDate.IsZero() {
		multibanda.TestEndDate = testEndDate
	}
	multibanda.UnderStartDate = testEndDate
}
