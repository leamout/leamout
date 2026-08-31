# Carrier test calls

Carrier test calls are an operator diagnostic and are deliberately separate
from SIP OPTIONS configuration validation.

Configure an exact comma-separated E.164 allowlist before using the workflow:

```env
CARRIER_TEST_CALL_DESTINATIONS=+14155550100,+442079460000
```

Create a test call with an organization-authenticated request:

```http
POST /v1/carrier-connections/{carrier_connection_id}/test-calls
Content-Type: application/json

{
  "trunk_id": "00000000-0000-0000-0000-000000000000",
  "from": "+14155550101",
  "to": "+14155550100"
}
```

The caller identity must already belong to the selected carrier connection,
and the trunk must resolve to that same connection. The workflow permits three
attempts per carrier connection per minute, allows at most 12 seconds for
origination, and hangs up an answered test call after five seconds.

Every admitted attempt is persisted with its actor, carrier connection, trunk,
selected endpoint, timestamps, final status, bounded error detail, and SIP call
identifier. List historical results with:

```http
GET /v1/carrier-connections/{carrier_connection_id}/test-calls?limit=50&offset=0
```

Configuration validation remains available independently at:

```http
POST /v1/carrier-connections/{carrier_connection_id}/validate
```
