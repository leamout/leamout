#!/usr/bin/env python3
"""Run lightweight repository quality checks when an agent tries to stop."""

from __future__ import annotations

import json
import os
import subprocess
import sys
from pathlib import Path


def run(*args: str) -> tuple[int, str]:
    proc = subprocess.run(args, text=True, capture_output=True, check=False)
    return proc.returncode, (proc.stdout + proc.stderr).strip()


def changed_files() -> list[str]:
    files: set[str] = set()
    commands = [
        ("git", "diff", "--name-only", "--diff-filter=ACMR"),
        ("git", "diff", "--cached", "--name-only", "--diff-filter=ACMR"),
    ]

    code, base = run("git", "merge-base", "HEAD", "origin/main")
    if code == 0 and base:
        commands.append(("git", "diff", "--name-only", "--diff-filter=ACMR", f"{base}...HEAD"))

    for command in commands:
        code, output = run(*command)
        if code == 0:
            files.update(line.strip() for line in output.splitlines() if line.strip())
    return sorted(files)


def block(reason: str) -> int:
    print(json.dumps({"decision": "block", "reason": reason}))
    return 0


def main() -> int:
    try:
        json.load(sys.stdin)
    except Exception:
        pass

    checks = [
        ("git", "diff", "--check"),
        ("git", "diff", "--cached", "--check"),
    ]
    code, base = run("git", "merge-base", "HEAD", "origin/main")
    if code == 0 and base:
        checks.append(("git", "diff", "--check", f"{base}...HEAD"))

    failures: list[str] = []
    for command in checks:
        code, output = run(*command)
        if code != 0:
            failures.append(output or "diff whitespace check failed")

    go_files = [path for path in changed_files() if path.endswith(".go") and Path(path).is_file()]
    if go_files:
        code, output = run("gofmt", "-l", *go_files)
        if code != 0:
            failures.append(output or "gofmt failed")
        elif output:
            failures.append("gofmt is required for: " + ", ".join(output.splitlines()))

    if failures:
        return block(
            "Repository quality checks are not clean. Fix these before finishing:\n- "
            + "\n- ".join(failures)
        )

    print(json.dumps({"decision": "allow"}))
    return 0


if __name__ == "__main__":
    os.chdir(os.environ.get("GITHUB_WORKSPACE") or os.getcwd())
    raise SystemExit(main())
