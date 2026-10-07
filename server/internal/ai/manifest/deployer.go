package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	agentcontract "github.com/leamout/contracts/agent"
	"github.com/leamout/leamout/server/internal/ai/agents"
	"github.com/leamout/leamout/server/internal/ai/tools"
)

type DeploymentInput struct {
	Integrations map[string]uuid.UUID
	Variables    map[string]string
}

type DeploymentResult struct {
	AgentID        uuid.UUID
	SigningSecrets map[string]string
}

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
		integrationID := input.Integrations[binding.Provider]
		if integrationID == uuid.Nil {
			return DeploymentResult{}, fmt.Errorf(
				"integration id is required for provider %q",
				binding.Provider,
			)
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
