package enums

var HomologationStatus_value = map[string]int{
	"IN_PROGRESS": 0,
	"APPROVED":    1,
	"REJECTED":    2,
}
var HomologationType_value = map[string]int{
	"INITIAL":     0,
	"MAINTENANCE": 1,
	"REGRETION":   2,
}
var HomologationPhase_value = map[string]int{
	"PLANNING":         0,
	"SAMPLE_RECEPTION": 1,
	"TEST":             2,
	"UNDER_EVALUATION": 3,
	"COMPLETE":         4,
}
var TestResult_value = map[string]int{
	"NORUN": 0,
	"NA":    1,
	"PASS":  2,
	"FAIL":  3,
}

var TestFailureFrequency_value = map[string]int{
	"ALWAYS": 0,
	"RANDOM": 1,
	"ONCE":   2,
}

var TestFailureSeverity_value = map[string]int{
	"HIGH":   0,
	"MEDIUM": 1,
	"LOW":    2,
}
