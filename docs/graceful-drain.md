# Graceful telecom node drain

This document defines the production drain contract for the OpenSIPS, RTPengine, and FreeSWITCH nodes in the self-hosted deployment.

The objective is to remove a node from service without dropping established calls. Draining is an explicit operational state, separate from process shutdown.

## Safety invariants

A drain must preserve these invariants:

1. Existing SIP dialogs continue to accept in-dialog requests such as ACK, BYE, CANCEL, and re-INVITE.
2. Existing RTPengine sessions remain available until their calls end.
3. FreeSWITCH does not accept new sessions after the drain begins, but existing sessions may complete normally.
4. New out-of-dialog call attempts fail fast and predictably while OpenSIPS is draining.
5. A node is stopped only after active signaling and application sessions reach zero, unless an operator explicitly forces termination.
6. Drain state is reversible before shutdown.
7. Normal container restart policy must not accidentally turn a drain command into an immediate restart loop.

## Drain sequence

The deployment drain command will coordinate the services in this order.

### 1. Drain OpenSIPS admission

Set an OpenSIPS runtime drain flag through its local management interface.

While the flag is active:

- existing in-dialog requests continue through the normal route;
- health/diagnostic OPTIONS requests continue to work;
- new REGISTER, INVITE, and MESSAGE requests are rejected before application routing;
- the rejection is a temporary service response so callers may retry another node.

OpenSIPS remains running during the entire call-drain window because it owns signaling for established dialogs.

### 2. Pause FreeSWITCH admission

Use the FreeSWITCH control interface to stop accepting new sessions without terminating active calls.

Do not use `fsctl shutdown elegant` as the first drain action under the current Compose `restart: unless-stopped` policy. An elegant process exit can be interpreted by the container runtime as a restart. The coordinated drain keeps the process alive until the host-side controller intentionally stops the service.

### 3. Wait for active calls

The drain controller waits until both signaling and application session counts are zero.

The authoritative checks are:

- OpenSIPS active dialog count;
- FreeSWITCH active session/channel count.

RTPengine stays online throughout this period. Because OpenSIPS no longer admits new calls, its active media sessions can only decrease as existing calls terminate.

The wait must have an operator-configurable deadline. Reaching the deadline is a failure, not permission to kill active calls automatically.

### 4. Stop the telecom services

After the active counts reach zero, stop services intentionally through Compose in dependency-safe order:

1. OpenSIPS;
2. FreeSWITCH;
3. RTPengine.

At this point there are no established dialogs or application sessions relying on RTPengine.

## Resume sequence

Before the services are stopped, the operator may cancel a drain:

1. resume FreeSWITCH admission;
2. clear the OpenSIPS runtime drain flag;
3. verify that new call admission succeeds.

A resume operation must be idempotent.

## Implementation slices

### Slice 1 — OpenSIPS drain state

- add a local-only OpenSIPS management transport;
- add a shared runtime `draining` flag initialized to false;
- reject new out-of-dialog application traffic while draining;
- preserve in-dialog routing before the drain check;
- expose commands to read, set, and clear the state;
- validate the configuration at image build time.

### Slice 2 — coordinated host command

Add a deployment command that:

- starts and cancels drain mode;
- pauses/resumes FreeSWITCH admission;
- polls OpenSIPS and FreeSWITCH active counts;
- exits non-zero when the drain deadline expires;
- intentionally stops OpenSIPS, FreeSWITCH, and RTPengine only at zero active calls;
- is safe to invoke repeatedly.

### Slice 3 — acceptance gate

Add automated coverage proving that:

1. an established call survives entry into drain mode;
2. a new call is rejected while draining;
3. BYE for the established call succeeds and releases media;
4. the drain command observes zero active calls;
5. services can then be stopped without truncating the established call;
6. cancelling drain restores new-call admission.

## Completion criteria

The BYOC production-hardening roadmap item is complete only after all three slices are implemented and the acceptance gate is green. Documentation alone does not satisfy the roadmap checkbox.
