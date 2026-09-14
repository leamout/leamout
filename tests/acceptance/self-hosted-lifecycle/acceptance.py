#!/usr/bin/env python3

import argparse
import json
import subprocess
from pathlib import Path


def require(path: Path) -> None:
    if not path.is_file():
        raise SystemExit(f"required lifecycle artifact is missing: {path}")


def release_version(runtime: Path) -> str:
    require(runtime / "release.json")
    return json.loads((runtime / "release.json").read_text())["release_version"]


def run_cli(cli: Path, *args: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [str(cli), *args],
        check=False,
        stdin=subprocess.DEVNULL,
        capture_output=True,
        text=True,
    )


def verify_cli(cli: Path, version: str) -> None:
    require(cli)

    help_result = run_cli(cli, "--help")
    if help_result.returncode != 0:
        raise SystemExit(f"CLI help failed: {help_result.stderr}")
    for command in (
        "init",
        "up",
        "down",
        "status",
        "logs",
        "doctor",
        "help",
        "license",
        "certs",
        "backup",
        "restore",
        "update",
        "version",
    ):
        if command not in help_result.stdout:
            raise SystemExit(f"CLI help does not expose {command!r}")
    for cloud_only in ("DIDWW", "CommPeak", "Stripe", "Paystack", "managed-carrier"):
        if cloud_only.lower() in help_result.stdout.lower():
            raise SystemExit(f"Self-Hosted CLI help exposes Cloud capability {cloud_only!r}")

    version_command = run_cli(cli, "version")
    version_flag = run_cli(cli, "--version")
    version_short_flag = run_cli(cli, "-v")
    expected_version = f"leamout {version}"
    for result in (version_command, version_flag, version_short_flag):
        if result.returncode != 0 or expected_version not in result.stdout:
            raise SystemExit(f"CLI version contract failed: {result.stdout}{result.stderr}")
    if len({version_command.stdout, version_flag.stdout, version_short_flag.stdout}) != 1:
        raise SystemExit("CLI version command and flags returned different build information")

    unknown = run_cli(cli, "not-a-command")
    if unknown.returncode != 2 or "unknown command: not-a-command" not in unknown.stderr:
        raise SystemExit(f"CLI unknown-command contract failed: {unknown.stdout}{unknown.stderr}")

    restore = run_cli(cli, "restore", "/tmp/nonexistent-leamout-backup.tar.gz")
    if restore.returncode != 1 or "restore cancelled" not in restore.stderr:
        raise SystemExit("non-interactive restore did not fail closed without --force")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("state", choices=("cli", "initialized", "updated", "uninstalled"))
    parser.add_argument("--version")
    parser.add_argument("--cli", type=Path, default=Path("/usr/local/bin/leamout"))
    parser.add_argument("--config", type=Path, default=Path("/etc/leamout"))
    parser.add_argument("--data", type=Path, default=Path("/var/lib/leamout"))
    args = parser.parse_args()

    if args.state == "cli":
        if not args.version:
            raise SystemExit("--version is required for CLI acceptance")
        verify_cli(args.cli, args.version)
        return

    require(args.config / "leamout.env")
    require(args.data / "deployment.json")
    runtime = args.data / "runtime"

    if args.state == "uninstalled":
        if runtime.exists():
            raise SystemExit("normal uninstall did not remove runtime software")
        return

    if not args.version:
        raise SystemExit("--version is required for initialized and updated states")
    verify_cli(args.cli, args.version)
    require(runtime / "compose.yaml")
    if (runtime / "compose.yaml.tmpl").exists():
        raise SystemExit("installed runtime retained its Compose template")
    compose = (runtime / "compose.yaml").read_text()
    if "@@IMAGE_" in compose or "\n    build:" in compose:
        raise SystemExit("installed runtime is not release-pinned")
    if release_version(runtime) != args.version:
        raise SystemExit(f"installed runtime version is not {args.version}")

    if args.state == "updated":
        if (args.data / "runtime.previous").exists():
            raise SystemExit("successful update retained rollback runtime")
        backups = list((args.data / "backups").glob(f"pre-update-*-to-{args.version}-*.tar.gz"))
        if not backups:
            raise SystemExit("successful update did not retain a pre-update backup")
        if (args.data / "update.json").exists():
            raise SystemExit("successful update retained interrupted-update state")


if __name__ == "__main__":
    main()
