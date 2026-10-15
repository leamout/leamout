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
