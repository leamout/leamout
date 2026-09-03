#!/usr/bin/env python3

from __future__ import annotations

import json
import sys
from pathlib import Path

TOKENS = {
    "server": "@@IMAGE_SERVER@@",
    "worker": "@@IMAGE_WORKER@@",
    "opensips": "@@IMAGE_OPENSIPS@@",
    "rtpengine": "@@IMAGE_RTPENGINE@@",
    "freeswitch": "@@IMAGE_FREESWITCH@@",
    "coturn": "@@IMAGE_COTURN@@",
    "postgres": "@@IMAGE_POSTGRES@@",
    "redis": "@@IMAGE_REDIS@@",
    "nats": "@@IMAGE_NATS@@",
    "atlas": "@@IMAGE_ATLAS@@",
}


def main() -> int:
    if len(sys.argv) != 4:
        print(
            f"usage: {Path(sys.argv[0]).name} <release-manifest.json> <compose-template> <output>",
            file=sys.stderr,
        )
        return 2

    manifest_path = Path(sys.argv[1])
    template_path = Path(sys.argv[2])
    output_path = Path(sys.argv[3])

    try:
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
        images = manifest["images"]
        rendered = template_path.read_text(encoding="utf-8")
        for name, token in TOKENS.items():
            image = images[name]
            if not isinstance(image, str) or "@sha256:" not in image:
                raise ValueError(f"manifest image {name!r} is not digest-pinned")
            if token not in rendered:
                raise ValueError(f"runtime template is missing {token}")
            rendered = rendered.replace(token, image)
        if "@@IMAGE_" in rendered:
            raise ValueError("runtime template contains unresolved image tokens")
    except (OSError, KeyError, json.JSONDecodeError, ValueError) as exc:
        print(f"render runtime failed: {exc}", file=sys.stderr)
        return 1

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(rendered, encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
