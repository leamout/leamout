"""Offline regression tests for the operator-only deployment preflight."""

import base64
from pathlib import Path
import tempfile
import unittest

from preflight import read_env, validate


DOMAIN = "leamout.com"
STRONG = "X9mR5vQ2tS7zB4nL8pA6cY3wF1dK0hJ7"
POSTGRES_PASSWORD = "P9ostgres5SecreT2ValueZ"
ENCRYPTION_KEY = base64.urlsafe_b64encode(bytes(range(1, 33))).decode("ascii")


def valid_settings():
    infra = {
        "DOMAIN": DOMAIN,
        "PUBLIC_IP": "8.8.8.8",
        "TURN_REALM": f"turn.{DOMAIN}",
        "POSTGRES_PASSWORD": POSTGRES_PASSWORD,
        "KEYCLOAK_DB_PASSWORD": STRONG,
        "KEYCLOAK_ADMIN_PASSWORD": STRONG,
    }
    app = {
        "APP_ENV": "production",
        "DOMAIN": DOMAIN,
        "FRONTEND_URL": f"https://console.{DOMAIN}",
        "BACKEND_URL": f"https://api.{DOMAIN}",
        "BACKOFFICE_URL": f"https://backoffice.{DOMAIN}",
        "KEYCLOAK_ISSUER": f"https://auth.{DOMAIN}/realms/leamout",
        "CORS_ORIGINS": f"https://console.{DOMAIN}",
        "TURN_PUBLIC_URLS": f"stun:turn.{DOMAIN}:3478,turn:turn.{DOMAIN}:3478?transport=udp",
        "FREESWITCH_ESL_PASSWORD": STRONG,
        "MEDIA_CONTROL_TOKEN": STRONG,
        "MEDIA_TOKEN_SECRET": STRONG,
        "TURN_AUTH_SECRET": STRONG,
        "KEYCLOAK_CLIENT_SECRET": STRONG,
        "ENCRYPTION_KEY": ENCRYPTION_KEY,
        "DATABASE_URL": f"postgres://leamout:{POSTGRES_PASSWORD}@postgres:5432/leamout?sslmode=disable",
        "POLAR_ACCESS_TOKEN": "",
        "POLAR_WEBHOOK_SECRET": "",
    }
    return infra, app


class DeploymentPreflightTests(unittest.TestCase):
    def test_accepts_explicit_production_settings(self):
        infra, app = valid_settings()
        self.assertEqual(validate(infra, app), [])

    def test_rejects_example_credentials_and_ip(self):
        infra, app = valid_settings()
        infra["PUBLIC_IP"] = "203.0.113.10"
        infra["POSTGRES_PASSWORD"] = "change-me-postgres"
        app["ENCRYPTION_KEY"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
        errors = validate(infra, app)
        self.assertTrue(any("PUBLIC_IP" in error for error in errors))
        self.assertTrue(any("POSTGRES_PASSWORD" in error for error in errors))
        self.assertTrue(any("ENCRYPTION_KEY" in error for error in errors))

    def test_rejects_mismatched_postgres_password(self):
        infra, app = valid_settings()
        app["DATABASE_URL"] = app["DATABASE_URL"].replace(POSTGRES_PASSWORD, STRONG)
        self.assertTrue(any("DATABASE_URL password" in error for error in validate(infra, app)))

    def test_rejects_insecure_origins_and_provider_config(self):
        infra, app = valid_settings()
        app["CORS_ORIGINS"] = "http://localhost:3000,*"
        app["POLAR_ACCESS_TOKEN"] = STRONG
        errors = validate(infra, app)
        self.assertTrue(any("CORS_ORIGINS" in error for error in errors))
        self.assertTrue(any("POLAR_ACCESS_TOKEN" in error for error in errors))

    def test_rejects_extra_https_origin(self):
        infra, app = valid_settings()
        app["CORS_ORIGINS"] += ",https://unreviewed.example.com"
        self.assertTrue(any("CORS_ORIGINS" in error for error in validate(infra, app)))

    def test_rejects_example_polar_credentials(self):
        infra, app = valid_settings()
        app["POLAR_ACCESS_TOKEN"] = "change-me-polar-access-token"
        app["POLAR_WEBHOOK_SECRET"] = "change-me-polar-webhook-secret"
        errors = validate(infra, app)
        self.assertTrue(any("POLAR_ACCESS_TOKEN" in error for error in errors))
        self.assertTrue(any("POLAR_WEBHOOK_SECRET" in error for error in errors))

    def test_rejects_cross_domain_configuration(self):
        infra, app = valid_settings()
        app["DOMAIN"] = "other.example"
        infra["TURN_REALM"] = "turn.other.example"
        errors = validate(infra, app)
        self.assertTrue(any("DOMAIN differs" in error for error in errors))
        self.assertTrue(any("TURN_REALM" in error for error in errors))

    def test_env_file_must_be_private_and_unique(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / ".env"
            path.write_text("DOMAIN=leamout.com\nDOMAIN=other.com\n", encoding="utf-8")
            path.chmod(0o600)
            with self.assertRaisesRegex(ValueError, "duplicate"):
                read_env(path)
            path.chmod(0o644)
            with self.assertRaisesRegex(ValueError, "chmod 600"):
                read_env(path)


if __name__ == "__main__":
    unittest.main()
