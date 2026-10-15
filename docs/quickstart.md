# Leamout quickstart

Leamout is a managed service. Customers do not deploy OpenSIPS,
FreeSWITCH, Redis or the Go processes. Leamout operates those services.

## 1. Sign in to the Leamout console

Use the hosted Leamout console at https://console.leamout.com to create an
account or sign in. Create or select your organization.

Choose the Leamout subscription plan that meets your intended call concurrency,
monthly runtime minutes, SIP trunk count, Voice Agent count and retention
requirements. Complete checkout if the selected plan requires payment.

## 2. Connect an existing carrier

Set up an organization-owned SIP trunk in the console using the address,
authentication credentials, trusted source networks and codec settings from
your existing SIP carrier, PBX or SBC. The carrier relationship and number
charges remain yours (BYOC). Do **not** install a local Leamout SIP server.

Attach or register your existing phone number to the trunk and configure
routing. For inbound calls, use the Cloud SIP ingress endpoint supplied by
Leamout for your account and trunk; do not invent a hostname or use a generic
shared endpoint without the assigned routing and security configuration.

## 3. Configure an AI provider

Add an organization-owned AI provider integration through the console or API.
Use credentials for your selected realtime model or composable STT/LLM/TTS
providers (BYOA). Verify the integration before activating an agent.

## 4. Create a Voice Agent

Create and configure a Voice Agent, select the engine and provider bindings,
provide instructions and any approved tools, and associate the appropriate
telephony routing configuration.

Use the agent's readiness diagnostics to verify its provider configuration,
connectivity and call-routing prerequisites before activation.

## 5. Make a test call

Originate or receive a call through your connected SIP trunk, then use the
Leamout console to review the call, event diagnostics, media session history,
conversation records and runtime usage. Active calls are not terminated
solely because the monthly allowance is subsequently reached.

## Operational note

The product workflow above is the intended Leamout onboarding flow. Certain
console screens, carrier ingress details and live deployment steps may still
require implementation or production validation. Do not confuse the
repository's internal Docker Compose and synthetic acceptance tests with
customer onboarding.

## Engineering acceptance (Leamout operators only)

From a Leamout development checkout, engineers can validate the runtime using
`sh tests/voice-agent-v1/run.sh`. This starts disposable local test infrastructure
and does not create a customer-hosted product or bypass subscription plans.

For implementation details, see [architecture.md](architecture.md),
[media-plane.md](media-plane.md), [monetization.md](monetization.md), and
[storage.md](storage.md).
