# Call orchestration roadmap

This roadmap tracks the path from authenticated SIP calls to durable,
API-controlled programmable voice. OpenSIPS remains the signaling authority,
RTPengine remains the media boundary, and FreeSWITCH acts as an internal media
worker controlled through ESL.

## Phase A — authoritative call lifecycle

- [x] Park trusted inbound voice-application channels without answering them.
- [x] Subscribe the worker to FreeSWITCH channel lifecycle events.
- [x] Create inbound calls from trusted OpenSIPS metadata.
- [x] Normalize answer, hold, resume, and hangup events into durable call state.
- [x] Publish call domain events through the transactional outbox.
- [x] Reconcile database calls against active FreeSWITCH channels after restarts.
- [x] Add an end-to-end acceptance gate for an inbound parked call controlled
      through the public answer, media, and hangup APIs.

## Runtime contract

1. OpenSIPS authenticates ingress, resolves the DID and voice application, and
   removes any externally supplied Leamout metadata.
2. OpenSIPS adds trusted organization and voice-application identifiers before
   relaying the call to FreeSWITCH through RTPengine.
3. FreeSWITCH sends provisional ringing and parks the channel without answering.
4. The worker consumes ESL lifecycle events and persists the inbound call.
5. API commands address the durable Leamout call; the media controller resolves
   its FreeSWITCH channel and executes the requested operation.
6. Hangup events or reconciliation close the durable call and publish its final
   lifecycle event.

FreeSWITCH dialplan actions must not select carriers, infer tenants, or execute a
default application flow. Those decisions belong to the Leamout control plane.

## Acceptance coverage

`tests/acceptance/voice-v1` now proves the authoritative call lifecycle against
the real Docker Compose telecom stack. The gate verifies that an authenticated
carrier INVITE can remain parked while ringing, become a durable inbound call,
be answered and controlled through the public API, emit normalized lifecycle
events, and leave no active application channel after hangup. The same suite
also exercises outbound calling, media controls, recordings, conferences,
webhooks, health checks, and restart recovery.
