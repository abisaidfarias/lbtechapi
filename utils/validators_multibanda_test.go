package utils_test

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

func TestValidateMultibandaUpdateRequestSuccess(t *testing.T) {
	req := validMultibandaRequest()
	req.CurrentPhase = 2

	if err := utils.ValidateMultibandaUpdateRequest(req); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestValidateMultibandaUpdateRequestRejectsInvalidPhase(t *testing.T) {
	req := validMultibandaRequest()
	req.CurrentPhase = 99

	if err := utils.ValidateMultibandaUpdateRequest(req); !utils.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestValidateMultibandaUpdateRequestAllowsArcotelEcuador(t *testing.T) {
	req := validMultibandaRequest()
	req.CurrentPhase = 2
	req.EvaluationType = []string{enums.MultibandaEvalArcotelEcuador}

	if err := utils.ValidateMultibandaUpdateRequest(req); err != nil {
		t.Fatalf("expected valid update request with arcotel_ecuador, got %v", err)
	}
}

func TestValidateMultibandaUpdateRequestRejectsMultipleEvaluationTypes(t *testing.T) {
	req := validMultibandaRequest()
	req.CurrentPhase = 2
	req.EvaluationType = []string{
		enums.MultibandaEvalSAEMultibandaCertificate,
		enums.MultibandaEvalArcotelEcuador,
	}

	if err := utils.ValidateMultibandaUpdateRequest(req); !utils.IsValidationError(err) {
		t.Fatalf("expected validation error for multiple evaluation types on update, got %v", err)
	}
}

func TestValidateMultibandaCreateRequestSuccess(t *testing.T) {
	req := validMultibandaRequest()

	if err := utils.ValidateMultibandaCreateRequest(req); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestValidateMultibandaCreateRequestRejectsNonPlanningPhase(t *testing.T) {
	req := validMultibandaRequest()
	req.CurrentPhase = 1

	err := utils.ValidateMultibandaCreateRequest(req)
	if !utils.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestValidateMultibandaCreateRequestRejectsSampleDatesOnCreate(t *testing.T) {
	req := validMultibandaRequest()
	now := time.Now()
	req.SampleStartDate = &now

	err := utils.ValidateMultibandaCreateRequest(req)
	if !utils.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestValidateMultibandaEvaluationTypesRejectsEmpty(t *testing.T) {
	err := enums.ValidateMultibandaEvaluationTypes(nil)
	if err == nil {
		t.Fatal("expected error for empty evaluation_type")
	}
}

func TestValidateMultibandaEvaluationTypesRejectsMultipleValues(t *testing.T) {
	err := enums.ValidateMultibandaEvaluationTypes([]string{
		enums.MultibandaEvalSAEMultibandaCertificate,
		enums.MultibandaEvalSAEOnlyCMASTest,
	})
	if err == nil {
		t.Fatal("expected error for multiple evaluation types")
	}
}

func TestValidateMultibandaEvaluationTypesAllowsArcotelEcuador(t *testing.T) {
	err := enums.ValidateMultibandaEvaluationTypes([]string{enums.MultibandaEvalArcotelEcuador})
	if err != nil {
		t.Fatalf("expected valid arcotel evaluation type, got %v", err)
	}
}

func TestValidateMultibandaCreateRequestRequiresReflashURLWhenNeeded(t *testing.T) {
	req := validMultibandaRequest()
	req.NeedReflash = true

	if err := utils.ValidateMultibandaCreateRequest(req); !utils.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}

	req.CommentsReflash = "https://example.com/reflash"
	if err := utils.ValidateMultibandaCreateRequest(req); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestValidateMultibandaCreateRequestRejectsReflashURLWithoutFlag(t *testing.T) {
	req := validMultibandaRequest()
	req.CommentsReflash = "https://example.com/reflash"

	if err := utils.ValidateMultibandaCreateRequest(req); !utils.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func validMultibandaRequest() *request.Multibanda {
	return &request.Multibanda{
		Company:         "507f1f77bcf86cd799439011",
		Device:          "507f1f77bcf86cd799439012",
		Brand:           "507f1f77bcf86cd799439013",
		Type:            enums.MultibandaTypeInitialProcess,
		EvaluationType:  []string{enums.MultibandaEvalSAEMultibandaCertificate},
		SoftwareVersion: "SW 1.0",
		OsVersion:       "17",
		CurrentPhase:    enums.MultibandaPhasePlanning,
		Status:          enums.HomologationStatus_value["IN_PROGRESS"],
		PlanningDate:    time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
	}
}
