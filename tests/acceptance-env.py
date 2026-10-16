"""Write a disposable Compose application env file without using deployment secrets."""

import os
from pathlib import Path
import shutil
import subprocess
import sys

values = {
    "APP_ENV": "production",
    "DOMAIN": os.environ["DOMAIN"],
    "FRONTEND_URL": "https://console.acceptance.invalid",
    "BACKEND_URL": "http://server:8080",
    "KEYCLOAK_ISSUER": "https://keycloak:8443/realms/leamout",
    "KEYCLOAK_CLIENT_ID": "leamout-console",
    "KEYCLOAK_CLIENT_SECRET": "acceptance-keycloak-client-not-live",
    "DATABASE_URL": (
        "postgres://leamout:"
        + os.environ["POSTGRES_PASSWORD"]
        + "@postgres:5432/leamout?sslmode=disable"
    ),
    "REDIS_URL": "redis://redis:6379/0",
    "NATS_URL": "nats://nats:4222",
    "FREESWITCH_ESL_ADDRESS": "freeswitch:8021",
    "MEDIA_CONTROL_URL": "http://media:8090",
    "AWS_ACCESS_KEY_ID": "acceptance-not-live",
    "AWS_SECRET_ACCESS_KEY": "acceptance-not-live",
}

for key in (
    "CORS_ORIGINS",
    "FREESWITCH_ESL_PASSWORD",
    "MEDIA_CONTROL_TOKEN",
    "MEDIA_TOKEN_SECRET",
    "ENCRYPTION_KEY",
    "TURN_AUTH_SECRET",
    "TURN_PUBLIC_URLS",
    "MINIO_APP_ACCESS_KEY",
    "MINIO_APP_SECRET_KEY",
):
    values[key] = os.environ[key]

lines = []
for key, value in values.items():
    if "\n" in value or "\r" in value:
        raise SystemExit(f"{key} must not contain newlines")
    escaped = value.replace("\\", "\\\\").replace("'", "\\'")
    lines.append(f"{key}='{escaped}'\n")

Path(sys.argv[1]).write_text("".join(lines), encoding="utf-8")

# Keep production HTTPS and issuer validation enabled for acceptance tests.
identity_dir = Path(sys.argv[1] + ".identity")
trust_dir = identity_dir / "trust"
trust_dir.mkdir(parents=True, exist_ok=True)
subprocess.run(
    [
        "openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes",
        "-days", "1", "-subj", "/CN=keycloak",
        "-addext", "subjectAltName=DNS:keycloak",
        "-keyout", str(identity_dir / "key.pem"),
        "-out", str(identity_dir / "certificate.pem"),
    ],
    check=True,
    stdout=subprocess.DEVNULL,
    stderr=subprocess.DEVNULL,
)
# Keycloak runs as a non-root container user. This is a disposable test key.
(identity_dir / "key.pem").chmod(0o644)
shutil.copyfile(identity_dir / "certificate.pem", trust_dir / "keycloak.pem")
