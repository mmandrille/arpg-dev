#!/usr/bin/env python3
"""Thin wrapper around the Go replay verifier (cmd/arpg-replay).

Replay verification re-simulates from Postgres-persisted seed + inputs, so it
lives in Go alongside the authoritative sim. This wrapper exists so the replay
step has a uniform Python entrypoint next to the bot; it simply shells out.

Usage:
    python -m tools.replay.run --session-id <id> [--json]
"""
from __future__ import annotations

import argparse
import os
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
SERVER_DIR = ROOT / "server"


def main() -> int:
    parser = argparse.ArgumentParser(description="Verify a recorded session by replay.")
    parser.add_argument("--session-id", required=True)
    parser.add_argument("--json", action="store_true")
    args = parser.parse_args()

    cmd = ["go", "run", "./cmd/arpg-replay", "--session-id", args.session_id]
    if args.json:
        cmd.append("--json")
    proc = subprocess.run(cmd, cwd=SERVER_DIR, env=os.environ.copy())
    return proc.returncode


if __name__ == "__main__":
    sys.exit(main())
