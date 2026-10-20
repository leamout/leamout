package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	agentcontract "github.com/leamout/contracts/agent"
	"github.com/leamout/leamout/server/internal/ai/agents"
	"github.com/leamout/leamout/server/internal/ai/providers"
	"github.com/leamout/leamout/server/internal/ai/tools"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Deployer struct {
	validator *Validator
	agents    *agents.Service
	tools     *tools.Service
	providers *providers.Service
	db        *pgxpool.Pool
}

func NewDeployer(
	agentService *agents.Service,
	toolService *tools.Service,
	providerService *providers.Service,
	db *pgxpool.Pool,
	validator *Validator,
) *Deployer {
	if validator == nil {
		validator = NewValidator()
	}
	return &Deployer{
		validator: validator,
		agents:    agentService,
		tools:     toolService,
		providers: providerService,
		db:        db,
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
	if d == nil || d.agents == nil || d.tools == nil || d.providers == nil || d.db == nil {
		return DeploymentResult{}, fmt.Errorf("agent manifest deployer is unavailable")
	}
	if err := d.validator.Validate(pkg); err != nil {
		return DeploymentResult{}, err
	}
	if err := d.validateIntegrations(ctx, organizationID, pkg.Agent.Providers, input); err != nil {
		return DeploymentResult{}, err
	}
	scoped, tx, err := d.begin(ctx)
	if err != nil {
		return DeploymentResult{}, err
	}
	defer rollbackDeployment(ctx, tx)
	result, err := scoped.deploy(ctx, organizationID, pkg, input)
	if err != nil {
		return DeploymentResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DeploymentResult{}, apperror.NewInternal("commit agent package deployment", err)
	}
	return result, nil
}

func (d *Deployer) deploy(
	ctx context.Context,
	organizationID uuid.UUID,
	pkg Package,
	input DeploymentInput,
) (DeploymentResult, error) {
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
			return DeploymentResult{}, err
		}
		_, secret, err := d.tools.Create(ctx, organizationID, created.ID, req)
		if err != nil {
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
	if d == nil || d.agents == nil || d.tools == nil || d.providers == nil || d.db == nil || d.validator == nil {
		return PortableDeploymentResult{}, apperror.NewServiceUnavailable("agent package deployer is unavailable", nil)
	}
	if err := d.validator.ValidatePortable(pkg); err != nil {
		return PortableDeploymentResult{}, apperror.NewBadRequest(err.Error())
	}

	// Resolve tools and validate every organization integration before writing.
	prepared, err := preparePortableTools(pkg, input.Variables)
	if err != nil {
		return PortableDeploymentResult{}, apperror.NewBadRequest(err.Error())
	}
	aliases := make([]string, 0, len(pkg.Agents))
	for alias := range pkg.Agents {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	for _, alias := range aliases {
		if err := d.validateIntegrations(ctx, organizationID, pkg.Agents[alias].Providers, input); err != nil {
			return PortableDeploymentResult{}, fmt.Errorf("agent %q: %w", alias, err)
		}
	}

	scoped, tx, err := d.begin(ctx)
	if err != nil {
		return PortableDeploymentResult{}, err
	}
	defer rollbackDeployment(ctx, tx)

	result := PortableDeploymentResult{
		Agents: make(map[string]DeploymentResult, len(aliases)),
		RoutingPlan: RoutingPlan{
			Manifest: pkg.Routing,
			AgentIDs: make(map[string]uuid.UUID, len(aliases)),
		},
	}
	for _, alias := range aliases {
		created, err := scoped.deploy(
			ctx,
			organizationID,
			Package{
				Agent: pkg.Agents[alias],
				Tools: prepared[alias],
			},
			DeploymentInput{Integrations: input.Integrations},
		)
		if err != nil {
			return PortableDeploymentResult{}, fmt.Errorf("deploy agent %q: %w", alias, err)
		}
		result.Agents[alias] = created
		result.RoutingPlan.AgentIDs[alias] = created.AgentID
	}
	if err := tx.Commit(ctx); err != nil {
		return PortableDeploymentResult{}, apperror.NewInternal("commit portable agent deployment", err)
	}
	return result, nil
}

// validateIntegrations verifies that every binding points to an existing
// organization-owned credential for the correct provider and role.
func (d *Deployer) validateIntegrations(
	ctx context.Context,
	organizationID uuid.UUID,
	bindings []agentcontract.ProviderBinding,
	input DeploymentInput,
) error {
	for _, binding := range bindings {
		credentialID := input.Integrations[binding.Provider]
		if credentialID == uuid.Nil {
			return apperror.NewBadRequest(fmt.Sprintf("AI integration for provider %q is required", binding.Provider))
		}
		if err := d.providers.ValidateBinding(
			ctx,
			organizationID,
			string(binding.Role),
			providers.UpsertBindingRequest{
				Provider: binding.Provider,
				CredentialID: &credentialID,
				Config: binding.Config,
			},
		); err != nil {
			return fmt.Errorf("provider %q: %w", binding.Provider, err)
		}
	}
	return nil
}

func (d *Deployer) begin(ctx context.Context) (*Deployer, pgx.Tx, error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return nil, nil, apperror.NewInternal("begin agent package deployment", err)
	}
	return &Deployer{
		validator: d.validator,
		agents: d.agents.WithTx(tx),
		tools: d.tools.WithTx(tx),
		providers: d.providers.WithTx(tx),
		db: d.db,
	}, tx, nil
}

func rollbackDeployment(ctx context.Context, tx pgx.Tx) {
	rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(rollbackCtx)
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
