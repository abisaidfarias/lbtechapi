package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	apputils "github.com/abisaidfarias/lbtechapi/utils"
	utils "github.com/abisaidfarias/lbtechapi/utils/errors"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/gin-gonic/gin"
)

func statusOf(err error) (int, string) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	handleErrorResponse(ctx, err)
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body["error"]
}

// Validation errors forwarded by any handler (e.g. the SMR guards on
// PUT /multibanda/:id/phase) must be 400 with their message, not 500.
func TestHandleErrorResponseValidationIs400(t *testing.T) {
	for _, msg := range []string{
		functions.SmrErrPendingValidation,
		functions.SmrErrNoTestingPhases,
		functions.SmrErrRejectedClosed,
		"multibanda has a pending delete request",
	} {
		code, got := statusOf(apputils.NewValidationError(msg))
		if code != http.StatusBadRequest || got != msg {
			t.Errorf("%q → %d %q", msg, code, got)
		}
		// Wrapped too.
		if code, _ := statusOf(fmt.Errorf("phase change: %w", apputils.NewValidationError(msg))); code != http.StatusBadRequest {
			t.Errorf("wrapped %q → %d", msg, code)
		}
	}
}

func TestHandleErrorResponseKeepsOtherMappings(t *testing.T) {
	if code, _ := statusOf(utils.ErrorInvalidCredentials); code != http.StatusUnauthorized {
		t.Errorf("invalid credentials → %d", code)
	}
	if code, _ := statusOf(utils.ErrorDuplicated); code != http.StatusBadRequest {
		t.Errorf("duplicated → %d", code)
	}
	if code, _ := statusOf(errors.New("boom")); code != http.StatusInternalServerError {
		t.Errorf("unexpected error → %d", code)
	}
}
