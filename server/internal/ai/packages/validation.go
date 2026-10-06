package packages

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Validate verifies portable package invariants without resolving tenant resources.
func Validate(manifest Manifest) error {
	if strings.TrimSpace(manifest.SchemaVersion) != SchemaVersionV1 {
		return fmt.Errorf("schema_version must be %q", SchemaVersionV1)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(manifest.Agents) == 0 {
		return fmt.Errorf("at least one agent is required")
	}

	toolAliases := make(map[string]struct{}, len(manifest.Tools))
	for i, tool := range manifest.Tools {
		alias := strings.TrimSpace(tool.Alias)
		if alias == "" {
			return fmt.Errorf("tools[%d].alias is required", i)
		}
		if _, exists := toolAliases[alias]; exists {
			return fmt.Errorf("duplicate tool alias %q", alias)
		}
		toolAliases[alias] = struct{}{}
		if strings.TrimSpace(tool.Type) == "" {
			return fmt.Errorf("tools[%d].type is required", i)
		}
		if strings.TrimSpace(tool.Name) == "" {
			return fmt.Errorf("tools[%d].name is required", i)
		}
		if err := validateJSONObject(tool.Parameters, false); err != nil {
			return fmt.Errorf("tools[%d].parameters: %w", i, err)
		}
	}

	agentAliases := make(map[string]struct{}, len(manifest.Agents))
	for i, agent := range manifest.Agents {
		alias := strings.TrimSpace(agent.Alias)
		if alias == "" {
			return fmt.Errorf("agents[%d].alias is required", i)
		}
		if _, exists := agentAliases[alias]; exists {
			return fmt.Errorf("duplicate agent alias %q", alias)
		}
		agentAliases[alias] = struct{}{}
		if err := validateAgent(i, agent, toolAliases); err != nil {
			return err
		}
	}

	if manifest.Routing != nil {
		if err := validateRouting(*manifest.Routing, agentAliases); err != nil {
			return err
		}
	}
	return nil
}

func validateAgent(index int, agent AgentDefinition, toolAliases map[string]struct{}) error {
	if strings.TrimSpace(agent.Name) == "" {
		return fmt.Errorf("agents[%d].name is required", index)
	}
	if strings.TrimSpace(agent.Instructions) == "" {
		return fmt.Errorf("agents[%d].instructions are required", index)
	}
	if agent.Engine != EngineComposable && agent.Engine != EngineRealtime {
		return fmt.Errorf("agents[%d].engine must be composable or realtime", index)
	}
	if err := validateJSONObject(agent.EngineConfig, true); err != nil {
		return fmt.Errorf("agents[%d].engine_config: %w", index, err)
	}

	roles := make(map[string]struct{}, len(agent.Providers))
	for providerIndex, binding := range agent.Providers {
		role := strings.TrimSpace(binding.Role)
		if role == "" {
			return fmt.Errorf("agents[%d].providers[%d].role is required", index, providerIndex)
		}
		if _, exists := roles[role]; exists {
			return fmt.Errorf("agents[%d] has duplicate provider role %q", index, role)
		}
		roles[role] = struct{}{}
		if strings.TrimSpace(binding.Provider) == "" {
			return fmt.Errorf("agents[%d].providers[%d].provider is required", index, providerIndex)
		}
		if err := validateJSONObject(binding.Config, true); err != nil {
			return fmt.Errorf("agents[%d].providers[%d].config: %w", index, providerIndex, err)
		}
	}

	if err := validateEngineRoles(index, agent.Engine, roles); err != nil {
		return err
	}
	seenTools := make(map[string]struct{}, len(agent.Tools))
	for _, alias := range agent.Tools {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			return fmt.Errorf("agents[%d] contains an empty tool alias", index)
		}
		if _, exists := seenTools[alias]; exists {
			return fmt.Errorf("agents[%d] references tool %q more than once", index, alias)
		}
		seenTools[alias] = struct{}{}
		if _, exists := toolAliases[alias]; !exists {
			return fmt.Errorf("agents[%d] references unknown tool %q", index, alias)
		}
	}
	return nil
}

func validateEngineRoles(index int, engine string, roles map[string]struct{}) error {
	required := []string{RoleSTT, RoleLLM, RoleTTS}
	allowed := map[string]struct{}{RoleSTT: {}, RoleLLM: {}, RoleTTS: {}}
	if engine == EngineRealtime {
		required = []string{RoleRealtime}
		allowed = map[string]struct{}{RoleRealtime: {}}
	}
	for _, role := range required {
		if _, exists := roles[role]; !exists {
			return fmt.Errorf("agents[%d] engine %q requires provider role %q", index, engine, role)
		}
	}
	for role := range roles {
		if _, exists := allowed[role]; !exists {
			return fmt.Errorf("agents[%d] engine %q does not allow provider role %q", index, engine, role)
		}
	}
	return nil
}

func validateRouting(routing RoutingDefinition, agentAliases map[string]struct{}) error {
	defaultAgent := strings.TrimSpace(routing.DefaultAgent)
	if defaultAgent == "" {
		return fmt.Errorf("routing.default_agent is required")
	}
	if _, exists := agentAliases[defaultAgent]; !exists {
		return fmt.Errorf("routing.default_agent references unknown agent %q", defaultAgent)
	}
	for i, rule := range routing.Rules {
		agent := strings.TrimSpace(rule.Agent)
		if agent == "" {
			return fmt.Errorf("routing.rules[%d].agent is required", i)
		}
		if _, exists := agentAliases[agent]; !exists {
			return fmt.Errorf("routing.rules[%d] references unknown agent %q", i, agent)
		}
		if err := validateJSONObject(rule.When, false); err != nil {
			return fmt.Errorf("routing.rules[%d].when: %w", i, err)
		}
	}
	return nil
}

func validateJSONObject(value json.RawMessage, optional bool) error {
	if len(value) == 0 {
		if optional {
			return nil
		}
		return fmt.Errorf("must be a JSON object")
	}
	if !json.Valid(value) {
		return fmt.Errorf("must be valid JSON")
	}
	var object map[string]any
	if err := json.Unmarshal(value, &object); err != nil || object == nil {
		return fmt.Errorf("must be a JSON object")
	}
	return nil
}
