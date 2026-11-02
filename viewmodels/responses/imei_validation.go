package responses

type ImeiLookupResult struct {
	IMEI           string `json:"imei"`
	TAC            string `json:"tac"`
	Brand          string `json:"brand"`
	TechnicalModel string `json:"technical_model"`
	CommercialName string `json:"commercial_name"`
	Source         string `json:"source"`
	QueriedAt      string `json:"queried_at"`
	ErrorMessage   string `json:"error_message,omitempty"`
}

type ImeiQueryHistoryItem struct {
	ID             string `json:"id"`
	IMEI           string `json:"imei"`
	TAC            string `json:"tac"`
	Brand          string `json:"brand"`
	TechnicalModel string `json:"technical_model"`
	CommercialName string `json:"commercial_name"`
	Source         string `json:"source"`
	ErrorMessage   string `json:"error_message,omitempty"`
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	QueriedAt      string `json:"queried_at"`
}

type ImeiQueryHistoryPage struct {
	Items    []ImeiQueryHistoryItem `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}
