# Usage

The usage module records immutable, organization-scoped observations of measured use. A usage event is operational commercial telemetry; recording one does not by itself make that usage billable or debit a customer wallet.

## Boundary

```text
telecom outcome
    ↓
commercial/usage
    ↓
usage_events
```

`commercial/usage` owns durable usage observations and idempotent ingestion. It does not own customer pricing, wallet debits, provider wholesale cost, invoicing, or a generalized rating engine.

## Event model

A usage event contains:

```text
organization_id
meter_id
quantity
source_type
source_id
idempotency_key
dimensions
occurred_at
created_at
```

There is no subscription association in the PAYG model.

`quantity` is a positive integer in the meter's defined unit. `source_type` and `source_id` connect the observation to the domain record that produced it.

## Meters

Meters are Catalog resources. A meter gives a stable identity to a measurable quantity such as:

```text
voice.outbound.seconds
sms.outbound.segments
recording.storage.bytes
```

A meter does not decide whether an event should be charged. It only defines what was measured.

## Authoritative source

Customers must not declare authoritative commercial usage directly. Telecom modules should first normalize provider or media events into a trusted domain outcome, then record the corresponding usage observation.

For voice, the intended direction is:

```text
FreeSWITCH / carrier events
        ↓
normalized call lifecycle
        ↓
terminal authoritative outcome
        ↓
usage event
```

## Idempotency

At-least-once delivery can produce retries. `idempotency_key` prevents the same authoritative outcome from being recorded twice.

Replaying the same key with the same event is safe. Reusing the same key for conflicting event data is rejected.

## Immutability

Usage events are append-only observations. Historical events should not be silently rewritten to change commercial meaning.

If a future workflow needs corrections, introduce an explicit correction model rather than turning usage events into mutable accounting rows.

## PAYG relationship

Managed-provider authorization happens before provider exposure through Wallet reservations. Usage events can support analytics, reconciliation, and future product behavior, but they are not the pay-before-use authorization boundary.

```text
wallet authorization
        ↓
managed operation
        ↓
authoritative outcome
        ↓
usage observation
```

Provider wholesale charges remain separate COGS records and must not be inferred from customer usage events.
