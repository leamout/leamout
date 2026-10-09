# Organization API list filters

The organization list endpoints below accept optional query parameters. Filters
combine with AND and always apply within the organization selected by the existing
authentication and organization middleware. UUIDs from another organization do not
expand that scope; they produce an empty result.

| Resource | Filters |
| --- | --- |
| Calls | `state`, `direction`, `trunk_id`, `voice_agent_id`, `created_from`, `created_before` |
| Recordings | `status`, `call_id`, `created_from`, `created_before` |
| Phone numbers | `status`, `country_code`, `trunk_id`, `voice_enabled` |
| Trunks | `status`, `direction`, `inbound_enabled` |
| Voice agents | `engine`, `language` |
| Audit events | `action`, `actor_type`, `actor_id`, `target_type`, `target_id`, `occurred_from`, `occurred_before` |

Text filters are exact matches after trimming surrounding whitespace. Country codes
are normalized to uppercase and must contain two ASCII letters, for example `gh`
or `GH`. Booleans must be `true` or `false`; explicit `false` filters are distinct
from omitted filters. UUID filters require a non-zero UUID. New filters supplied
as empty values, repeated parameters, malformed UUIDs, timestamps, or booleans
return HTTP 400. Unknown parameters retain the API's existing behavior and are
ignored.

Accepted enum values:

- Call state: `initiating`, `ringing`, `answered`, `active`, `completed`, `failed`, `cancelled`.
- Call direction: `inbound`, `outbound`.
- Recording status: `recording`, `uploading`, `completed`, `failed`.
- Phone number status: `active`, `disabled`, `porting`.
- Trunk status: `active`, `disabled`.
- Trunk direction: `inbound`, `outbound`, `bidirectional` (exact match).
- Agent engine: `composable`, `realtime`.
- Audit actor type: `user`, `organization_token`.

Date bounds require RFC3339 timestamps with a timezone, for example
`2026-01-01T00:00:00Z`. `*_from` is inclusive and `*_before` is exclusive. Either
bound may be omitted; when both are present, the start must precede the end.
Dates refer to `created_at` for calls and recordings and `occurred_at` for audit
events. URL-encode `+` in timezone offsets.

For example, a calls list query can combine
`?state=completed&direction=inbound&created_from=2026-01-01T00:00:00Z&created_before=2026-02-01T00:00:00Z&limit=50&offset=0`.
A numbers list can use `?country_code=GH&voice_enabled=false`.

Existing response envelopes, field names, and pagination contracts are preserved.
Calls and recordings retain the default limit of 50 and maximum of 200. Audit
events retain the default limit of 50 and maximum of 100. All three retain an
offset default of zero. Numbers, trunks, and agents retain their existing unpaginated
lists. Results order newest first with descending ID as a tie-breaker.

Deleted recordings, released numbers, and inactive agents remain excluded.
Agent lists also retain their active, non-deleted organization check. Filtering
does not introduce access to hidden records or change authorization requirements.

## Validation

Run `sqlc generate` and `sqlc vet` in `server/`, followed by `go test ./...`,
`go vet ./...`, and `go build ./...`.

The database integration test is opt-in. Set `TEST_DATABASE_URL` to a
disposable PostgreSQL database and run:

```sh
cd server
TEST_DATABASE_URL='postgres://localhost/leamout_test?sslmode=disable' \
  go test ./internal/database -run TestOrganizationListFilters -v
```

The test runs migrations and fixtures in a transaction with an isolated schema,
then rolls back. The database user needs permission to create schemas and the
extensions used by the migrations. It checks combined filters, omitted filters,
explicit false values, tenant isolation, hidden records, date boundaries, and
ordering when timestamps tie.
