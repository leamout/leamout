#!/usr/bin/env python3

import hashlib
import json
import os
from pathlib import Path

version = os.environ["VERSION"]
digest_dir = Path(os.environ["DIGEST_DIR"])
output = Path(os.environ["OUTPUT"])
cli = output.parent / f"leamout_{version}_linux_amd64.tar.gz"
images = {name: (digest_dir / name).read_text().strip() for name in (
    "server", "worker", "opensips", "rtpengine", "freeswitch",
    "coturn", "postgres", "redis", "nats", "atlas",
)}
manifest = {
    "schema_version": 1,
    "release_version": version,
    "channel": "preview",
    "source_commit": os.environ["GITHUB_SHA"],
    "minimum_cli_version": version,
    "supported_hosts": [{"os": "ubuntu", "version": "24.04", "arch": "amd64"}],
    "database": {"migration": "048_create_provider_cdr_pages.sql"},
    "cli_artifacts": [{
        "os": "linux", "arch": "amd64", "filename": cli.name,
        "sha256": hashlib.sha256(cli.read_bytes()).hexdigest(),
    }],
    "images": images,
}
output.write_text(json.dumps(manifest, indent=2) + "\n")
