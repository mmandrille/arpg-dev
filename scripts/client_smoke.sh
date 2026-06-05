#!/usr/bin/env bash
# Godot headless client smoke test.
# Loads the Godot project, runs the headless smoke scene which logs in, creates
# a session, completes move/attack/pickup/equip, and verifies client state via
# the debug API. Skips gracefully (exit 0 with a notice) when the pinned Godot
# runtime is not installed, since CI runners may not yet provision it.
set -euo pipefail

GODOT="${GODOT:-godot}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLIENT_DIR="$ROOT/client"

export BASE_URL="${BASE_URL:-http://localhost:8080}"
export DEV_TOKEN="${DEV_TOKEN:-local-dev-token}"
export DEBUG_TOKEN="${DEBUG_TOKEN:-local-debug-token}"

if ! command -v "$GODOT" >/dev/null 2>&1; then
  echo "[client-smoke] SKIP: Godot runtime '$GODOT' not found on PATH."
  echo "[client-smoke] Install Godot $(cat "$ROOT/.godot-version") and re-run, or set GODOT=/path/to/godot."
  exit 0
fi

echo "[client-smoke] Using Godot: $("$GODOT" --version 2>/dev/null | tail -1)"

# Import resources once so headless --script runs cleanly.
"$GODOT" --headless --path "$CLIENT_DIR" --import >/dev/null 2>&1 || true

# 1. GDScript golden-fixture test (server-independent; ADR D6 / acceptance #7).
echo "[client-smoke] running GDScript golden test"
"$GODOT" --headless --path "$CLIENT_DIR" --script res://tests/test_golden.gd

# 2. Headless slice smoke against the running server.
echo "[client-smoke] running headless slice smoke"
"$GODOT" --headless --path "$CLIENT_DIR" --script res://scripts/smoke.gd
echo "[client-smoke] PASS"
