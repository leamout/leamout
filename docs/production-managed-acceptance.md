# Production managed carrier acceptance

Production DIDWW/CommPeak acceptance follows the synthetic Cloud + Managed
coexistence gate. It is deliberately not part of pull-request CI because it
orders billable resources, places real PSTN calls, and requires protected
provider credentials.

## Entry criteria

- the Cloud + Managed synthetic acceptance workflow is green;
- a dedicated production test organization and spend-limited plan exist;
- DIDWW and CommPeak credentials are stored in the protected CI environment;
- the managed ingress and termination resources are operator-owned and tagged
  for acceptance cleanup;
- provider-operation and CDR polling have durable cursors and retry diagnostics.

## Required proof

1. Search and purchase a real voice-capable DID through the customer API.
2. Confirm DIDWW routes the DID to the production Leamout managed ingress.
3. Receive a real PSTN call in the Leamout Cloud voice application runtime.
4. Use that DID as caller ID on a trunkless call terminated by CommPeak.
5. Poll the CommPeak CDR, correlate the provider SIP identity to the Leamout
   call, and create exactly one wholesale charge.
6. Replay provider polling and prove neither CDR nor charge duplication.
7. Disable the number, organization, and managed route in turn and prove each
   path fails closed.
8. Release the DID, reconcile provider state, and verify no routable resource
   or billable test artifact remains.

## CI boundary

The future workflow must use `workflow_dispatch`, a protected GitHub
environment, concurrency serialization, a strict spend cap, and an unconditional
cleanup job. It must never run for forks or ordinary pull requests. Until all
entry criteria exist, the production workflow should not be added as a runnable
stub that could imply unsupported readiness.
