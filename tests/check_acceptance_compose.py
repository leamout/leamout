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
        if not isinstance(statement.value, ast.List):
            raise ValueError(f"{path}: COMPOSE must be a list")

        command = []
        for item in statement.value.elts:
            if isinstance(item, ast.Constant) and isinstance(item.value, str):
                command.append(item.value)
                continue

            # Acceptance suites use a disposable file generated at runtime.
            if (
                isinstance(item, ast.Subscript)
                and isinstance(item.value, ast.Attribute)
                and isinstance(item.value.value, ast.Name)
                and item.value.value.id == "os"
                and item.value.attr == "environ"
                and isinstance(item.slice, ast.Constant)
                and item.slice.value == "APP_ENV_FILE"
            ):
                command.append("$APP_ENV_FILE")
                continue

            raise ValueError(f"{path}: unexpected dynamic Compose argument")
        return command

    raise ValueError(f"{path}: missing COMPOSE command")


def validate(path: Path) -> None:
    command = compose_command(path)
    if command[:2] != ["docker", "compose"]:
        raise ValueError(f"{path}: expected docker compose command")

    if command[2:4] != ["--env-file", "$APP_ENV_FILE"]:
        raise ValueError(f"{path}: Compose must load the generated acceptance env file")

    index = 4
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
