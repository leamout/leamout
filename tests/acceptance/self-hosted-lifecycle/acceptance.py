#!/usr/bin/env python3

import argparse
import json
from pathlib import Path


def require(path: Path) -> None:
    if not path.is_file():
        raise SystemExit(f"required lifecycle artifact is missing: {path}")


def release_version(runtime: Path) -> str:
    require(runtime / "release.json")
    return json.loads((runtime / "release.json").read_text())["release_version"]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("state", choices=("initialized", "updated", "uninstalled"))
    parser.add_argument("--version")
    parser.add_argument("--config", type=Path, default=Path("/etc/leamout"))
    parser.add_argument("--data", type=Path, default=Path("/var/lib/leamout"))
    args = parser.parse_args()

    require(args.config / "leamout.env")
    require(args.data / "deployment.json")
    runtime = args.data / "runtime"

    if args.state == "uninstalled":
        if runtime.exists():
            raise SystemExit("normal uninstall did not remove runtime software")
        return

    if not args.version:
        raise SystemExit("--version is required for initialized and updated states")
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
