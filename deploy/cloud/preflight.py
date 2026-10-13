#!/usr/bin/env python3
"""Fail-closed, secret-redacting production configuration preflight.

Checks existing Compose/app settings; does not provision a server or modify files.
Use on the host before starting the private-beta stack.
"""
import argparse
import base64
import ipaddress
import os
from pathlib import Path
import re
import subprocess
import sys
from urllib.parse import unquote, urlsplit

REPO = Path(__file__).resolve().parents[2]
CERT = Path("/etc/leamout/certs")
PRODUCTION_SECRETS = (
    "POSTGRES_PASSWORD", "MINIO_ROOT_PASSWORD",
    "FREESWITCH_ESL_PASSWORD", "MEDIA_CONTROL_TOKEN",
    "MEDIA_TOKEN_SECRET", "TURN_AUTH_SECRET",
    "MINIO_APP_SECRET_KEY",
)
KEY = re.compile(r"^[A-Za-z_][A-Za-z_0-9]*$")
DOMAIN = re.compile(r"^(?=.{4,253}$)(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[A-Za-z]{2,63}$")


def read_env(path):
    values = {}
    if not path.is_file():
        raise ValueError(f"missing environment file: {path}")
    for lineno, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if line.startswith("export "):
            line = line[7:].strip()
        if "=" not in line:
            raise ValueError(f"{path}:{lineno}: expected KEY=value")
        key, value = line.split("=", 1)
        key = key.strip()
        value = value.strip()
        if not KEY.fullmatch(key):
            raise ValueError(f"{path}:{lineno}: invalid environment key")
        if key in values:
            raise ValueError(f"{path}:{lineno}: duplicate key {key}")
        if value.startswith(("'", '"')) and value.endswith(value[0]) and len(value) >= 2:
            value = value[1:-1]
        values[key] = value
    return values


def bad_secret(value):
    low = (value or "").lower()
    return len(value or "") < 24 or low.startswith(("change-me", "your-", "example-"))


def check_config(root, app):
    """Return redacted errors; never include secret contents in messages."""
    errors = []
    def fail(reason):
        errors.append(reason)

    if app.get("APP_ENV") != "production":
        fail("APP_ENV must be production")
    domain = app.get("DOMAIN", "")
    if not DOMAIN.fullmatch(domain) or domain.endswith((".example", ".test", ".invalid")):
        fail("DOMAIN must be a real public domain")
    try:
        public_ip = ipaddress.ip_address(root.get("PUBLIC_IP", ""))
        if public_ip.version != 4 or not public_ip.is_global:
            fail("PUBLIC_IP must be a globally routable IPv4 address")
    except ValueError:
        fail("PUBLIC_IP must be a valid IPv4 address")

    for name in PRODUCTION_SECRETS:
        value = (root if name in ("POSTGRES_PASSWORD", "MINIO_ROOT_PASSWORD") else app).get(name, "")
        if bad_secret(value):
            fail(f"{name} must be a non-default secret of at least 24 characters")

    encrypted = app.get("ENCRYPTION_KEY", "")
    try:
        raw = base64.urlsafe_b64decode(encrypted + "=" * (-len(encrypted) % 4))
        if len(raw) not in (16, 24, 32) or raw == bytes(len(raw)):
            fail("ENCRYPTION_KEY must be a nonzero 16/24/32-byte URL-safe base64 key")
    except (ValueError, base64.binascii.Error):
        fail("ENCRYPTION_KEY must be valid URL-safe base64")
    if any(c not in "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_" for c in encrypted):
        fail("ENCRYPTION_KEY must be unpadded URL-safe base64")

    database = urlsplit(app.get("DATABASE_URL", ""))
    if database.scheme not in ("postgres", "postgresql") or database.hostname != "postgres" or database.path != "/leamout":
        fail("DATABASE_URL must target the internal Compose PostgreSQL service")
    elif database.username != "leamout" or unquote(database.password or "") != root.get("POSTGRES_PASSWORD"):
        fail("DATABASE_URL username/password must match Compose PostgreSQL configuration")
    if app.get("REDIS_URL") != "redis://redis:6379/0":
        fail("REDIS_URL must target the internal Redis service")
    if app.get("NATS_URL") != "nats://nats:4222":
        fail("NATS_URL must target the internal NATS service")

    if app.get("MINIO_APP_ACCESS_KEY") in (None, "", root.get("MINIO_ROOT_USER")):
        fail("MINIO_APP_ACCESS_KEY must be a separately provisioned non-root MinIO user")
    if app.get("MEDIA_CONTROL_URL") != "http://media:8090":
        fail("MEDIA_CONTROL_URL must point to the private Media Runtime")
    if app.get("FREESWITCH_ESL_ADDRESS") != "freeswitch:8021":
        fail("FREESWITCH_ESL_ADDRESS must be internal to Compose")

    for name in ("FRONTEND_URL", "BACKEND_URL"):
        url = urlsplit(app.get(name, ""))
        if url.scheme != "https" or not url.hostname or url.username or url.password:
            fail(f"{name} must use a public HTTPS origin")
    if app.get("BACKEND_URL", "").rstrip("/") != f"https://api.{domain}":
        fail("BACKEND_URL must be the externally reachable Caddy API origin")
    frontend = app.get("FRONTEND_URL", "").rstrip("/")
    allowed_origins = [v.strip() for v in app.get("CORS_ORIGINS", "").split(",")]
    if not frontend or frontend not in allowed_origins or any(
        o and (not o.startswith("https://") or "localhost" in o) for o in allowed_origins
    ):
        fail("CORS_ORIGINS must include FRONTEND_URL and contain only HTTPS origins")
    if "turn:" not in app.get("TURN_PUBLIC_URLS", "") or f"turn.{domain}" not in app.get("TURN_PUBLIC_URLS", ""):
        fail("TURN_PUBLIC_URLS must reference the deployment TURN hostname")
    if root.get("TURN_REALM") != f"turn.{domain}":
        fail("TURN_REALM must match the TURN hostname")

    # Warn about controls that cannot be established from a configuration file.
    return errors


