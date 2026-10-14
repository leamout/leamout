# Prometheus integration

The Community API and media plane already provide /metrics on ports 8080 and 8090.
This change adds worker metrics on port 9091 and a minimal FreeSWITCH ESL exporter on 9282.
It also enables the native OpenSIPS Prometheus HTTP module and restricts OpenSIPS and
RTPEngine scraping to exact-path read-only HTTP proxy endpoints.

## Optional Compose overlay

Start the leamout/enterprise observability Compose stack first, creating the
`leamout-observability` external Docker network. From the public Core checkout:

```sh
docker compose -f deploy/compose.yaml -f deploy/compose.observability.yaml up -d --build
```

FreeSWITCH ESL password must be supplied to Compose interpolation via
`FREESWITCH_ESL_PASSWORD`, matching the credentials in the Core environment.
The original Core Compose deployment continues working without the overlay.

## Prometheus targets

```json
[
  {"targets":["server:8080"],"labels":{"component":"server"}},
  {"targets":["worker:9091"],"labels":{"component":"worker"}},
  {"targets":["media:8090"],"labels":{"component":"media"}},
  {"targets":["opensips-metrics-proxy:8080"],"labels":{"component":"opensips"}},
  {"targets":["freeswitch-exporter:9282"],"labels":{"component":"freeswitch"}},
  {"targets":["rtpengine-metrics-proxy:8080"],"labels":{"component":"rtpengine"}}
]
```

Place the JSON in the Enterprise Prometheus `targets/leamout.json` file.
Never publish the metrics ports on the host; prevent public access to the
API /metrics endpoint through Caddy. The app and media HTTP listeners contain
other routes: ensure the observability network is trusted and isolated.

The worker exporter currently exposes process metrics, not job throughput.
Prometheus collection is independent of SIP/RTP and never blocks voice calls.
Verify OpenSIPS packaging and exporter behavior with a real Docker build and
end-to-end scrape test before production deployment. Existing per-trunk and
endpoint metric labels require cardinality review at scale.
