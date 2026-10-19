#!/usr/bin/env python3
"""Reject malformed Docker Compose file arguments in acceptance suites."""

import ast
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

SUITES = (
    "tests/byoc-v1/acceptance.py",
    "tests/voice-v1/acceptance.py",
    "tests/voice-agent-v1/acceptance.py",
    "tests/graceful-drain/acceptance.py",
)


def compose_command(path: Path) -> list[str]:
    tree = ast.parse(path.read_text(encoding="utf-8"), filename=str(path))

    for statement in tree.body:
        if not isinstance(statement, ast.Assign):
            continue
        if not any(
            isinstance(target, ast.Name) and target.id == "COMPOSE"
            for target in statement.targets
        ):
            continue
        value = ast.literal_eval(statement.value)
        if not isinstance(value, list) or not all(
            isinstance(item, str) for item in value
        ):
            raise ValueError(f"{path}: COMPOSE must be a literal string list")
        return value

    raise ValueError(f"{path}: missing COMPOSE command")


def validate(path: Path) -> None:
    command = compose_command(path)
    if command[:2] != ["docker", "compose"]:
        raise ValueError(f"{path}: expected docker compose command")

    index = 2
    file_count = 0
    while index < len(command):
        if command[index] != "-f":
            raise ValueError(f"{path}: unexpected Compose argument {command[index]!r}")
        if index + 1 >= len(command):
            raise ValueError(f"{path}: -f requires a Compose file path")
        filename = command[index + 1]
        if filename.startswith("-"):
            raise ValueError(f"{path}: invalid Compose file path {filename!r}")
        if not (ROOT / filename).is_file():
            raise ValueError(f"{path}: missing Compose file {filename!r}")
        file_count += 1
        index += 2

    if file_count < 2:
        raise ValueError(f"{path}: expected base and suite Compose files")


def main() -> None:
    for suite in SUITES:
        validate(ROOT / suite)
        print(f"PASS {suite}: Compose arguments reference existing files")


if __name__ == "__main__":
    main()
