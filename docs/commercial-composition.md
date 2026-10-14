# Core and Cloud composition

Core `/leamout/server`, `/leamout/worker`, and `/leamout/media` do not
require `DEPLOYMENT_MODE`, `DEPLOYMENT_ID`, or a commercial license.
They run with organization-scoped admission and SIP trunk CPS/concurrency
limits, but no commercial subscription quota enforcement.

Leamout-hosted Cloud uses explicit `/leamout/cloud-server` and
`/leamout/cloud-worker` binaries, selectable together through
`deploy/compose.cloud.yaml`. These install billing, subscription and
entitlement services plus mandatory plan limits and monthly runtime
allowance checks before serving requests or processing call events.

```sh
docker compose -f deploy/compose.yaml -f deploy/compose.cloud.yaml up -d
```

**Cloud deployment warning:** running the ordinary Core commands omits
subscription enforcement by design. Cloud operators must deploy both Cloud
commands together with a validated release configuration; never infer
Cloud identity from optional Polar environment values.

Cloud services remain in this public repository but do not form a dependency
in the Core Voice Agent, media, or telephony execution path.
