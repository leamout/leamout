# Leamout scrape targets

Prometheus discovers files matching `*.json` in this directory. No Leamout-specific metric endpoints are assumed by the bootstrap stack.

Copy `leamout.json.example` to `leamout.json` and **replace** the target with a verified, reachable `/metrics` endpoint exposed by the Community core or supported exporter. The default Prometheus metrics path is `/metrics`.

Targets must be reachable from the shared `leamout-observability` Docker network. Use a Core Compose override to attach the relevant service(s) to this network; do not publish metrics to the public internet.

Use bounded, low-cardinality target labels (such as environment and component). Do not use phone numbers, call IDs, or tenant IDs as labels. Prometheus discovers changes to matching files automatically.
