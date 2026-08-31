# Secure WebRTC acceptance gate

This Playwright gate registers a real Chromium SIP user over WSS, requests
short-lived ICE credentials from Leamout, calls the FreeSWITCH echo service,
and fails unless Chromium selects a TURN relay candidate and receives audio.

The environment must point at a running, publicly reachable Leamout stack and
an API-provisioned SIP subscriber:

```sh
npm install
npx playwright install chromium
LEAMOUT_API_URL=https://api.example.test \
LEAMOUT_API_TOKEN=... \
LEAMOUT_ORGANIZATION_ID=... \
LEAMOUT_WSS_URL=wss://sip.example.test:5062 \
LEAMOUT_SIP_URI=sip:browser@example.test \
LEAMOUT_SIP_USERNAME=browser \
LEAMOUT_SIP_PASSWORD=... \
LEAMOUT_DESTINATION_URI=sip:9196@example.test \
npm test
```

Use a trusted certificate for both WSS and TURN-over-TLS. The test sets
`iceTransportPolicy` to `relay`; it does not silently fall back to host or
server-reflexive candidates.
