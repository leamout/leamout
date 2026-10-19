package manifest

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	agentcontract "github.com/leamout/contracts/agent"
	"github.com/leamout/leamout/server/pkg/apperror"
)

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

func (d *Deployer) DeployPortable(
	ctx context.Context,
	organizationID uuid.UUID,
	pkg PortablePackage,
	input DeploymentInput,
) (PortableDeploymentResult, error) {
	if organizationID == uuid.Nil {
		return PortableDeploymentResult{}, apperror.NewBadRequest("organization context required")
	}
	if d == nil || d.agents == nil || d.tools == nil || d.validator == nil {
		return PortableDeploymentResult{}, apperror.NewServiceUnavailable("agent package deployer is unavailable", nil)
	}
	if err := d.validator.ValidatePortable(pkg); err != nil {
		return PortableDeploymentResult{}, apperror.NewBadRequest(err.Error())
	}

	// Resolve and validate every referenced tool before any database writes.
	// This catches missing variables and invalid resolved endpoints up front.
	prepared, err := preparePortableTools(pkg, input.Variables)
	if err != nil {
		return PortableDeploymentResult{}, apperror.NewBadRequest(err.Error())
	}

	aliases := make([]string, 0, len(pkg.Agents))
	for alias := range pkg.Agents {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)

	result := PortableDeploymentResult{
		Agents: make(map[string]DeploymentResult, len(aliases)),
		RoutingPlan: RoutingPlan{
			Manifest: pkg.Routing,
			AgentIDs: make(map[string]uuid.UUID, len(aliases)),
		},
	}
	for _, alias := range aliases {
		created, deployErr := d.Deploy(
			ctx,
			organizationID,
			Package{
				Agent: pkg.Agents[alias],
				Tools: prepared[alias],
			},
			DeploymentInput{
				Integrations: input.Integrations,
			},
		)
		if deployErr != nil {
			// Disable agents installed earlier in this request, preserving
			// their IDs for an administrator to inspect and clean up.
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			var cleanupError error
			for _, previous := range aliases {
				item, ok := result.Agents[previous]
				if !ok {
					continue
				}
				if disableErr := d.agents.Disable(cleanupCtx, organizationID, item.AgentID); disableErr != nil {
					cleanupError = fmt.Errorf("disable %q: %w", previous, disableErr)
				}
			}
			cancel()
			if cleanupError != nil {
				return PortableDeploymentResult{}, fmt.Errorf(
					"deploy agent %q: %w; cleanup failed: %w",
					alias,
					deployErr,
					cleanupError,
				)
			}
			return PortableDeploymentResult{}, fmt.Errorf("deploy agent %q: %w", alias, deployErr)
		}
		result.Agents[alias] = created
		result.RoutingPlan.AgentIDs[alias] = created.AgentID
	}
	return result, nil
}

func preparePortableTools(
	pkg PortablePackage,
	variables map[string]string,
) (map[string][]agentcontract.ToolDefinition, error) {
	definitions := make(map[string]agentcontract.ToolDefinition, len(pkg.Tools))
	for _, definition := range pkg.Tools {
		definitions[strings.TrimSpace(definition.Name)] = definition
	}

	prepared := make(map[string][]agentcontract.ToolDefinition, len(pkg.Agents))
	for alias, agent := range pkg.Agents {
		selected := make([]agentcontract.ToolDefinition, 0, len(agent.Tools))
		for _, name := range agent.Tools {
			definition := definitions[strings.TrimSpace(name)]
			if definition.EndpointURL != "" {
				resolved, err := resolveValue(definition.EndpointURL, variables)
				if err != nil {
					return nil, fmt.Errorf("agent %q: %w", alias, err)
				}
				definition.EndpointURL = resolved
				if err := definition.Validate(); err != nil {
					return nil, fmt.Errorf("agent %q tool %q: %w", alias, name, err)
				}
			}
			selected = append(selected, definition)
		}
		prepared[alias] = selected
	}
	return prepared, nil
}
