package enums

var HomologationStatus_value = map[string]int{
	"IN_PROGRESS": 0,
	"APPROVED":    1,
	"REJECTED":    2,
}
var HomologationType_value = map[string]int{
	"INITIAL":     0,
	"MAINTENANCE": 1,
	"REGRESSION":  2,
}
var HomologationType_key = map[int]string{
	0: "Initial",
	1: "Maintenance",
	2: "Regression",
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
var MimeTypes_value = map[string]string{
	"image/gif":                ".gif",
	"text/html; charset=utf-8": ".html",
	"image/jpeg":               ".jpg",
	"application/json":         ".json",
	"application/pdf":          ".pdf",
	"image/png":                ".png",
	"image/svg+xml":            ".svg",
	"text/xml; charset=utf-8":  ".xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   ".docx",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
}
var HomologationStatus_type = map[int]string{
	0: "Ongoing",
	1: "Approved",
	2: "Rejected",
	3: "Finished",
}
