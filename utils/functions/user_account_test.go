package functions

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

func TestEffectiveAccountStatus(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		user   *responses.AuthUser
		expect string
	}{
		{"explicit pending", &responses.AuthUser{AccountStatus: models.AccountStatusPending}, models.AccountStatusPending},
		{"legacy empty hash", &responses.AuthUser{}, models.AccountStatusPending},
		{"legacy with hash", &responses.AuthUser{PasswordHash: "hash"}, models.AccountStatusActive},
		{"explicit active", &responses.AuthUser{AccountStatus: models.AccountStatusActive, PasswordHash: ""}, models.AccountStatusActive},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := EffectiveAccountStatus(tc.user)
			if got != tc.expect {
				t.Fatalf("got %q want %q", got, tc.expect)
			}
		})
	}
}
