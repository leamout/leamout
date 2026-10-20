package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	agentcontract "github.com/leamout/contracts/agent"
	"github.com/leamout/leamout/server/internal/ai/agents"
	"github.com/leamout/leamout/server/internal/ai/tools"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Deployer struct {
	validator *Validator
	agents    *agents.Service
	tools     *tools.Service
}

func NewDeployer(
	agentService *agents.Service,
	toolService *tools.Service,
	validator *Validator,
) *Deployer {
	if validator == nil {
		validator = NewValidator()
	}
	return &Deployer{
		validator: validator,
		agents:    agentService,
		tools:     toolService,
	}
}

func (d *Deployer) Deploy(
	ctx context.Context,
	organizationID uuid.UUID,
	pkg Package,
	input DeploymentInput,
) (DeploymentResult, error) {
	if organizationID == uuid.Nil {
		return DeploymentResult{}, fmt.Errorf("organization id is required")
	}
	if d == nil || d.agents == nil || d.tools == nil {
		return DeploymentResult{}, fmt.Errorf("agent manifest deployer is unavailable")
	}
	if err := d.validator.Validate(pkg); err != nil {
		return DeploymentResult{}, err
	}

	bindings := make([]agents.ProviderBindingRequest, 0, len(pkg.Agent.Providers))
	for _, binding := range pkg.Agent.Providers {
		var integrationID *uuid.UUID
		if value := input.Integrations[binding.Provider]; value != uuid.Nil {
			integrationID = &value
		}
		bindings = append(bindings, agents.ProviderBindingRequest{
			Role:          string(binding.Role),
			Provider:      binding.Provider,
			IntegrationID: integrationID,
			Config:        append(json.RawMessage(nil), binding.Config...),
		})
	}

	voice := optionalString(pkg.Agent.Voice)
	language := optionalString(pkg.Agent.Language)
	created, err := d.agents.Create(ctx, organizationID, agents.CreateRequest{
		Name:               pkg.Agent.Name,
		Engine:             string(pkg.Agent.Engine),
		Instructions:       pkg.Agent.Instructions,
		Voice:              voice,
		Language:           language,
		EngineConfig:       json.RawMessage(`{}`),
		Bindings:           bindings,
		InterruptionPolicy: string(pkg.Agent.InterruptionPolicy),
		RecordingPolicy:    string(pkg.Agent.RecordingPolicy),
	})
	if err != nil {
		return DeploymentResult{}, err
	}

	result := DeploymentResult{
		AgentID:        created.ID,
		SigningSecrets: make(map[string]string),
	}
	for _, definition := range pkg.Tools {
		req, err := toolCreateRequest(definition, input.Variables)
		if err != nil {
			_ = d.agents.Disable(ctx, organizationID, created.ID)
			return DeploymentResult{}, err
		}
		_, secret, err := d.tools.Create(ctx, organizationID, created.ID, req)
		if err != nil {
			_ = d.agents.Disable(ctx, organizationID, created.ID)
			return DeploymentResult{}, err
		}
		if secret != "" {
			result.SigningSecrets[definition.Name] = secret
		}
	}
	return result, nil
}

func toolCreateRequest(
	definition agentcontract.ToolDefinition,
	variables map[string]string,
) (tools.CreateRequest, error) {
	var endpoint *string
	if definition.EndpointURL != "" {
		value, err := resolveValue(definition.EndpointURL, variables)
		if err != nil {
			return tools.CreateRequest{}, fmt.Errorf(
				"resolve tool %q endpoint: %w",
				definition.Name,
				err,
			)
		}
		endpoint = &value
	}
	return tools.CreateRequest{
		Type:        string(definition.Type),
		Name:        definition.Name,
		Description: definition.Description,
		Parameters:  append(json.RawMessage(nil), definition.Parameters...),
		EndpointURL: endpoint,
		TimeoutMS:   definition.TimeoutMS,
		Enabled:     definition.Enabled,
	}, nil
}

func resolveValue(value string, variables map[string]string) (string, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return value, nil
	}
	name := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
	resolved := strings.TrimSpace(variables[name])
	if resolved == "" {
		return "", fmt.Errorf("deployment variable %q is required", name)
	}
	return resolved, nil
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
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
