# Managed SIP edge acceptance

This suite exercises the hosted OpenSIPS boundary with real SIP Digest
challenge/response traffic and a synthetic wholesale SIP endpoint.

```sh
sh tests/acceptance/managed-sip-edge/run.sh
```

It covers unauthenticated and incorrect credentials, caller-ID authorization,
managed entitlement enforcement, inactive trunk/organization state, successful
platform routing, and removal of customer `Proxy-Authorization` before the
wholesale leg. CPS, concurrency, and credit exhaustion remain follow-up cases.
