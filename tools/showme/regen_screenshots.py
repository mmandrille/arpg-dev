#!/usr/bin/env python3
"""Regenerate data-driven showme screenshots for visual regression review."""
from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from pathlib import Path

from tools.showme.screenshot_catalog import DEFAULT_SUITES, discover_jobs, suite_summary

ROOT = Path(__file__).resolve().parents[2]
RENDER_FOCUS = ROOT / "skills" / "showme" / "scripts" / "render_focus.py"


def _default_out_dir() -> Path:
    stamp = time.strftime("%Y%m%d-%H%M%S")
    return ROOT / ".artifacts" / "screenshots" / stamp


def _run_capture(job, output: Path, godot: str, dry_run: bool) -> tuple[bool, str]:
    cmd = [
        sys.executable,
        str(RENDER_FOCUS),
        "--focus",
        job.focus,
        "--output",
        str(output),
    ]
    cmd.extend(job.extra_args)
    if godot != "godot":
        cmd.extend(["--godot", godot])

    if dry_run:
        print("[regen-screenshots] dry-run:", " ".join(cmd))
        return True, ""

    print(f"[regen-screenshots] {job.suite}/{job.slug}")
    result = subprocess.run(cmd, cwd=ROOT)
    if result.returncode != 0:
        return False, f"render_focus failed ({result.returncode}) for {job.output_rel}"
    if not output.exists():
        return False, f"missing output for {job.output_rel}"
    return True, ""


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--suite",
        action="append",
        dest="suites",
        choices=DEFAULT_SUITES,
        help="Capture suite to run (repeatable; default: all suites)",
    )
    parser.add_argument("--out-dir", default="", help="Output directory for PNG captures")
    parser.add_argument("--godot", default="godot", help="Godot binary")
    parser.add_argument("--dry-run", action="store_true", help="Print planned captures without rendering")
    parser.add_argument("--fail-fast", action="store_true", help="Stop on first capture failure")
    parser.add_argument("--list", action="store_true", help="List suites and job counts")
    args = parser.parse_args()

    if args.list:
        for row in suite_summary():
            print(f"{row['suite']:<12} focus={row['focus']:<12} count={row['count']:>4}  {row['description']}")
        return 0

    try:
        jobs = discover_jobs(args.suites)
    except ValueError as exc:
        print(f"[regen-screenshots] {exc}", file=sys.stderr)
        return 2

    if not jobs:
        print("[regen-screenshots] no capture jobs discovered", file=sys.stderr)
        return 2

    out_dir = Path(args.out_dir) if args.out_dir else _default_out_dir()
    if not out_dir.is_absolute():
        out_dir = ROOT / out_dir
    out_dir.mkdir(parents=True, exist_ok=True)

    manifest = {
        "generated_at": time.strftime("%Y-%m-%dT%H:%M:%S"),
        "out_dir": str(out_dir.relative_to(ROOT)) if out_dir.is_relative_to(ROOT) else str(out_dir),
        "suites": args.suites or list(DEFAULT_SUITES),
        "jobs": [],
    }

    failures: list[str] = []
    for job in jobs:
        output = out_dir / job.output_rel
        output.parent.mkdir(parents=True, exist_ok=True)
        ok, message = _run_capture(job, output, args.godot, args.dry_run)
        entry = {
            "suite": job.suite,
            "focus": job.focus,
            "slug": job.slug,
            "output": str(output.relative_to(ROOT)) if output.is_relative_to(ROOT) else str(output),
            "extra_args": list(job.extra_args),
            "ok": ok,
        }
        if message:
            entry["error"] = message
        manifest["jobs"].append(entry)
        if not ok:
            failures.append(message)
            if args.fail_fast:
                break

    if not args.dry_run:
        index_path = out_dir / "index.json"
        index_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
        latest = ROOT / ".artifacts" / "screenshots" / "latest"
        latest.parent.mkdir(parents=True, exist_ok=True)
        if latest.exists() or latest.is_symlink():
            latest.unlink()
        latest.symlink_to(out_dir.name, target_is_directory=True)

    total = len(jobs)
    passed = sum(1 for job in manifest["jobs"] if job.get("ok"))
    print(f"[regen-screenshots] {passed}/{total} captures ok -> {out_dir}")
    if failures:
        for message in failures:
            print(f"[regen-screenshots] FAIL: {message}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
