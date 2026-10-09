package entitlements

import (
	"time"

	"github.com/google/uuid"
)

type Capability string

const (
	CapabilitySSO               Capability = "sso"
	CapabilitySCIM              Capability = "scim"
	CapabilityAdvancedRBAC      Capability = "advanced_rbac"
	CapabilityRetentionPolicies Capability = "retention_policies"
	CapabilityPrivateNetworking Capability = "private_networking"
)

var capabilities = []Capability{
	CapabilitySSO,
	CapabilitySCIM,
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
