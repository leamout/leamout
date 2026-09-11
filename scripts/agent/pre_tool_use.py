#!/usr/bin/env python3
"""Deny a small set of destructive shell operations for repository agents.

The hook intentionally returns an empty decision for normal commands so the host's
usual permission policy still applies. It is a guardrail, not an allowlist.
"""

from __future__ import annotations

import json
import re
import sys
from typing import Any


def strings(value: Any) -> list[str]:
    if isinstance(value, str):
        return [value]
    if isinstance(value, dict):
        out: list[str] = []
        for item in value.values():
            out.extend(strings(item))
        return out
    if isinstance(value, list):
        out = []
        for item in value:
            out.extend(strings(item))
        return out
    return []


def main() -> int:
    try:
        payload = json.load(sys.stdin)
    except Exception:
        print("{}")
        return 0

    tool = str(payload.get("toolName") or payload.get("tool_name") or "").lower()
    if tool not in {"bash", "powershell"}:
        print("{}")
        return 0

    args = payload.get("toolArgs")
    if args is None:
        args = payload.get("tool_input")
    command = "\n".join(strings(args))

    blocked = [
        (r"(?i)\bgit\s+reset\s+--hard\b", "git reset --hard can discard repository work"),
        (r"(?i)\bgit\s+clean\s+-[^\s]*f", "git clean with force can delete untracked work"),
        (r"(?i)\bgit\s+push\b[^\n]*--force(?:-with-lease)?\b", "force-pushing can rewrite shared history"),
        (r"(?i)\brm\s+-[^\s]*r[^\s]*f[^\s]*\s+/(?:\s|$|\*)", "recursive deletion from filesystem root is not allowed"),
        (r"(?i)\bdrop\s+(?:database|schema)\b", "dropping a database or schema requires explicit human execution"),
    ]

    for pattern, reason in blocked:
        if re.search(pattern, command):
            print(json.dumps({
                "permissionDecision": "deny",
                "permissionDecisionReason": reason,
            }))
            return 0

    print("{}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
