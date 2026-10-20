# Leamout ecosystem versioning

Leamout's repositories are synchronized by compatibility, not by forcing one shared version number.

## Dependency direction

```text
leamout/contracts
      ↓
leamout/ai-providers
      ↓
leamout/leamout
      ↓
leamout/agents
```

`leamout/agents` is a compatibility-tested catalog and is not a runtime dependency of `leamout/leamout`.

## Portable agent packages

The only portable JSON contract is defined by `github.com/leamout/contracts/agent`. A package is a directory of contract-defined documents, as used by `leamout/agents`:

| File | Canonical Go contract | Purpose |
| --- | --- | --- |
| `agents/<alias>.json` | `agent.Manifest` | One credential-free voice agent and its provider bindings |
| `tools.json` | `[]agent.ToolDefinition` | Shared built-in and webhook tools |
| `routing.json` | `agent.RoutingManifest` | Default agent, schedule rules and file-stem aliases |

Standalone examples may use `agent.json` for a single `agent.Manifest`. Manifests and routing documents use **numeric** `schema_version: 1`; provider bindings contain public provider configuration only. Organization credentials, integration IDs, tool secrets and deployment variables are resolved by the control plane when deploying, never shipped in a portable package.

The older, unused `server/internal/ai/packages` schema combined multiple agents, tools and routing into one object with string `"1"` for `schema_version`. It is not part of the portable contract and has been removed. Do not introduce another runtime-specific portable JSON model.

`server/internal/ai/manifest` validates the shared contracts, reads package directories, and deploys individual or multi-agent packages. The `leamout/agents` validator independently checks published templates, including tool and routing references.

### Package deployment

The organization-scoped `POST /agent-packages/deploy` endpoint (under the API router) accepts a JSON deployment envelope:

```json
{
  "package": {
    "agents": {
      "receptionist": { "schema_version": 1, "name": "Receptionist", "engine": "realtime", "instructions": "Answer calls.", "providers": [{ "role": "realtime", "provider": "gemini" }] }
    },
    "tools": [],
    "routing": { "schema_version": 1, "name": "Receptionist routing", "default_agent": "receptionist" }
  },
  "integrations": {},
  "variables": {}
}
```

The `agents` keys match the file stems in `agents/*.json` from the portable directory. `tools` and `routing` are the same canonical contract documents. The request must be authenticated with organization `voice-agents:write` access; it is limited to 1 MiB, 32 agents, and 128 shared tool definitions. The deployer validates every manifest, tool reference, routing alias and schedule, then resolves webhook variables before creating agents. Each agent gets only the tools listed in its manifest. The response returns agent IDs, per-agent webhook signing secrets (show once and store securely), and a `routing_plan` with the original routing manifest and resolved alias IDs.

**Routing is not yet active.** The returned routing plan is not stored or attached to DIDs, and schedule-based agent switching is not implemented by telephony routing. Apply an appropriate default agent to the existing phone-number routing separately. Multi-agent installation is not an atomic database transaction: if a later agent fails, earlier agents are disabled, but these records still consume the organization agent quota until cleanup support exists. Operators should review deployments before retrying partial failures.

## Release policy

Use semantic version tags for released dependency boundaries.

- `leamout/contracts` tags changes to portable contracts.
- `leamout/ai-providers` tags official adapter releases and pins a released contracts version.
- `leamout/leamout` pins released contracts and provider versions.
- `leamout/agents` validates against released contracts and providers.

Pseudo-versions are allowed on coordinated development branches and stacked pull requests only. They should not remain in a normal release branch once the dependency PRs have merged.

Repositories do not need matching version numbers.

## Release order

When a change crosses repository boundaries, release in dependency order:

1. merge and tag `leamout/contracts`;
2. update, merge, and tag `leamout/ai-providers`;
3. update and merge `leamout/leamout`;
4. update and merge `leamout/agents`.

Breaking contract changes require a new major version once the ecosystem reaches v1.

## Compatibility gate

`.github/workflows/ecosystem.yml` checks the latest default branches together every day and on relevant runtime pull requests. It replaces module dependencies with the checked-out repositories for the test run, so it can detect future-main incompatibilities before the next release.

A green repository-local CI run is not sufficient for an ecosystem boundary change; the cross-repository compatibility gate must also pass before tagging.
