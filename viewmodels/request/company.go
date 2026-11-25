package request

// Company is the create/update payload. Required fields are validated in the
// service (utils.ValidateCompanyRequest) so the messages are explicit.
// email and logo_url are no longer used and are ignored if sent.
type Company struct {
	ID                     string `bson:"_id" json:"_id"`
	Name                   string `json:"name"`
	Address                string `json:"address"`
	ClientID               string `json:"client_id"`
	Rut                    string `json:"rut"`
	RazonSocial            string `json:"razon_social"`
	AllowSmrWithoutTesting bool   `json:"allow_smr_without_testing"`
}
