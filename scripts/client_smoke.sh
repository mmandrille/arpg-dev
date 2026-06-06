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

# Godot exits 0 even on a GDScript PARSE/load error when run via --script, so a
# bare exit-code check could pass silently on a broken gate. run_gate captures
# each gate's combined output and asserts its expected success sentinel appears,
# failing nonzero (with a clear message) if it does not.
run_gate() {
  local label="$1" sentinel="$2" script="$3"
  local out
  echo "[client-smoke] running $label"
  if ! out="$("$GODOT" --headless --path "$CLIENT_DIR" --script "$script" 2>&1)"; then
    printf '%s\n' "$out"
    echo "[client-smoke] FAIL: $label exited nonzero" >&2
    exit 1
  fi
  printf '%s\n' "$out"
  if ! grep -qF -- "$sentinel" <<<"$out"; then
    echo "[client-smoke] FAIL: $label did not emit expected sentinel: $sentinel" >&2
    exit 1
  fi
}

# Import resources once so headless --script runs cleanly.
"$GODOT" --headless --path "$CLIENT_DIR" --import >/dev/null 2>&1 || true

# 1. GDScript golden-fixture test (server-independent; ADR D6 / acceptance #7).
run_gate "GDScript golden test" "[gdtest] PASS" res://tests/test_golden.gd

# 2. Item visual resolution test (server-independent; acceptance #14).
run_gate "GDScript item visual resolution test" "[gdtest] PASS" res://tests/test_item_visuals.gd

# 2b. Rig gate: both GLBs import as skinned Skeleton3D (spec §10 fail-fast).
run_gate "GDScript rig gate" "[rig-gate] PASS" res://tools/inspect_rig.gd

# 2c. Animation controller + rigged scene test (server-independent).
run_gate "GDScript animation test" "[gdtest] PASS: animation controller + scenes" res://tests/test_animation.gd

# 2d. Bot scenario runner unit tests (server-independent).
run_gate "GDScript client bot unit test" "[gdtest] PASS: test_client_bot" res://tests/test_client_bot.gd

if [[ "${CLIENT_UNIT_ONLY:-}" == "1" ]]; then
  echo "[client-unit] PASS"
  exit 0
fi

# 3. Headless slice smoke against the running server.
run_gate "headless slice smoke" "[smoke] PASS" res://scripts/smoke.gd
echo "[client-smoke] PASS"
