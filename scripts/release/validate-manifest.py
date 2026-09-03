#!/usr/bin/env python3

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

SEMVER = re.compile(r"^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$")
SHA40 = re.compile(r"^[0-9a-f]{40}$")
SHA256 = re.compile(r"^[0-9a-f]{64}$")
IMAGE = re.compile(r"^.+@sha256:([0-9a-f]{64})$")
MIGRATION = re.compile(r"^[0-9]{3}_[a-z0-9_]+\.sql$")

REQUIRED_IMAGES = {
    "server",
    "worker",
    "opensips",
    "rtpengine",
    "freeswitch",
    "coturn",
    "postgres",
    "redis",
    "nats",
    "atlas",
}

SUPPORTED_HOSTS = {
    ("ubuntu", "24.04", "amd64"),
    ("debian", "13", "amd64"),
}


def fail(message: str) -> None:
    raise ValueError(message)


def require_keys(value: dict, expected: set[str], context: str) -> None:
    actual = set(value)
    missing = expected - actual
    extra = actual - expected
    if missing:
        fail(f"{context}: missing fields: {', '.join(sorted(missing))}")
    if extra:
        fail(f"{context}: unknown fields: {', '.join(sorted(extra))}")


def nonzero_hex(value: str) -> bool:
    return any(ch != "0" for ch in value)


def validate(path: Path) -> None:
    with path.open("r", encoding="utf-8") as handle:
        manifest = json.load(handle)

    if not isinstance(manifest, dict):
        fail("manifest root must be an object")

    require_keys(
        manifest,
        {
            "schema_version",
            "release_version",
            "channel",
            "source_commit",
            "minimum_cli_version",
            "supported_hosts",
            "database",
            "cli_artifacts",
            "images",
        },
        "manifest",
    )

    if manifest["schema_version"] != 1:
        fail("schema_version must be 1")

    release_version = manifest["release_version"]
    minimum_cli_version = manifest["minimum_cli_version"]
    if not isinstance(release_version, str) or not SEMVER.fullmatch(release_version):
        fail("release_version must be semantic-version shaped and omit a leading v")
    if not isinstance(minimum_cli_version, str) or not SEMVER.fullmatch(minimum_cli_version):
        fail("minimum_cli_version must be semantic-version shaped and omit a leading v")

    if manifest["channel"] not in {"stable", "preview"}:
        fail("channel must be stable or preview")

    source_commit = manifest["source_commit"]
    if not isinstance(source_commit, str) or not SHA40.fullmatch(source_commit):
        fail("source_commit must be a lowercase 40-character Git SHA")
    if not nonzero_hex(source_commit):
        fail("source_commit cannot be an all-zero placeholder")

    hosts = manifest["supported_hosts"]
    if not isinstance(hosts, list) or not hosts:
        fail("supported_hosts must be a non-empty array")

    seen_hosts: set[tuple[str, str, str]] = set()
    for index, host in enumerate(hosts):
        if not isinstance(host, dict):
            fail(f"supported_hosts[{index}] must be an object")
        require_keys(host, {"os", "version", "arch"}, f"supported_hosts[{index}]")
        host_tuple = (host["os"], host["version"], host["arch"])
        if host_tuple not in SUPPORTED_HOSTS:
            fail(
                f"supported_hosts[{index}] is outside the Phase 1 support matrix: "
                f"{host_tuple[0]} {host_tuple[1]} {host_tuple[2]}"
            )
        if host_tuple in seen_hosts:
            fail(f"supported_hosts[{index}] duplicates an earlier host")
        seen_hosts.add(host_tuple)

    database = manifest["database"]
    if not isinstance(database, dict):
        fail("database must be an object")
    require_keys(database, {"migration"}, "database")
    migration = database["migration"]
    if not isinstance(migration, str) or not MIGRATION.fullmatch(migration):
        fail("database.migration must be a numbered migration filename")

    artifacts = manifest["cli_artifacts"]
    if not isinstance(artifacts, list) or not artifacts:
        fail("cli_artifacts must be a non-empty array")

    expected_filename = f"leamout_{release_version}_linux_amd64.tar.gz"
    seen_artifacts: set[tuple[str, str]] = set()
    for index, artifact in enumerate(artifacts):
        if not isinstance(artifact, dict):
            fail(f"cli_artifacts[{index}] must be an object")
        require_keys(
            artifact,
            {"os", "arch", "filename", "sha256"},
            f"cli_artifacts[{index}]",
        )
        if artifact["os"] != "linux" or artifact["arch"] != "amd64":
            fail(f"cli_artifacts[{index}] is outside the Phase 1 artifact matrix")
        key = (artifact["os"], artifact["arch"])
        if key in seen_artifacts:
            fail(f"cli_artifacts[{index}] duplicates an earlier OS/architecture")
        seen_artifacts.add(key)
        if artifact["filename"] != expected_filename:
            fail(
                f"cli_artifacts[{index}].filename must be {expected_filename!r} "
                f"for release {release_version}"
            )
        checksum = artifact["sha256"]
        if not isinstance(checksum, str) or not SHA256.fullmatch(checksum):
            fail(f"cli_artifacts[{index}].sha256 must be 64 lowercase hex characters")
        if not nonzero_hex(checksum):
            fail(f"cli_artifacts[{index}].sha256 cannot be an all-zero placeholder")

    images = manifest["images"]
    if not isinstance(images, dict):
        fail("images must be an object")
    require_keys(images, REQUIRED_IMAGES, "images")

    for name, reference in images.items():
        if not isinstance(reference, str):
            fail(f"images.{name} must be a string")
        match = IMAGE.fullmatch(reference)
        if match is None:
            fail(
                f"images.{name} must be an immutable OCI digest reference "
                "(<registry>/<repository>@sha256:<digest>)"
            )
        digest = match.group(1)
        if not nonzero_hex(digest):
            fail(f"images.{name} cannot use an all-zero digest placeholder")


def main() -> int:
    if len(sys.argv) != 2:
        print(f"usage: {Path(sys.argv[0]).name} <release-manifest.json>", file=sys.stderr)
        return 2

    path = Path(sys.argv[1])
    try:
        validate(path)
    except (OSError, json.JSONDecodeError, ValueError) as exc:
        print(f"release manifest invalid: {exc}", file=sys.stderr)
        return 1

    print(f"release manifest valid: {path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
