#!/usr/bin/env python3
"""Check the resolved Compose configuration for secret and network isolation.

Pipe `docker compose ... config --format json` into this script. Explicit
allowlists ensure adding a secret or a bridge requires deliberate review.
"""

import json
import sys

EXPECTED_ENVIRONMENTS = {
    "keycloak-postgres": {
        "POSTGRES_DB",
        "POSTGRES_USER",
        "POSTGRES_PASSWORD",
    },
    "keycloak": {
        "FRONTEND_URL",
        "BACKEND_URL",
        "BACKOFFICE_URL",
        "KEYCLOAK_CLIENT_SECRET",
        "KC_DB",
        "KC_DB_URL",
        "KC_DB_USERNAME",
        "KC_DB_PASSWORD",
        "KC_HOSTNAME",
        "KC_HTTP_ENABLED",
        "KC_PROXY_HEADERS",
        "KC_PROXY_TRUSTED_ADDRESSES",
        "KC_BOOTSTRAP_ADMIN_USERNAME",
        "KC_BOOTSTRAP_ADMIN_PASSWORD",
    },
    "postgres": {
        "POSTGRES_DB",
        "POSTGRES_USER",
        "POSTGRES_PASSWORD",
    },
    "migrate": {"DATABASE_URL"},
    "redis": set(),
    "nats": set(),
    "server": {
        "APP_ENV",
        "FRONTEND_URL",
        "BACKEND_URL",
        "DOMAIN",
        "DATABASE_URL",
        "REDIS_URL",
        "FREESWITCH_ESL_ADDRESS",
        "FREESWITCH_ESL_PASSWORD",
        "ENCRYPTION_KEY",
        "TURN_AUTH_SECRET",
        "TURN_PUBLIC_URLS",
        "CORS_ORIGINS",
        "TRUSTED_PROXY_CIDRS",
        "KEYCLOAK_ISSUER",
        "KEYCLOAK_CLIENT_ID",
        "KEYCLOAK_CLIENT_SECRET",
        "POLAR_ACCESS_TOKEN",
        "POLAR_WEBHOOK_SECRET",
        "POLAR_ENVIRONMENT",
        "POLAR_DEVELOPER_PRODUCT_ID",
        "POLAR_PRO_PRODUCT_ID",
        "POLAR_SCALE_PRODUCT_ID",
    },
    "backoffice": {
        "APP_ENV",
        "FRONTEND_URL",
        "BACKOFFICE_URL",
        "DATABASE_URL",
        "ENCRYPTION_KEY",
        "KEYCLOAK_ISSUER",
        "KEYCLOAK_CLIENT_ID",
        "KEYCLOAK_CLIENT_SECRET",
    },
    "worker": {
        "APP_ENV",
        "FRONTEND_URL",
        "DATABASE_URL",
        "REDIS_URL",
        "NATS_URL",
        "FREESWITCH_ESL_ADDRESS",
        "FREESWITCH_ESL_PASSWORD",
        "MEDIA_CONTROL_URL",
        "MEDIA_CONTROL_TOKEN",
        "ENCRYPTION_KEY",
        "AWS_ACCESS_KEY_ID",
        "AWS_SECRET_ACCESS_KEY",
        "AWS_FROM_EMAIL",
        "AWS_REGION",
        "AWS_CONFIGURATION_SET",
        "POLAR_ACCESS_TOKEN",
        "POLAR_WEBHOOK_SECRET",
        "POLAR_ENVIRONMENT",
        "POLAR_DEVELOPER_PRODUCT_ID",
        "POLAR_PRO_PRODUCT_ID",
        "POLAR_SCALE_PRODUCT_ID",
    },
    "media": {
        "REDIS_URL",
        "MEDIA_NODE_ID",
        "MEDIA_CONTROL_URL",
        "MEDIA_LISTEN_ADDRESS",
        "MEDIA_PUBLIC_WS_URL",
        "MEDIA_CONTROL_TOKEN",
        "MEDIA_TOKEN_SECRET",
        "MEDIA_TOKEN_TTL",
        "MEDIA_ATTACH_TIMEOUT",
        "MEDIA_MAX_SESSIONS",
        "MEDIA_MAX_FRAME_BYTES",
        "MEDIA_HANDSHAKE_TIMEOUT",
        "MEDIA_DRAIN_TIMEOUT",
    },
    "caddy": {
        "DOMAIN",
    },
    "freeswitch": {
        "FREESWITCH_ESL_PASSWORD",
    },
    "rtpengine": {
        "RTPENGINE_PUBLIC_IP",
    },
    "opensips": {
        "OPENSIPS_DATABASE_URL",
    },
    "coturn": {
        "TURN_AUTH_SECRET",
        "TURN_REALM",
        "TURN_EXTERNAL_IP",
        "TURN_MIN_PORT",
        "TURN_MAX_PORT",
    },
}

EXPECTED_NETWORKS = {
    "keycloak-postgres": {
        "identity",
    },
    "keycloak": {
        "edge",
        "identity",
    },
    "postgres": {
        "database",
    },
    "migrate": {
        "database",
    },
    "redis": {
        "cache",
    },
    "nats": {
        "events",
    },
    "server": {
        "edge",
        "control",
        "database",
        "cache",
    },
    "backoffice": {
        "edge",
        "database",
    },
    "worker": {
        "control",
        "database",
        "cache",
        "events",
    },
    "media": {
        "control",
        "cache",
    },
    "caddy": {
        "edge",
    },
    "freeswitch": {
        "control",
    },
    "rtpengine": {
        "control",
        "media",
    },
    "opensips": {
        "control",
        "database",
        "signaling",
    },
    "coturn": {
        "media",
    },
}

INTERNAL_NETWORKS = {"database", "cache", "events", "identity"}


def main():
    config = json.load(sys.stdin)
    services = config["services"]
    assert set(services) == set(EXPECTED_ENVIRONMENTS), "service allowlist changed"
    for name, expected in EXPECTED_ENVIRONMENTS.items():
        service = services[name]
        assert "env_file" not in service, f"{name}: full environment file is forbidden"
        actual = set(service.get("environment", {}))
        assert actual == expected, (
            f"{name}: missing {sorted(expected - actual)}; "
            f"unexpected environment keys {sorted(actual - expected)}"
        )

    assert set(EXPECTED_NETWORKS) == set(services), "network allowlist incomplete"

    for name, expected in EXPECTED_NETWORKS.items():
        actual = set(services[name].get("networks", {}))
        assert actual == expected, (
            f"{name}: expected networks {sorted(expected)}, got {sorted(actual)}"
        )

    for name in INTERNAL_NETWORKS:
        assert config["networks"][name]["internal"] is True, (
            f"{name}: network must be internal"
        )

    assert (
        services["caddy"]["networks"]["edge"]["ipv4_address"]
        == "172.33.0.10"
    ), "Caddy must retain its explicit trusted proxy IP"
    assert (
        services["keycloak"]["environment"]["KC_PROXY_TRUSTED_ADDRESSES"]
        == "172.33.0.10/32"
    ), "Keycloak must trust only Caddy, not the entire edge network"

    for name, network, address in (
        ("freeswitch", "control", "172.32.0.20"),
        ("rtpengine", "control", "172.32.0.10"),
        ("rtpengine", "media", "172.31.0.10"),
    ):
        assert (
            services[name]["networks"][network]["ipv4_address"] == address
        ), f"{name}: {network} fixed address changed"

    print("PASS resolved deployment service secrets and network isolation")


if __name__ == "__main__":
    main()
