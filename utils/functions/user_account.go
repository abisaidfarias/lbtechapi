package functions

import (
	"strings"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func EffectiveAccountStatus(user *responses.AuthUser) string {
	if user == nil {
		return models.AccountStatusActive
	}
	status := strings.TrimSpace(user.AccountStatus)
	if status != "" {
		return status
	}
	if strings.TrimSpace(user.PasswordHash) == "" {
		return models.AccountStatusPending
	}
	return models.AccountStatusActive
}

func NormalizeListAccountStatus(status string, passwordHash string) string {
	status = strings.TrimSpace(status)
	if status != "" {
		return status
	}
	if strings.TrimSpace(passwordHash) == "" {
		return models.AccountStatusPending
	}
	return models.AccountStatusActive
}
