package responses

type DashboardChart struct {
	Countries               []string            `json:"countries" bson:"countries"`
	Times                   []string            `json:"times" bson:"times"`
	StackedSerieChart       []StackedSerieChart `json:"stacked_serie" bson:"stacked_serie"`
	PieSerieChart           map[string]int      `json:"pie_serie_chart" bson:"pie_serie_chart"`
	CertificationSerieChart map[string][3]int   `json:"certification_serie_chart" bson:"certification_serie_chart"`
}
