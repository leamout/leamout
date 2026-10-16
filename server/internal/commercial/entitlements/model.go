package entitlements

import (
	"time"

	"github.com/google/uuid"
)

type Capability string

const (
	CapabilityAdvancedRBAC      Capability = "advanced_rbac"
	CapabilityRetentionPolicies Capability = "retention_policies"
	CapabilityPrivateNetworking Capability = "private_networking"
)

var capabilities = []Capability{
	CapabilityAdvancedRBAC,
	CapabilityRetentionPolicies,
	CapabilityPrivateNetworking,
}

type Entitlement struct {
	OrganizationID uuid.UUID
	Capability     Capability
	Enabled        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type EffectiveEntitlement struct {
	Capability Capability `json:"capability"`
	Enabled    bool       `json:"enabled"`
}
