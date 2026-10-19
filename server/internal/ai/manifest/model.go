package manifest

import (
	"github.com/google/uuid"
	agentcontract "github.com/leamout/contracts/agent"
)

type Package struct {
	Agent agentcontract.Manifest
	Tools []agentcontract.ToolDefinition
}

// PortablePackage groups canonical contracts for a package directory or an
// API deployment request. It does not define another portable agent schema.
type PortablePackage struct {
	Agents  map[string]agentcontract.Manifest `json:"agents"`
	Tools   []agentcontract.ToolDefinition    `json:"tools"`
	Routing agentcontract.RoutingManifest     `json:"routing"`
}

type DeploymentInput struct {
	Integrations map[string]uuid.UUID `json:"integrations,omitempty"`
	Variables    map[string]string    `json:"variables,omitempty"`
}

type DeploymentResult struct {
	AgentID        uuid.UUID         `json:"agent_id"`
	SigningSecrets map[string]string `json:"signing_secrets,omitempty"`
}

// RoutingPlan is a validated, resolved plan. It does not activate a DID or
// install a schedule evaluator in the inbound telephony routing service.
type RoutingPlan struct {
	Manifest agentcontract.RoutingManifest `json:"manifest"`
	AgentIDs map[string]uuid.UUID          `json:"agent_ids"`
}

type PortableDeploymentResult struct {
	Agents      map[string]DeploymentResult `json:"agents"`
	RoutingPlan RoutingPlan                 `json:"routing_plan"`
}

type PortableDeploymentRequest struct {
	Package      PortablePackage      `json:"package"`
	Integrations map[string]uuid.UUID `json:"integrations,omitempty"`
	Variables    map[string]string    `json:"variables,omitempty"`
}

