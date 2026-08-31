package catalog

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNormalizeCreateProduct(t *testing.T) {
	description := "  Programmable voice  "
	input, err := normalizeCreateProduct(CreateProductInput{
		Code:        "  voice-pro  ",
		Name:        "  Voice Pro  ",
		Description: &description,
	})
	if err != nil {
		t.Fatalf("normalizeCreateProduct() error = %v", err)
	}
	if input.Code != "voice-pro" {
		t.Fatalf("Code = %q, want %q", input.Code, "voice-pro")
	}
	if input.Name != "Voice Pro" {
		t.Fatalf("Name = %q, want %q", input.Name, "Voice Pro")
	}
	if input.Description == nil || *input.Description != "Programmable voice" {
		t.Fatalf("Description = %v, want normalized description", input.Description)
	}
}

func TestNormalizeCodeRejectsWhitespace(t *testing.T) {
	_, err := normalizeCode("voice pro")
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("normalizeCode() error = %v, want %v", err, ErrInvalidCode)
	}
}

func TestNormalizeUpdateProductRequiresChange(t *testing.T) {
	_, err := normalizeUpdateProduct(UpdateProductInput{})
	if !errors.Is(err, ErrNoChanges) {
		t.Fatalf("normalizeUpdateProduct() error = %v, want %v", err, ErrNoChanges)
	}
}

func TestNormalizeCreatePlanRequiresProduct(t *testing.T) {
	_, err := normalizeCreatePlan(CreatePlanInput{Code: "pro", Name: "Pro"})
	if !errors.Is(err, ErrIDRequired) {
		t.Fatalf("normalizeCreatePlan() error = %v, want %v", err, ErrIDRequired)
	}

	_, err = normalizeCreatePlan(CreatePlanInput{ProductID: uuid.New(), Code: "pro", Name: "Pro"})
	if err != nil {
		t.Fatalf("normalizeCreatePlan() unexpected error = %v", err)
	}
}
