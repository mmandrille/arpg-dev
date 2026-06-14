#!/usr/bin/env python3
from __future__ import annotations

import os
import re
import subprocess
import sys
from pathlib import Path


ROOT = Path(os.environ.get("ROOT", Path(__file__).resolve().parents[1]))
BASELINE = Path(
    os.environ.get(
        "EXTRACTION_COUPLING_BASELINE",
        ROOT / ".maintainability" / "extraction-coupling-baseline.tsv",
    )
)
COUPLING_PATTERN = re.compile(r"helpers\s*=\s*globals\(\)")
SOURCE_SUFFIXES = {".go", ".gd", ".py", ".sh", ".mk"}
EXEMPT_PREFIXES = ("docs/", "shared/golden/", "client/.godot/", "client/imports/")


def is_source_file(path: str) -> bool:
    name = Path(path).name
    return Path(path).suffix in SOURCE_SUFFIXES or name == "Makefile"


def is_exempt_path(path: str) -> bool:
    return path.startswith(EXEMPT_PREFIXES)


def git_ls_files() -> list[str]:
    result = subprocess.run(
        ["git", "ls-files"],
        cwd=ROOT,
        text=True,
        check=True,
        capture_output=True,
    )
    return [line for line in result.stdout.splitlines() if line]


def load_baseline() -> dict[str, int]:
    if not BASELINE.exists():
        raise SystemExit(f"missing extraction coupling baseline: {BASELINE}")
    entries: dict[str, int] = {}
    for line_no, raw in enumerate(BASELINE.read_text().splitlines(), start=1):
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        parts = line.split("\t")
        if len(parts) < 2:
            raise SystemExit(f"{BASELINE}:{line_no}: expected '<path>\\t<count>'")
        path, count = parts[0], parts[1]
        try:
            entries[path] = int(count)
        except ValueError as exc:
            raise SystemExit(f"{BASELINE}:{line_no}: invalid count {count!r}") from exc
    return entries


def count_couplings(path: str) -> int:
    full_path = ROOT / path
    if not full_path.exists() or not is_source_file(path) or is_exempt_path(path):
        return 0
    text = full_path.read_text(errors="ignore")
    return len(COUPLING_PATTERN.findall(text))


def main() -> int:
    baseline = load_baseline()
    tracked = set(git_ls_files())
    failures: list[str] = []
    coupled_files = 0
    coupled_occurrences = 0

    for path, baseline_count in sorted(baseline.items()):
        if path not in tracked:
            failures.append(f"{path}: baseline entry exists but file is no longer tracked; remove it.")
            continue
        actual = count_couplings(path)
        coupled_files += 1
        coupled_occurrences += actual
        if actual > baseline_count:
            failures.append(
                f"{path}: {actual} helper-global injections exceeds baseline {baseline_count}. "
                "Pass a typed context or explicit helper set instead of laundering the origin module namespace."
            )
        elif actual < baseline_count:
            failures.append(
                f"{path}: {actual} helper-global injections is below baseline {baseline_count}; "
                "lower or remove the extraction coupling baseline in the same slice."
            )

    for path in sorted(tracked - set(baseline)):
        if not is_source_file(path) or is_exempt_path(path):
            continue
        actual = count_couplings(path)
        if actual > 0:
            failures.append(
                f"{path}: {actual} helper-global injections are unbaselined. "
                "New extractions must be importable and unit-testable without importing the file they came from."
            )

    if failures:
        print(
            f"coupled helper injections: {coupled_files} files, {coupled_occurrences} occurrences (target: down)"
        )
        print("Extraction coupling ratchet failed:", file=sys.stderr)
        for failure in failures:
            print(f"  - {failure}", file=sys.stderr)
        return 1

    print("extraction-coupling ratchet passed")
    print(f"coupled helper injections: {coupled_files} files, {coupled_occurrences} occurrences (target: down)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
