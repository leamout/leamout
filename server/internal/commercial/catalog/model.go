package catalog

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.New("catalog product not found")
	ErrPlanNotFound    = errors.New("catalog plan not found")
	ErrCodeRequired    = errors.New("catalog code is required")
	ErrInvalidCode     = errors.New("catalog code must not contain whitespace")
	ErrIDRequired      = errors.New("catalog id is required")
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
