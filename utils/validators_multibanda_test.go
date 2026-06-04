package utils_test

import (
	"testing"
	"time"

	"github.com/abisaidfarias/lbtechapi/utils"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
)

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

func TestValidateMultibandaEvaluationTypesRejectsMutuallyExclusiveValues(t *testing.T) {
	err := enums.ValidateMultibandaEvaluationTypes([]string{
		enums.MultibandaEvalSAEMultibandaCertificate,
		enums.MultibandaEvalSAEOnlyCMASTest,
	})
	if err == nil {
		t.Fatal("expected error for mutually exclusive evaluation types")
	}
}

func TestValidateMultibandaEvaluationTypesAllowsSismateCombination(t *testing.T) {
	err := enums.ValidateMultibandaEvaluationTypes([]string{
		enums.MultibandaEvalSAEMultibandaCertificate,
		enums.MultibandaEvalSismatePeru,
	})
	if err != nil {
		t.Fatalf("expected valid combination, got %v", err)
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
		PlanningDate:    time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
	}
}
