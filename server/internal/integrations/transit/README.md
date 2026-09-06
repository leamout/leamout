# Leamout Transit

Leamout Transit is the Leamout-managed bridge used when a deployment consumes Managed Carrier without connecting directly to the upstream telecom provider.

```text
Self-Hosted Leamout
        ↓
Managed Carrier enabled
        ↓
Leamout Transit
        ↓
DIDWW / future managed carriers
```

This package contains the client-side integration for that bridge. It is not a carrier adapter and it is not part of BYOC.

Transit authenticates a Leamout deployment to Leamout-managed carrier services, searches managed inventory, and submits managed number operations using opaque Leamout handles. Upstream carrier credentials, inventory IDs, product/SKU IDs, routing resource IDs, and other provider-specific metadata stay behind Transit.

Direct upstream provider integrations remain under `internal/integrations/carriers/*` (for example `carriers/didww`). Transit is the boundary a self-hosted deployment uses instead of receiving Leamout's provider credentials.