def check_files(root_path, app_path):
    problems = []
    for path in (root_path, app_path):
        if path.stat().st_mode & 0o077:
            problems.append(f"{path}: environment file must not be group/world accessible (chmod 600)")
    for name in ("fullchain.pem", "privkey.pem"):
        if not (CERT / name).is_file():
            problems.append(f"missing host SIP/TURN certificate: {CERT / name}")
    if not problems:
        certificate = subprocess.run(
            ["openssl", "x509", "-in", str(CERT / "fullchain.pem"), "-checkend", "604800", "-noout"],
            capture_output=True, text=True, check=False,
        )
        if certificate.returncode != 0:
            problems.append("SIP/TURN TLS certificate is invalid or expires within seven days")
    return problems


def check_compose(root_path, app_path):
    env = dict(os.environ)
    env["APP_ENV_FILE"] = str(app_path.resolve())
    cmd = [
        "docker", "compose", "--env-file", str(root_path.resolve()),
        "-f", str(REPO / "deploy/compose.yaml"),
        "-f", str(REPO / "deploy/cloud/compose.yaml"),
        "config", "--quiet",
    ]
    try:
        result = subprocess.run(cmd, cwd=REPO, env=env, capture_output=True,
                                text=True, timeout=30, check=False)
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return ["Docker Compose is unavailable or timed out"]
    # Compose errors might echo secret environment values: deliberately redact.
    return [] if result.returncode == 0 else ["Docker Compose configuration validation failed"]


def main(argv=None):
    parser = argparse.ArgumentParser(description="Leamout Cloud private-beta preflight")
    parser.add_argument("--root-env", type=Path, default=REPO / ".env")
    parser.add_argument("--app-env", type=Path, default=REPO / "server/.env")
    parser.add_argument("--config-only", action="store_true",
                        help="check settings, not Docker or host certificates")
    args = parser.parse_args(argv)
    try:
        root, app = read_env(args.root_env), read_env(args.app_env)
        problems = check_config(root, app)
        if not args.config_only:
            problems.extend(check_files(args.root_env, args.app_env))
            problems.extend(check_compose(args.root_env, args.app_env))
    except (OSError, ValueError) as error:
        print(f"FAIL preflight: {error}", file=sys.stderr)
        return 1
    for problem in problems:
        print(f"FAIL preflight: {problem}", file=sys.stderr)
    if problems:
        return 1
    print("PASS Cloud configuration preflight (not a deployment or live-call test)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
