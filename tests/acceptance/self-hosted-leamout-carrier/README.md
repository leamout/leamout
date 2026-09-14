# Self-Hosted + Leamout Carrier acceptance

This suite proves that a self-hosted Leamout runtime consumes Leamout Carrier through the same ordinary BYOC/SIP carrier model used for any customer-selected carrier:

```text
Leamout Carrier
      ↓ SIP
self-hosted OpenSIPS
      ↓
FreeSWITCH
```

The fixture creates an organization-scoped carrier connection using the built-in `leamout` carrier provider, assigns the test number to that connection, and routes an inbound SIP INVITE through normal carrier ingress. The connection and number remain BYOC resources. There is no self-hosted managed mode, hosted-edge forwarding hop, deployment lookup, or runtime attachment.

The suite also disables the carrier connection and proves that generic carrier state alone controls whether the inbound call may reach the self-hosted runtime.

Run it with:

```sh
sh tests/acceptance/self-hosted-leamout-carrier/run.sh
```
