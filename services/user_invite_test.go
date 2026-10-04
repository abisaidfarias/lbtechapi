package services

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func TestUserPendingInvitation(t *testing.T) {
	setAt := time.Now()
	cases := []struct {
		name string
		user *responses.User
		want bool
	}{
		{"pending status", &responses.User{AccountStatus: models.AccountStatusPending}, true},
		{"empty status no password set", &responses.User{}, true},
		{"active status", &responses.User{AccountStatus: models.AccountStatusActive}, false},
		{"password set at", &responses.User{AccountStatus: models.AccountStatusPending, PasswordSetAt: &setAt}, false},
		{"nil user", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := userPendingInvitation(tc.user); got != tc.want {
				t.Fatalf("userPendingInvitation() = %v, want %v", got, tc.want)
			}
		})
	}
}
