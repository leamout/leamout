package catalog

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.New("catalog product not found")
	ErrPlanNotFound    = errors.New("catalog plan not found")
	ErrProductInactive = errors.New("catalog product is inactive")
	ErrCodeConflict    = errors.New("catalog code already exists")
	ErrCodeRequired    = errors.New("catalog code is required")
	ErrInvalidCode     = errors.New("catalog code must not contain whitespace")
	ErrNameRequired    = errors.New("catalog name is required")
	ErrIDRequired      = errors.New("catalog id is required")
	ErrNoChanges       = errors.New("at least one catalog field is required")
)

// Product is a commercial product family that groups reusable plans.
type Product struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description *string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Plan is a reusable commercial offer within a product.
type Plan struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	Code        string
	Name        string
	Description *string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateProductInput describes a new reusable commercial product family.
type CreateProductInput struct {
	Code        string
	Name        string
	Description *string
	Active      *bool
}

// UpdateProductInput describes mutable product fields.
type UpdateProductInput struct {
	Code        *string
	Name        *string
	Description *string
	Active      *bool
}

// CreatePlanInput describes a new reusable commercial offer.
type CreatePlanInput struct {
	ProductID   uuid.UUID
	Code        string
	Name        string
	Description *string
	Active      *bool
}

// UpdatePlanInput describes mutable plan fields. A plan cannot move between products.
type UpdatePlanInput struct {
	Code        *string
	Name        *string
	Description *string
	Active      *bool
}
