#!/usr/bin/env python3

import os
import pathlib
import stat
import sys
import tempfile


def main() -> int:
    if len(sys.argv) != 4:
        print("usage: set-env.py <env-file> <key> <source-env-var>", file=sys.stderr)
        return 2

    env_path = pathlib.Path(sys.argv[1])
    key = sys.argv[2]
    source_name = sys.argv[3]
    value = os.environ.get(source_name)
    if value is None or value == "":
        print(f"{source_name} is required", file=sys.stderr)
        return 2
    if "\n" in value or "\r" in value:
        print(f"{source_name} must be a single line", file=sys.stderr)
        return 2
    if not env_path.is_file():
        print(f"environment file does not exist: {env_path}", file=sys.stderr)
        return 2

    lines = env_path.read_text(encoding="utf-8").splitlines()
    replacement = f"{key}={value}"
    found = False
    output: list[str] = []
    for line in lines:
        stripped = line.lstrip()
        if stripped.startswith(f"{key}=") and not stripped.startswith("#"):
            if found:
                print(f"environment file contains duplicate {key} entries", file=sys.stderr)
                return 2
            output.append(replacement)
            found = True
        else:
            output.append(line)

    if not found:
        output.append(replacement)

    mode = stat.S_IMODE(env_path.stat().st_mode)
    fd, temporary = tempfile.mkstemp(prefix=env_path.name + ".", dir=env_path.parent)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            handle.write("\n".join(output) + "\n")
            handle.flush()
            os.fsync(handle.fileno())
        os.chmod(temporary, mode)
        os.replace(temporary, env_path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
