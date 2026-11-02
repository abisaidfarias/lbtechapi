package request

type ForgotPassword struct {
	Email string `json:"email" binding:"required,email"`
}
