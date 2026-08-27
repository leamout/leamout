package organization

import (
	"time"

	"github.com/google/uuid"
)

type Response struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type UpdateRequest struct {
	Slug *string `json:"slug"`
	Name *string `json:"name"`
}
