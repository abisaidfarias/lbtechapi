package functions

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/abisaidfarias/lbtechapi/config"
	"github.com/abisaidfarias/lbtechapi/models"
)

type PasswordActionEmailData struct {
	Title         string
	Name          string
	BodyText      string
	ActionButton  template.HTML
	ExpiresHours  int
}

const passwordActionTemplatePath = "utils/htmlMessageTemplate/password_action.html"

func FrontendBaseURL() string {
	base := strings.TrimSpace(os.Getenv("FRONTEND_BASE_URL"))
	return strings.TrimRight(base, "/")
}

func PasswordSetURL(rawToken string) string {
	return FrontendBaseURL() + "/auth/set-password?token=" + strings.TrimSpace(rawToken)
}

func PasswordTokenTTLHours() int {
	const defaultHours = 48
	s := strings.TrimSpace(os.Getenv("PASSWORD_TOKEN_TTL_HOURS"))
	if s == "" {
		return defaultHours
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return defaultHours
	}
	return n
}

func RenderPasswordActionEmail(name, purpose, rawToken string) ([]byte, string, error) {
	setURL := PasswordSetURL(rawToken)
	hours := PasswordTokenTTLHours()

	var title, body, buttonLabel string
	switch purpose {
	case models.PasswordTokenPurposeInvite:
		title = "Activate your account"
		body = "Your account has been created. Use the button below to set your password and sign in."
		buttonLabel = "Set password"
	case models.PasswordTokenPurposeReset:
		title = "Reset your password"
		body = "We received a request to reset your password. Use the button below to choose a new one."
		buttonLabel = "Reset password"
	default:
		return nil, "", fmt.Errorf("unknown password email purpose: %s", purpose)
	}

	data := PasswordActionEmailData{
		Title:        title,
		Name:         strings.TrimSpace(name),
		BodyText:     body,
		ActionButton: RenderOutlookEmailButton(setURL, buttonLabel, EmailButtonColorPrimary),
		ExpiresHours: hours,
	}
	if data.Name == "" {
		data.Name = "there"
	}

	t, err := parseNotificationEmailTemplate(passwordActionTemplatePath)
	if err != nil {
		return nil, "", err
	}
	var htmlBuf bytes.Buffer
	if err := t.Execute(&htmlBuf, data); err != nil {
		return nil, "", err
	}
	html := htmlBuf.Bytes()
	html = bytes.ReplaceAll(html, []byte("\r\n"), []byte("\n"))
	html = bytes.ReplaceAll(html, []byte("\n"), []byte("\r\n"))
	return html, title, nil
}

func SendPasswordActionEmail(toEmail, name, purpose, rawToken string) error {
	if strings.TrimSpace(toEmail) == "" {
		return fmt.Errorf("recipient email is required")
	}
	if FrontendBaseURL() == "" {
		return fmt.Errorf("FRONTEND_BASE_URL is not configured")
	}
	html, subject, err := RenderPasswordActionEmail(name, purpose, rawToken)
	if err != nil {
		return err
	}
	subjectLine := "Subject: " + subject
	recipients := resolvePasswordEmailRecipients(toEmail)
	if err := sendBrandedHTMLEmail(recipients, subjectLine, html, nil, ""); err != nil {
		log.Printf("password email send failed for %s: %v", toEmail, err)
		return err
	}
	log.Printf("password email accepted by SMTP: intended=%s smtp_to=%v subject=%q from=%s",
		toEmail, recipients, subject, config.GetValue("EMAIL_FROM"))
	return nil
}

// resolvePasswordEmailRecipients returns SMTP To addresses. Optional DEV_EMAIL_TO redirects
// all password emails when USE_LOCAL_ENV=true (see .env.example).
func resolvePasswordEmailRecipients(intended string) []string {
	intended = strings.TrimSpace(intended)
	if intended == "" {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("USE_LOCAL_ENV")), "true") {
		if override := strings.TrimSpace(os.Getenv("DEV_EMAIL_TO")); override != "" {
			log.Printf("password email DEV_EMAIL_TO redirect: %s -> %s", intended, override)
			return []string{override}
		}
	}
	return []string{intended}
}

func PasswordTokenExpiresAt(now time.Time) time.Time {
	return now.Add(time.Duration(PasswordTokenTTLHours()) * time.Hour)
}
