# Opt-in live Voice Agent validation

The live runner is **manual**, uses **real customer-owned SIP connectivity**
and **real organization-owned AI provider credentials**, and can incur carrier
and provider charges. It does not create agents, SIP trunks, numbers, routes
or credentials.

The existing simulated release gate remains unchanged:

    sh tests/voice-agent-v1/run.sh

## Preconditions

An operator must provision a real inbound number and outbound caller ID,
consented test destination, active Voice Agent with verified BYOA provider
binding, and organization-scoped API access. Python 3.11+ and the PostgreSQL
psql client are required. PostgreSQL access should be read-only.

Before starting, configure these environment variable names using your own
private values:

- LEAMOUT_LIVE_API_BASE: public HTTPS API origin
- LEAMOUT_LIVE_TOKEN: organization-scoped bearer token
- LEAMOUT_LIVE_ORGANIZATION_ID: organization UUID
- LEAMOUT_LIVE_AGENT_ID: active Voice Agent UUID
- LEAMOUT_LIVE_TRUNK_ID: real BYOC SIP trunk UUID
- LEAMOUT_LIVE_INBOUND_TO: inbound test number
- LEAMOUT_LIVE_OUTBOUND_FROM: carrier-authorized outbound caller ID
- LEAMOUT_LIVE_OUTBOUND_TO: consented test destination
- PGHOST, PGPORT, PGDATABASE, PGUSER, PGPASSWORD: private read-only PostgreSQL connection
- PGSSLMODE: verify-full recommended for remote PostgreSQL
- LEAMOUT_LIVE_CONFIRM: set to YES only when ready to place billable calls

Run from the repository root in an **interactive terminal**:

    python3 tests/voice-agent-live/run.py --direction both

Each call must be answered, complete successfully, and produce exactly one
completed Voice Agent session, nonzero durable media interval, and ordered
user/assistant turns. The operator must confirm audible two-way speech
by typing AUDIO-OK. Do not confirm if either audio direction failed.

Optional scenarios, which must be configured and exercised beforehand:

    python3 tests/voice-agent-live/run.py --direction both --require-tool --require-barge-in --require-recording

The runner does not simulate these scenarios. A failure of required evidence
fails the run. Use --direction inbound or --direction outbound to narrow scope.
Use --timeout with a value from 30 through 900 seconds.

On failure, the runner tries to hang up any outbound call it created. If
cleanup cannot be confirmed, inspect and terminate the call manually.

## CI policy

Only the following offline tests should run in CI:

    python3 -m unittest discover -s tests/voice-agent-live -p 'test_*.py' -v

Never schedule the live runner in normal CI or attach real provider credentials
to pull-request workflows. Keep phone numbers, credentials, transcripts,
recordings, and live customer identifiers out of public logs and PR comments.

For a private manual test report, record commit SHA, date, direction, redacted
call IDs, optional assertions, PASS/FAIL and any recovery steps. A successful
manual run does not establish reliability under worker restarts, media-node
failure, or high-concurrency soak conditions.
