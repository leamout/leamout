#!/usr/bin/env python3
"""Static production-deployment invariants (no live infrastructure required)."""

from pathlib import Path
import json
import re

ROOT = Path(__file__).resolve().parents[2]


def dotenv(path):
    values = {}
    for line in (ROOT / path).read_text(encoding="utf-8").splitlines():
        if line and not line.startswith("#") and "=" in line:
            key, value = line.split("=", 1)
            values[key] = value.strip()
    return values


def main():
    infra = dotenv(".env.example")
    app = dotenv("server/.env.example")
    domain = app["DOMAIN"]

    assert infra["DOMAIN"] == domain, "root and app DOMAIN disagree"
    assert app["FRONTEND_URL"] == f"https://console.{domain}"
    assert app["BACKEND_URL"] == f"https://api.{domain}"
    assert app["KEYCLOAK_ISSUER"] == f"https://auth.{domain}/realms/leamout"
    assert app["FRONTEND_URL"] in app["CORS_ORIGINS"].split(",")
    assert app["FRONTEND_URL"] != app["BACKEND_URL"]

    realm = json.loads(
        (ROOT / "deploy/keycloak/leamout-realm.json").read_text(encoding="utf-8")
    )
    client, = [item for item in realm["clients"] if item["clientId"] == "leamout-console"]
    assert client["redirectUris"] == ["${FRONTEND_URL}/api/v1/auth/callback"]
    assert client["webOrigins"] == ["${FRONTEND_URL}"]
    assert client["attributes"]["backchannel.logout.url"] == (
        "${BACKEND_URL}/v1/identity/backchannel-logout"
    )
    assert client["attributes"]["pkce.code.challenge.method"] == "S256"

    compose = (ROOT / "deploy/compose/docker.yaml").read_text(encoding="utf-8")
    assert "KC_HOSTNAME: https://auth.${DOMAIN:-leamout.com}" in compose
    assert 'http://127.0.0.1:8081/readyz' in compose
    assert "172.32.0.0/24" in compose

    caddy = (ROOT / "deploy/Caddyfile").read_text(encoding="utf-8")
    assert re.search(r"@privateMetrics\s+path /metrics", caddy)
    assert "respond @privateMetrics 404" in caddy
    assert "reverse_proxy server:8080" in caddy

    target = (
        ROOT / "deploy/observability/prometheus/targets/leamout.yml"
    ).read_text(encoding="utf-8")
    assert "server:8080" in target
    assert "media:8090" in target
    print("PASS deployment origins, Keycloak, worker health and metrics boundary")


if __name__ == "__main__":
    main()
