# Self-Hosted + Managed Carrier acceptance

This suite runs two real OpenSIPS instances to prove the managed inbound
attachment boundary:

```text
synthetic carrier -> hosted managed edge -> runtime attachment
                  -> self-hosted OpenSIPS -> FreeSWITCH
```

It verifies that the hosted edge derives the organization from the managed
DID, resolves the healthy and verified attachment, forwards the INVITE to the
self-hosted runtime, and that the self-hosted runtime independently resolves
its local voice binding. It then marks the attachment unhealthy and proves a
new INVITE fails closed at the hosted edge.

Run it with:

```sh
sh tests/acceptance/self-hosted-managed/run.sh
```
