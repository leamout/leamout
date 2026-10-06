"""Write a disposable Compose application env file without using deployment secrets."""

import os
from pathlib import Path
import sys

values = {
    "APP_ENV": "production",
    "DOMAIN": os.environ["DOMAIN"],
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
