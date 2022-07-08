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
var DashBoardPhase_value = map[string]int{
	"PLANNING":         0,
	"SAMPLE_RECEPTION": 1,
	"ONGOING":          2,
	"COMPLETE":         3,
}
var HomologationPhase_key = map[int]string{
	0: "Planning Date",
	1: "Sample Reception Date",
	2: "Testing Start Date",
	3: "Testing End Date",
	4: "Homologation Process Finishes",
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
var ExcelHomologationHeaders = map[string]string{
	"A1": "Client",
	"B1": "Country",
	"C1": "Brand",
	"D1": "Commercial Model",
	"E1": "Technical Model",
	"F1": "Os Version",
	"G1": "Approval Type",
	"H1": "Test Plan",
	"I1": "Test Start Date",
	"J1": "Test End Date",
	"K1": "Project Type",
	"L1": "Status",
}
var ExcelDeviceTrackinHeaders = map[string]string{
	"A1": "Imei",
	"B1": "Brand",
	"C1": "Commercial Model",
	"D1": "Client",
	"E1": "Country",
	"F1": "Location",
	"G1": "Responsible",
	"H1": "External Responsible",
	"I1": "Tracking Date",
}
var ExcelFailsHeaders = map[string]string{
	"A6": "Code",
	"B6": "Issue Overview",
	"C6": "Actual Result",
	"D6": "Expected Result",
	"E6": "Steps To Reproduce",
	"F6": "Issue Frecuency",
	"G6": "Issue Severity",
	"H6": "Hiperlinks",
}
var TestFailureSeverity_key = map[int]string{
	0: "High",
	1: "Medium",
	2: "Low",
}
var TestFailureFrequency_key = map[int]string{
	0: "Always",
	1: "Random",
	2: "Once",
}
var NotificationType_value = map[string]int{
	"COMPANY":  0,
	"INTERNAL": 1,
	"MANUAL":   2,
}

var HomologationTemplatePath = map[string]string{
	"Create": "utils/htmlMessageTemplate/createHomologation.html",
	"Phase":  "utils/htmlMessageTemplate/createHomologation.html",
}
