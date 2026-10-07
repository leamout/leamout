package manifest

import (
	"encoding/json"
	"fmt"
	"strings"

	agentcontract "github.com/leamout/contracts/agent"
	"github.com/leamout/contracts/ai"
	aicatalog "github.com/leamout/leamout/server/internal/ai/catalog"
)

type Package struct {
	Agent agentcontract.Manifest
	Tools []agentcontract.ToolDefinition
}

type Validator struct {
	catalog *aicatalog.Catalog
}

func NewValidator(catalogs ...*aicatalog.Catalog) *Validator {
	var catalog *aicatalog.Catalog
	if len(catalogs) != 0 {
		catalog = catalogs[0]
	}
	if catalog == nil {
		catalog, _ = aicatalog.Builtins()
	}
	return &Validator{catalog: catalog}
}

func (v *Validator) Validate(pkg Package) error {
	if err := pkg.Agent.Validate(); err != nil {
		return err
	}
	if err := agentcontract.ValidateTools(pkg.Tools); err != nil {
		return err
	}
	if v == nil || v.catalog == nil {
		return fmt.Errorf("AI provider catalog is unavailable")
	}

	toolNames := make(map[string]struct{}, len(pkg.Tools))
	for _, tool := range pkg.Tools {
		toolNames[strings.TrimSpace(tool.Name)] = struct{}{}
	}
	for _, name := range pkg.Agent.Tools {
		if _, ok := toolNames[strings.TrimSpace(name)]; !ok {
			return fmt.Errorf("agent references unknown tool %q", name)
		}
	}

	for _, binding := range pkg.Agent.Providers {
		descriptor, ok := v.catalog.Get(binding.Role, binding.Provider)
		if !ok {
			return fmt.Errorf(
				"provider %q does not support role %q",
				binding.Provider,
				binding.Role,
			)
		}
		if descriptor.Kind != binding.Role {
			return fmt.Errorf(
				"provider %q kind %q does not match role %q",
				binding.Provider,
				descriptor.Kind,
				binding.Role,
			)
		}
		if len(binding.Config) == 0 {
			continue
		}
		var object map[string]json.RawMessage
		if err := json.Unmarshal(binding.Config, &object); err != nil || object == nil {
			return fmt.Errorf(
				"provider %q config must be a JSON object",
				binding.Provider,
			)
		}
		validator, ok := v.catalog.ConfigValidator(
			ai.Kind(binding.Role),
			binding.Provider,
		)
		if !ok {
			continue
		}
		if err := validator.ValidateConfig(binding.Config); err != nil {
			return fmt.Errorf(
				"invalid %s:%s config: %w",
				binding.Role,
				binding.Provider,
				err,
			)
		}
	}
	return nil
}
