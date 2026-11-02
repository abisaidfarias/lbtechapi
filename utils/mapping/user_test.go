package mapping

import (
	"testing"

	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func TestUserRequestToPendingUser(t *testing.T) {
	user := UserRequestToPendingUser(&request.User{
		Email:    "a@b.com",
		Name:     "A",
		LastName: "B",
	})
	if user.AccountStatus != models.AccountStatusPending {
		t.Fatalf("status=%q", user.AccountStatus)
	}
	if user.PasswordHash != "" {
		t.Fatal("expected empty password hash")
	}
}
