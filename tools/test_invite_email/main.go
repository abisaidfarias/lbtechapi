package main

import (
	"fmt"
	"os"

	"github.com/abisaidfarias/lbtechapi/config"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
)

func main() {
	_, _ = config.LoadSecrets()
	to := "fariasabisaid@gmail.com"
	if len(os.Args) > 1 {
		to = os.Args[1]
	}
	fmt.Println("FRONTEND_BASE_URL:", functions.FrontendBaseURL())
	fmt.Println("EMAIL_FROM:", config.GetValue("EMAIL_FROM"))
	fmt.Println("SMTP:", config.GetValue("SMTP_CLIENTE")+":"+config.GetValue("EMAIL_PORT"))
	fmt.Println("SMTP user:", config.SMTPAuthUsername())
	fmt.Println("Sending test invite to:", to)
	err := functions.SendPasswordActionEmail(to, "Test User", models.PasswordTokenPurposeInvite, "test-token-not-valid")
	if err != nil {
		fmt.Println("SEND FAILED:", err)
		os.Exit(1)
	}
	fmt.Println("SEND OK (SMTP accepted message; check inbox/spam and Office365 Sent Items)")
}
