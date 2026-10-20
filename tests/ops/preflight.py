#!/usr/bin/env python3
"""Fail closed on unsafe operator settings before a managed deployment.

This validates local environment files without loading secrets into a shell,
contacting production services, or printing credential values.
"""

import argparse
import base64
import binascii
import ipaddress
from pathlib import Path
import re
import stat
from urllib.parse import unquote, urlsplit


REQUIRED_SECRETS = {
    "POSTGRES_PASSWORD": 16,
    "KEYCLOAK_DB_PASSWORD": 16,
    "KEYCLOAK_ADMIN_PASSWORD": 16,
    "FREESWITCH_ESL_PASSWORD": 16,
    "MEDIA_CONTROL_TOKEN": 32,
    "MEDIA_TOKEN_SECRET": 32,
    "TURN_AUTH_SECRET": 32,
    "KEYCLOAK_CLIENT_SECRET": 32,
}


def read_env(path: Path) -> dict[str, str]:
    if not path.is_file():
        raise ValueError(f"{path}: environment file is missing")
    if stat.S_IMODE(path.stat().st_mode) & 0o077:
        raise ValueError(f"{path}: environment file must be private (chmod 600)")

    result = {}
    for number, raw_line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            raise ValueError(f"{path}:{number}: expected KEY=value")
        key, value = line.split("=", 1)
        if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", key):
            raise ValueError(f"{path}:{number}: invalid environment key")
        if key in result:
            raise ValueError(f"{path}:{number}: duplicate environment key {key}")
        value = value.strip()
        if value.startswith(("'", '"')):
            if len(value) < 2 or value[-1] != value[0]:
                raise ValueError(f"{path}:{number}: malformed quoted value for {key}")
            value = value[1:-1]
        result[key] = value
    return result


def is_weak_secret(value: str, minimum_length: int) -> bool:
    lowered = value.lower()
    return (
        len(value) < minimum_length
        or lowered.startswith(("change-me", "replace-", "placeholder", "example"))
        or len(set(value)) < 6
    )


def validate(infra: dict[str, str], app: dict[str, str]) -> list[str]:
    errors = []

    def check(condition: bool, message: str) -> None:
        if not condition:
            errors.append(message)

    domain = infra.get("DOMAIN", "").strip().lower()
    check(bool(re.fullmatch(r"(?:[a-z0-9-]+\.)+[a-z]{2,63}", domain)),
          "DOMAIN must be a public DNS name")
    check(domain == app.get("DOMAIN", "").strip().lower(),
          "DOMAIN differs between infrastructure and application")
    check(app.get("APP_ENV") == "production", "APP_ENV must be production")

    try:
        public_ip = ipaddress.IPv4Address(infra.get("PUBLIC_IP", ""))
        check(public_ip.is_global, "PUBLIC_IP must be a globally routable IPv4 address")
    except ipaddress.AddressValueError:
        errors.append("PUBLIC_IP must be a globally routable IPv4 address")

    expected = {
        "FRONTEND_URL": f"https://console.{domain}",
        "BACKEND_URL": f"https://api.{domain}",
        "BACKOFFICE_URL": f"https://backoffice.{domain}",
        "KEYCLOAK_ISSUER": f"https://auth.{domain}/realms/leamout",
    }
    for key, value in expected.items():
        check(app.get(key) == value, f"{key} must match the deployed {domain} hostname")

    origins = [
        origin.strip()
        for origin in app.get("CORS_ORIGINS", "").split(",")
        if origin.strip()
    ]
    check(origins == [expected["FRONTEND_URL"]],
          "CORS_ORIGINS must contain only the deployed console HTTPS origin")
    check(app.get("TRUSTED_PROXY_CIDRS") == "172.33.0.10/32",
          "TRUSTED_PROXY_CIDRS must trust only the Compose Caddy proxy")

    turn_host = f"turn.{domain}"
    check(infra.get("TURN_REALM") == turn_host,
          "TURN_REALM must match the deployed TURN hostname")
    advertised = app.get("TURN_PUBLIC_URLS", "").split(",")
    check(any(url.startswith(f"turn:{turn_host}:") for url in advertised),
          "TURN_PUBLIC_URLS must advertise the deployed TURN hostname")

    for key, minimum in REQUIRED_SECRETS.items():
        source = (
            infra
            if key in {"POSTGRES_PASSWORD", "KEYCLOAK_DB_PASSWORD", "KEYCLOAK_ADMIN_PASSWORD"}
            else app
        )
        check(
            not is_weak_secret(source.get(key, ""), minimum),
            f"{key} must be a strong, non-example secret ({minimum}+ characters)",
        )

    encryption_key = app.get("ENCRYPTION_KEY", "")
    try:
        raw_key = base64.b64decode(encryption_key, altchars=b"-_", validate=True)
        check(len(raw_key) in (16, 24, 32) and any(raw_key),
              "ENCRYPTION_KEY must decode to a nonzero 16, 24, or 32 byte key")
    except (ValueError, binascii.Error):
        errors.append("ENCRYPTION_KEY must be valid base64 for a 16, 24, or 32 byte key")

    try:
        database_url = urlsplit(app.get("DATABASE_URL", ""))
        check(
            database_url.scheme in ("postgres", "postgresql")
            and database_url.hostname == "postgres"
            and database_url.path == "/leamout",
            "DATABASE_URL must target the Compose PostgreSQL database",
        )
        check(
            unquote(database_url.password or "") == infra.get("POSTGRES_PASSWORD"),
            "DATABASE_URL password must match POSTGRES_PASSWORD",
        )
    except ValueError:
        errors.append("DATABASE_URL must be a valid PostgreSQL URL")

    polar_token = app.get("POLAR_ACCESS_TOKEN", "")
    polar_webhook = app.get("POLAR_WEBHOOK_SECRET", "")
    check(bool(polar_token) == bool(polar_webhook),
          "POLAR_ACCESS_TOKEN and POLAR_WEBHOOK_SECRET must be configured together")
    if polar_token or polar_webhook:
        check(not is_weak_secret(polar_token, 24),
              "POLAR_ACCESS_TOKEN must be a strong, non-example credential")
        check(not is_weak_secret(polar_webhook, 24),
              "POLAR_WEBHOOK_SECRET must be a strong, non-example credential")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--infra-env", type=Path, default=Path(".env"))
    parser.add_argument("--app-env", type=Path, default=Path("server/.env"))
    args = parser.parse_args()

    try:
        errors = validate(read_env(args.infra_env), read_env(args.app_env))
    except (OSError, UnicodeError, ValueError) as exc:
        errors = [str(exc)]

    if errors:
        print("FAIL: deployment preflight")
        for error in errors:
            print(f"  - {error}")
        return 1

    print("PASS: deployment environment preflight (secrets not displayed)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
