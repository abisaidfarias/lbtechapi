package responses

type DeleteProcessResult struct {
	Deleted       bool `json:"deleted,omitempty"`
	RequestDelete bool `json:"request_delete,omitempty"`
}
