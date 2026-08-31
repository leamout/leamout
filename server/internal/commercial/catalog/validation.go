package catalog

import (
	"strings"
	"unicode"

	"github.com/google/uuid"
)

func normalizeID(id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrIDRequired
	}
	return nil
}

func normalizeCode(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrCodeRequired
	}
	if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", ErrInvalidCode
	}
	return value, nil
}

func normalizeName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrNameRequired
	}
	return value, nil
}

func normalizeDescription(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func normalizeCreateProduct(input CreateProductInput) (CreateProductInput, error) {
	code, err := normalizeCode(input.Code)
	if err != nil {
		return CreateProductInput{}, err
	}
	name, err := normalizeName(input.Name)
	if err != nil {
		return CreateProductInput{}, err
	}
	input.Code = code
	input.Name = name
	input.Description = normalizeDescription(input.Description)
	return input, nil
}

func normalizeUpdateProduct(input UpdateProductInput) (UpdateProductInput, error) {
	if input.Code == nil && input.Name == nil && input.Description == nil && input.Active == nil {
		return UpdateProductInput{}, ErrNoChanges
	}
	if input.Code != nil {
		code, err := normalizeCode(*input.Code)
		if err != nil {
			return UpdateProductInput{}, err
		}
		input.Code = &code
	}
	if input.Name != nil {
		name, err := normalizeName(*input.Name)
		if err != nil {
			return UpdateProductInput{}, err
		}
		input.Name = &name
	}
	input.Description = normalizeDescription(input.Description)
	return input, nil
}

func normalizeCreatePlan(input CreatePlanInput) (CreatePlanInput, error) {
	if err := normalizeID(input.ProductID); err != nil {
		return CreatePlanInput{}, err
	}
	code, err := normalizeCode(input.Code)
	if err != nil {
		return CreatePlanInput{}, err
	}
	name, err := normalizeName(input.Name)
	if err != nil {
		return CreatePlanInput{}, err
	}
	input.Code = code
	input.Name = name
	input.Description = normalizeDescription(input.Description)
	return input, nil
}

func normalizeUpdatePlan(input UpdatePlanInput) (UpdatePlanInput, error) {
	if input.Code == nil && input.Name == nil && input.Description == nil && input.Active == nil {
		return UpdatePlanInput{}, ErrNoChanges
	}
	if input.Code != nil {
		code, err := normalizeCode(*input.Code)
		if err != nil {
			return UpdatePlanInput{}, err
		}
		input.Code = &code
	}
	if input.Name != nil {
		name, err := normalizeName(*input.Name)
		if err != nil {
			return UpdatePlanInput{}, err
		}
		input.Name = &name
	}
	input.Description = normalizeDescription(input.Description)
	return input, nil
}
