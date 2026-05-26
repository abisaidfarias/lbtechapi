package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

var ErrorForbidden = errors.New("forbidden")

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

func ValidateMultibandaCreateRequest(multibanda *request.Multibanda) error {
	if multibanda.CurrentPhase != enums.MultibandaPhasePlanning {
		return NewValidationError("current_phase must be 0 (Planning) on create")
	}

	if multibanda.PlanningDate.IsZero() {
		return NewValidationError("planning_date is required")
	}

	if hasDateValue(multibanda.SampleStartDate) {
		return NewValidationError("sample_start_date must be null on create")
	}

	if hasDateValue(multibanda.SampleEndDate) {
		return NewValidationError("sample_end_date must be null on create")
	}

	if multibanda.Company == "" {
		return NewValidationError("company is required")
	}

	if multibanda.Device == "" {
		return NewValidationError("device is required")
	}

	if multibanda.Brand == "" {
		return NewValidationError("brand is required")
	}

	if multibanda.Type == "" {
		return NewValidationError("type is required")
	}

	if err := enums.ValidateMultibandaType(multibanda.Type); err != nil {
		return NewValidationError(err.Error())
	}

	if err := enums.ValidateMultibandaEvaluationTypes(multibanda.EvaluationType); err != nil {
		return NewValidationError(err.Error())
	}

	return nil
}

func hasDateValue(value *time.Time) bool {
	return value != nil && !value.IsZero()
}

func ValidateObjectIDField(fieldName, value string) (string, error) {
	if value == "" {
		return "", NewValidationError(fmt.Sprintf("%s is required", fieldName))
	}
	if len(value) != 24 {
		return "", NewValidationError(fmt.Sprintf("invalid %s id", fieldName))
	}
	return value, nil
}
