# Review Pull Request

Review the current pull request against Leamout's engineering contract in `AGENTS.md` and the relevant path-specific instructions under `.github/instructions/`.

Focus on correctness before style.

Check for:

- domain-boundary violations;
- tenant-isolation mistakes;
- idempotency/retry bugs;
- payment/wallet invariants;
- BYOC versus managed-routing regressions;
- provider-specific behavior leaking into core domains;
- raw SQL or generated-file drift;
- unsafe secret/provider metadata exposure;
- acceptance-test weakening;
- stale or contradictory documentation;
- unnecessary abstractions or compatibility aliases;
- missing failure-path tests.

For each issue, explain the concrete failure mode and point to the smallest reasonable fix.

Do not invent problems merely to fill a checklist. If the change is correct, say which invariants it preserves.

Finish with:

1. blocking issues;
2. non-blocking improvements;
3. test/CI coverage that should run;
4. whether the PR is ready to merge.
