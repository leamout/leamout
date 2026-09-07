# Cloud + Managed Carrier acceptance

This suite establishes the Leamout-operated Cloud + Managed topology and proves
the customer-facing managed-number purchase workflow against a deterministic
fake DIDWW provider:

```text
customer API -> Leamout Cloud control plane -> provider operation worker
             -> fake managed-number provider -> cloud managed ingress
```

The current phase proves inventory search, opaque selection, durable purchase
intent, asynchronous order execution, provider routing convergence, activation,
and suppression of provider/platform identifiers from customer responses.
Inbound and outbound SIP calls through this topology are the next phase.

Run it with:

```sh
sh tests/acceptance/cloud-managed/run.sh
```
