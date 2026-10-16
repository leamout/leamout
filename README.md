# Leamout

**Leamout is the carrier-grade runtime and control plane for autonomous voice agents.**

Leamout connects customer-owned SIP infrastructure to realtime AI agents. It
provides programmable call control, realtime media, AI orchestration, secure
tool execution, and observability. Customers bring their own carriers (BYOC)
and AI providers (BYOA).

## Architecture

```text
Customer SIP carrier / PBX / SBC
             |
          OpenSIPS
             |
        FreeSWITCH + RTPengine
             |
       Leamout Media Runtime
             |
      Voice Agent Runtime
             |
     STT / LLM / TTS
       or realtime AI

Control plane: Go API + worker, PostgreSQL, Redis, NATS JetStream
```

The three Go processes are `server` (API and control plane),
`worker` (background jobs), and `media` (realtime audio and agents).
OpenSIPS, FreeSWITCH, RTPengine, and Coturn handle SIP and media connectivity.

## Capabilities

- SIP-native calling with customer-owned trunks and numbers.
- Voice Agents with composable STT/LLM/TTS or integrated realtime models.
- Streaming audio, interruptions, tool execution, and call control.
- Organization-scoped credentials, routing, webhooks, and recordings.
- Call diagnostics, usage limits, and observability.

## Development

From the repository root:

```sh
cd server
go test ./...
go build ./...
```

Infrastructure is deployed and validated with the configurations in
[`deploy/`](deploy/README.md). Local acceptance tests live under `tests/`.

## Documentation

- [Architecture](docs/architecture.md)
- [Realtime media plane](docs/media-plane.md)
- [Getting started](docs/quickstart.md)
- [Deployment and observability](deploy/README.md)
- [Ecosystem](docs/ecosystem.md)

## Proprietary software

See [PROPRIETARY.md](PROPRIETARY.md) and [NOTICE](NOTICE) for software and
trademark notices.

User authentication is provided by Keycloak. See [Keycloak setup](docs/keycloak.md) for the cloud identity deployment and console configuration.
