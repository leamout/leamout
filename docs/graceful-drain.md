# Graceful telecom node drain

This document defines the production drain contract for the OpenSIPS, RTPengine, and FreeSWITCH nodes in the self-hosted deployment.

The objective is to remove a node from service without dropping established calls. Draining is an explicit operational state, separate from process shutdown.

## Safety invariants

A drain preserves these invariants:

1. Existing SIP dialogs continue to accept in-dialog requests such as ACK, BYE, CANCEL, and re-INVITE.
2. Existing RTPengine sessions remain available until their calls end.
3. FreeSWITCH does not accept new sessions after the drain begins, but existing sessions may complete normally.
4. New out-of-dialog application traffic fails fast while OpenSIPS is draining.
5. A node is stopped only after active signaling and application sessions reach zero.
6. Reaching the drain deadline never authorizes automatic call termination.
7. Drain state is reversible before shutdown.
8. Normal container restart policy does not turn a drain into an immediate restart loop.

## Implemented controls

### OpenSIPS

`deploy/opensips/drain.cfg` reserves global flag position `0` for admission drain state and exposes it through a local-only MI FIFO under `/run/opensips`.

The image installs `/usr/local/bin/leamout-opensips-drain` with these commands:

- `enable` — set drain state;
- `resume` / `disable` — clear drain state;
- `status` — report `draining` or `accepting`;
- `dialogs` — return the `dialog:active_dialogs` statistic.

The main request route handles established dialogs first. Diagnostic OPTIONS remains available. The drain route then returns `503 Service Draining` for new out-of-dialog application traffic before REGISTER, INVITE, or MESSAGE processing.

### FreeSWITCH

The image installs `/usr/local/bin/leamout-freeswitch-drain`, which controls the local FreeSWITCH process through `fs_cli` and the existing ESL password.

Commands:

- `enable` — `fsctl pause`;
- `resume` / `disable` — `fsctl resume`;
- `status` — `fsctl pause_check`;
- `channels` — parse `show channels count`.

`fsctl pause` prevents new inbound and outbound sessions without hanging up established channels.

### RTPengine

RTPengine does not need a separate admission switch because OpenSIPS is the component that creates new RTPengine sessions. Once OpenSIPS admission is drained, RTPengine is kept online while active dialogs/channels finish and is intentionally stopped last.

## Operator commands

Run commands from the repository root.

### Drain and stop the telecom node

```sh
./deploy/drain.sh
```

The default deadline is 300 seconds and polling interval is 2 seconds. Override them when necessary:

```sh
LEAMOUT_DRAIN_TIMEOUT_SECONDS=900 \
LEAMOUT_DRAIN_POLL_SECONDS=5 \
./deploy/drain.sh
```

The command performs this sequence:

1. enable OpenSIPS admission drain;
2. pause FreeSWITCH admission;
3. poll OpenSIPS active dialogs and FreeSWITCH active channels;
4. if both reach zero, stop OpenSIPS;
5. stop FreeSWITCH;
6. stop RTPengine last.

If the deadline expires while calls remain, the command exits non-zero and deliberately leaves OpenSIPS and FreeSWITCH drained. It does not kill active calls.

### Cancel a drain before shutdown

```sh
./deploy/resume.sh
```

Resume occurs in dependency-safe order:

1. FreeSWITCH admission resumes;
2. OpenSIPS admission resumes.

The operation is idempotent while both services are running.

## Compose shutdown behavior

Compose gives the telecom services explicit shutdown grace periods as a final safety net:

- OpenSIPS: 30 seconds;
- FreeSWITCH: 60 seconds;
- RTPengine: 30 seconds.

The coordinated drain normally reaches zero active calls before any process receives its stop signal, so these periods protect unexpected or manual stop paths rather than replacing the drain protocol.

## Validation

The Docker workflow validates:

- POSIX shell syntax for every drain helper;
- OpenSIPS drain configuration insertion and route ordering;
- Compose configuration;
- OpenSIPS image configuration through the existing `opensips -C` build check;
- FreeSWITCH image build;
- a live FreeSWITCH drain helper smoke test covering accepting -> draining -> accepting state and active-channel count parsing;
- RTPengine image build.

A full call-level acceptance test that establishes a SIP call, enters drain mode, rejects a second call, then completes BYE/media release is still required before marking the BYOC production-hardening roadmap checkbox complete.

## Completion criteria

The runtime drain implementation is present when the controls above build and pass CI. The roadmap item is complete only when the full call-level acceptance gate is also green; documentation or component-level smoke tests alone do not satisfy that checkbox.
