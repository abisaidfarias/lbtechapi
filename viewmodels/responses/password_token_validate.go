package responses

type PasswordTokenValidate struct {
	Valid     bool   `json:"valid"`
	Purpose   string `json:"purpose,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
}
