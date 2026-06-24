#!/usr/bin/env python3
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DEFAULT_CODEMAP = ROOT / "docs" / "CODEMAP.md"
PATH_RE = re.compile(r"`([^`]+)`")


def looks_like_path(token: str) -> bool:
    if token.startswith("docs/specs/") or token.startswith("docs/plans/") or token.startswith("docs/as-built/"):
        return False
    return "/" in token and not token.startswith(("http://", "https://"))


def validate_codemap(path: Path, root: Path = ROOT) -> list[str]:
    errors: list[str] = []
    text = path.read_text()
    for token in PATH_RE.findall(text):
        if not looks_like_path(token):
            continue
        if not (root / token).exists():
            errors.append(f"missing path: {token}")

    for line_no, line in enumerate(text.splitlines(), start=1):
        if not line.startswith("|") or line.startswith("|---") or line.startswith("| Domain "):
            continue
        cells = [cell.strip() for cell in line.strip().strip("|").split("|")]
        if len(cells) < 6:
            continue
        domain = cells[0]
        path_count = sum(1 for token in PATH_RE.findall(line) if looks_like_path(token))
        if path_count == 0:
            errors.append(f"line {line_no}: domain row {domain!r} has no paths")
    return errors


def check_unlisted_files(repo_root: Path, codemap_text: str) -> list[str]:
    """Return paths for game/*.go (non-test) and client/scripts/*.gd not mentioned in CODEMAP."""
    errors: list[str] = []

    game_dir = repo_root / "server" / "internal" / "game"
    for f in sorted(game_dir.glob("*.go")):
        if f.name.endswith("_test.go"):
            continue
        if f.name not in codemap_text:
            errors.append(f"unlisted in CODEMAP: server/internal/game/{f.name}")

    scripts_dir = repo_root / "client" / "scripts"
    for f in sorted(scripts_dir.glob("*.gd")):
        if f.name not in codemap_text:
            errors.append(f"unlisted in CODEMAP: client/scripts/{f.name}")

    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate docs/CODEMAP.md path references.")
    parser.add_argument("codemap", nargs="?", default=str(DEFAULT_CODEMAP))
    args = parser.parse_args()

    codemap_path = Path(args.codemap)
    codemap_text = codemap_path.read_text()

    errors = validate_codemap(codemap_path)
    errors += check_unlisted_files(ROOT, codemap_text)

    if errors:
        for error in errors:
            print(error, file=sys.stderr)
        return 1
    print("codemap ok")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
