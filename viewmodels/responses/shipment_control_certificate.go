package responses

type ShipmentControlCertificate struct {
	URL            string `json:"url"`
	RegistroOABI   string `json:"registro_oabi"`
	ControlNumber  string `json:"control_number"`
	Regenerated    bool   `json:"regenerated"`
}
