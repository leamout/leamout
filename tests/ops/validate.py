#!/usr/bin/env python3
"""Static production-deployment invariants (no live infrastructure required)."""

from pathlib import Path
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
    assert app["BACKOFFICE_URL"] == f"https://backoffice.{domain}"
    assert app["KINDE_ISSUER"].startswith("https://")
    assert app["KINDE_ISSUER"].endswith(".kinde.com")
    assert bool(app["KINDE_CLIENT_ID"])
    assert bool(app["KINDE_CLIENT_SECRET"])
    assert app["FRONTEND_URL"] in app["CORS_ORIGINS"].split(",")
    assert app["FRONTEND_URL"] != app["BACKEND_URL"]
    assert app["TRUSTED_PROXY_CIDRS"] == "172.33.0.10/32"

    compose = (ROOT / "deploy/compose/docker.yaml").read_text(encoding="utf-8")
    assert "KINDE_ISSUER:" in compose
    assert "keycloak:" not in compose
    assert "keycloak-postgres:" not in compose
    assert 'http://127.0.0.1:8081/readyz' in compose
    assert "172.32.0.0/24" in compose
    assert "172.33.0.0/24" in compose
    assert "\n  identity:\n" not in compose
    assert "env_file:" not in compose, "services must use explicit environment allowlists"

    caddy = (ROOT / "deploy/Caddyfile").read_text(encoding="utf-8")
    assert re.search(r"@privateMetrics\s+path /metrics", caddy)
    assert "respond @privateMetrics 404" in caddy
    assert "reverse_proxy server:8080" in caddy
    assert re.search(
        r"backoffice\.\{\$DOMAIN\}\s*\{\s*reverse_proxy backoffice:8082", caddy
    )

    target = (
        ROOT / "deploy/observability/prometheus/targets/leamout.yml"
    ).read_text(encoding="utf-8")
    assert "server:8080" in target
    assert "media:8090" in target
    print("PASS deployment origins, Kinde, worker health and metrics boundary")


if __name__ == "__main__":
    main()
