package enums

import "fmt"

const MultibandaPhasePlanning = 0

const (
	MultibandaTypeInitialProcess = "initial_process"
	MultibandaTypeSMR            = "smr"
	MultibandaTypeMR             = "mr"
	MultibandaTypeOSUpgrade      = "os_upgrade"
	MultibandaTypePretestSAE     = "pretest_sae"
)

const (
	MultibandaEvalSAEMultibandaCertificate = "sae_multibanda_certificate"
	MultibandaEvalSAEOnlyCMASTest          = "sae_only_cmas_test"
	MultibandaEvalSismatePeru              = "sismate_peru"
	MultibandaEvalArcotelEcuador           = "arcotel_ecuador"
)

var allowedMultibandaTypes = map[string]struct{}{
	MultibandaTypeInitialProcess: {},
	MultibandaTypeSMR:            {},
	MultibandaTypeMR:             {},
	MultibandaTypeOSUpgrade:      {},
	MultibandaTypePretestSAE:     {},
}

var allowedMultibandaEvaluationTypes = map[string]struct{}{
	MultibandaEvalSAEMultibandaCertificate: {},
	MultibandaEvalSAEOnlyCMASTest:          {},
	MultibandaEvalSismatePeru:              {},
	MultibandaEvalArcotelEcuador:           {},
}

var AllowedMultibandaTypes = []string{
	MultibandaTypeInitialProcess,
	MultibandaTypeSMR,
	MultibandaTypeMR,
	MultibandaTypeOSUpgrade,
	MultibandaTypePretestSAE,
}

var AllowedMultibandaEvaluationTypes = []string{
	MultibandaEvalSAEMultibandaCertificate,
	MultibandaEvalSAEOnlyCMASTest,
	MultibandaEvalSismatePeru,
	MultibandaEvalArcotelEcuador,
}

var MultibandaTypeLabels = map[string]string{
	MultibandaTypeInitialProcess: "Initial Process",
	MultibandaTypeSMR:            "SMR",
	MultibandaTypeMR:             "MR",
	MultibandaTypeOSUpgrade:      "OS Upgrade",
	MultibandaTypePretestSAE:     "Pretest SAE",
}

var MultibandaEvaluationTypeLabels = map[string]string{
	MultibandaEvalSAEMultibandaCertificate: "SAE Multibanda Certificate",
	MultibandaEvalSAEOnlyCMASTest:          "SAE Only CMAS Test",
	MultibandaEvalSismatePeru:              "Sismate Peru",
	MultibandaEvalArcotelEcuador:           "ARCOTEL (Ecuador)",
}

var MultibandaPhaseLabels = map[int]string{
	0: "Planning",
	1: "Sample Reception",
	2: "Testing",
	3: "Under Evaluation",
	4: "Complete",
}

func ValidateMultibandaType(value string) error {
	if _, ok := allowedMultibandaTypes[value]; !ok {
		return fmt.Errorf("invalid type %q: must be one of %v", value, AllowedMultibandaTypes)
	}
	return nil
}

func ValidateMultibandaPhase(phase int) error {
	if _, ok := MultibandaPhaseLabels[phase]; !ok {
		return fmt.Errorf("invalid current_phase %d", phase)
	}
	return nil
}

func ValidateMultibandaEvaluationTypes(values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("evaluation_type must include exactly one value")
	}
	if len(values) > 1 {
		return fmt.Errorf("evaluation_type must include exactly one value (single selection)")
	}

	value := values[0]
	if _, ok := allowedMultibandaEvaluationTypes[value]; !ok {
		return fmt.Errorf("invalid evaluation_type value %q: must be one of %v", value, AllowedMultibandaEvaluationTypes)
	}

	return nil
}
