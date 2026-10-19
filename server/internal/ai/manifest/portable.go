package manifest

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"

	agentcontract "github.com/leamout/contracts/agent"
)

var agentAliasPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

const maxPortableAgents = 32
const maxPortableTools = 128

func (v *Validator) ValidatePortable(pkg PortablePackage) error {
	if len(pkg.Agents) == 0 {
		return fmt.Errorf("package must contain at least one agent")
	}
	if len(pkg.Agents) > maxPortableAgents {
		return fmt.Errorf("package cannot contain more than %d agents", maxPortableAgents)
	}
	if len(pkg.Tools) > maxPortableTools {
		return fmt.Errorf("package cannot contain more than %d tools", maxPortableTools)
	}

	aliases := make(map[string]struct{}, len(pkg.Agents))
	names := make(map[string]struct{}, len(pkg.Agents))
	for alias, agent := range pkg.Agents {
		if !agentAliasPattern.MatchString(alias) {
			return fmt.Errorf("invalid agent alias %q", alias)
		}
		if err := v.Validate(Package{
			Agent: agent,
			Tools: pkg.Tools,
		}); err != nil {
			return fmt.Errorf("agent %q: %w", alias, err)
		}
		name := strings.ToLower(strings.TrimSpace(agent.Name))
		if _, ok := names[name]; ok {
			return fmt.Errorf("duplicate agent name %q", agent.Name)
		}
		names[name] = struct{}{}
		aliases[alias] = struct{}{}
	}
	if err := pkg.Routing.Validate(); err != nil {
		return fmt.Errorf("routing: %w", err)
	}
	if err := pkg.Routing.ValidateAgentReferences(aliases); err != nil {
		return fmt.Errorf("routing: %w", err)
	}
	return nil
}

// ReadPortable reads the same directory layout maintained in leamout/agents.
// The file stems in agents/*.json are the routing aliases.
func ReadPortable(fsys fs.FS) (PortablePackage, error) {
	var pkg PortablePackage
	pkg.Agents = make(map[string]agentcontract.Manifest)

	if err := readPortableJSON(fsys, "tools.json", &pkg.Tools); err != nil {
		return PortablePackage{}, err
	}
	if err := readPortableJSON(fsys, "routing.json", &pkg.Routing); err != nil {
		return PortablePackage{}, err
	}

	entries, err := fs.ReadDir(fsys, "agents")
	if err != nil {
		return PortablePackage{}, fmt.Errorf("read agents directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".json" {
			return PortablePackage{}, fmt.Errorf("unexpected agents entry %q", entry.Name())
		}
		alias := strings.TrimSuffix(entry.Name(), ".json")
		if !agentAliasPattern.MatchString(alias) {
			return PortablePackage{}, fmt.Errorf("invalid agent alias %q", alias)
		}
		var agent agentcontract.Manifest
		if err := readPortableJSON(fsys, path.Join("agents", entry.Name()), &agent); err != nil {
			return PortablePackage{}, err
		}
		pkg.Agents[alias] = agent
	}
	if err := NewValidator().ValidatePortable(pkg); err != nil {
		return PortablePackage{}, err
	}
	return pkg, nil
}

func readPortableJSON(fsys fs.FS, name string, target any) error {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}
	if !json.Valid(data) {
		return fmt.Errorf("%s: invalid JSON", name)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}
	return nil
}
