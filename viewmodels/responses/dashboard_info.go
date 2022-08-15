package responses

type DashboardInfo struct {
	CompanyName          string `json:"company_name" bson:"company_name"`
	LogoImage            string `json:"logo_image" bson:"logo_image"`
	TotalOngoing         int    `json:"total_ongoing" bson:"total_ongoing"`
	TotalPlanning        int    `json:"total_planning" bson:"total_planning"`
	TotalSampleReception int    `json:"total_sample" bson:"total_sample"`
	TotalFinished        int    `json:"total_finished" bson:"total_finished"`
	Month                string `json:"month" bson:"month"`
}
