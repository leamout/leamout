# Metering

Metering converts authoritative telecom domain outcomes into an immutable commercial usage ledger.

## Flow

```text
telecom domain
    │
    │ normalized domain event
    ▼
commercial/metering/consumer.go
    ↓
metering/service.go
    ↓
usage_events
```

Raw carrier or FreeSWITCH events should not be billed directly. Telecom modules first normalize lifecycle state and determine the authoritative billable outcome.

## Meters

A meter defines a billable unit.

Current meter fields are:

```text
key
name
unit
active
```

Example meter keys:

```text
voice.outbound.seconds
voice.inbound.seconds
sms.outbound.segments
sms.inbound.segments
recording.minutes
recording.storage.bytes
ai.tokens
```

Prefer integer quantities in the smallest useful unit:

```text
seconds
segments
bytes
tokens
numbers
```

The current model intentionally does not implement a generic aggregation DSL. A usage event carries an explicit quantity, and billing aggregates those quantities as needed.

## Usage event

A usage event contains:

```text
organization_id
subscription_id (optional)
meter_id
quantity
source_type
source_id
idempotency_key
dimensions
occurred_at
created_at
```

Example:

```json
{
  "meter": "voice.outbound.seconds",
  "quantity": 187,
  "source_type": "call",
  "source_id": "call_uuid",
  "dimensions": {
    "country": "GH",
    "network": "MTN",
    "direction": "outbound"
  }
}
```

`source_type` and `source_id` connect commercial usage to the domain record that produced it.

## Authoritative usage

Customers must not directly declare authoritative billable usage.

For voice, the intended path is:

```text
FreeSWITCH event stream
        ↓
normalized call lifecycle
        ↓
terminal call state + billable duration
        ↓
metering consumer
        ↓
usage_event
```

This prevents billing from depending on untrusted client counters or raw infrastructure event noise.

## Idempotency

At-least-once asynchronous delivery means the same telecom outcome can be processed more than once.

`idempotency_key` is globally unique in the database and insertion uses conflict protection so a retry does not create duplicate billable usage.

Normal lookup remains organization-scoped for tenant safety.

## Immutability

Usage events should be treated as an append-only accounting ledger. Corrections should be represented deliberately rather than silently mutating historical quantities.

The current schema accepts only positive quantities. If negative adjustments become necessary, introduce an explicit adjustment model rather than weakening the meaning of the existing ledger without a migration/design decision.

## Subscription association

`subscription_id` is optional. When present, it must belong to the same organization.

Metering does not require that a referenced subscription still be active at query time. Historical usage can outlive subsequent subscription-state changes.

## Dimensions

`dimensions` is a JSON object for rating context such as:

```text
country
network
direction
carrier/provider
```

Keep dimensions intentional and bounded. They are rating/audit context, not an unstructured dumping ground for entire telecom events.
