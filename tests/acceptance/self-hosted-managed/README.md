# Self-Hosted + Managed Carrier acceptance

This suite proves that a self-hosted Leamout runtime consumes Leamout Managed Carrier through the same ordinary SIP carrier model used for BYOC carriers:

```text
Leamout Managed Carrier
        ↓ SIP
self-hosted OpenSIPS
        ↓
FreeSWITCH
```

The fixture creates an organization-scoped carrier connection using the built-in `leamout` carrier provider, assigns the test number to that connection, and routes an inbound SIP INVITE through normal carrier ingress. There is no hosted-edge forwarding hop, deployment lookup, or runtime attachment.

The suite also disables the carrier connection and proves that carrier state alone controls whether the inbound call may reach the self-hosted runtime.

Run it with:

```sh
sh tests/acceptance/self-hosted-managed/run.sh
```
