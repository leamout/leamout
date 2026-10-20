package organizations

import "github.com/google/uuid"

const pageSize = 25

type Filter struct {
	Search string
	Status string
	Page   int32
}

type Organization struct {
	ID        uuid.UUID
	Name      string
	Status    string
	CreatedAt string
}

type List struct {
	Organizations []Organization
	Filter        Filter
	HasNext       bool
}

type Detail struct {
	Organization
	UpdatedAt          string
	DeletedAt          string
	ActiveMembers      int64
	Agents             int64
	Trunks             int64
	PlanName           string
	SubscriptionStatus string
	PeriodEnd          string
}
