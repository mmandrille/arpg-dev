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
# shellcheck source=quiet_helpers.sh
source "$ROOT/scripts/quiet_helpers.sh"

export BASE_URL="${BASE_URL:-http://localhost:8888}"
export DEV_TOKEN="${DEV_TOKEN:-local-dev-token}"
export DEBUG_TOKEN="${DEBUG_TOKEN:-local-debug-token}"
export ARPG_EMAIL="${ARPG_EMAIL:-client-smoke+$(date -u +%Y%m%d%H%M%S)-$$@example.test}"

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
  local gate_log gate_started elapsed
  gate_log="$(mktemp -t arpg-client-gate.XXXXXX.log)"

  echo "[client-smoke] running $label"
  gate_started=$SECONDS
  set +e
  "$GODOT" --headless --path "$CLIENT_DIR" --script "$script" >"$gate_log" 2>&1
  local status=$?
  set -e
  elapsed=$((SECONDS - gate_started))

  if [[ $status -ne 0 ]]; then
    echo "FAILED: $label (exit $status, elapsed=${elapsed}s)"
    show_log "$gate_log" "$label"
    rm -f "$gate_log"
    exit 1
  fi

  if ! grep -qF -- "$sentinel" "$gate_log"; then
    echo "FAILED: $label (missing sentinel: $sentinel, elapsed=${elapsed}s)"
    show_log "$gate_log" "$label"
    rm -f "$gate_log"
    exit 1
  fi

  if is_quiet_mode; then
    echo "OK: $label (${elapsed}s)"
  else
    cat "$gate_log"
    echo "[client-smoke] OK: $label elapsed=${elapsed}s"
  fi

  rm -f "$gate_log"
}

# Import resources once so headless --script runs cleanly.
echo "[client-smoke] running Godot asset import"
import_started=$SECONDS
"$GODOT" --headless --path "$CLIENT_DIR" --import >/dev/null 2>&1 || true
echo "OK: Godot asset import ($((SECONDS - import_started))s)"

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

# 2e. Co-op local/remote player handling test (server-independent; v33).
run_gate "GDScript co-op client unit test" "[gdtest] PASS: test_coop_client" res://tests/test_coop_client.gd

# 2f. Waypoint panel scroll/layout test (server-independent; v19).
run_gate "GDScript waypoint panel test" "[gdtest] PASS: test_waypoint_panel" res://tests/test_waypoint_panel.gd

# 2g. Sustained click hold state (server-independent; v27).
run_gate "GDScript sustained input test" "[gdtest] PASS: test_sustained_input" res://tests/test_sustained_input.gd

# 2h. Force-stand directional attack helpers (server-independent; v37).
run_gate "GDScript directional attack input test" "[gdtest] PASS: test_directional_attack_input" res://tests/test_directional_attack_input.gd

# 2i. Shop panel render and intent payload test (server-independent; v41).
run_gate "GDScript shop panel test" "[gdtest] PASS: test_shop_panel" res://tests/test_shop_panel.gd

# 2j. Stash panel render and intent payload test (server-independent; v50).
run_gate "GDScript stash panel test" "[gdtest] PASS: test_stash_panel" res://tests/test_stash_panel.gd

# 2k. Skill point panel and skill bar tests (server-independent; v44).
run_gate "GDScript skill rules loader test" "[gdtest] PASS: test_skill_rules_loader" res://tests/test_skill_rules_loader.gd
run_gate "GDScript skills panel test" "[gdtest] PASS: test_skills_panel" res://tests/test_skills_panel.gd
run_gate "GDScript skill bar test" "[gdtest] PASS: test_skill_bar" res://tests/test_skill_bar.gd
run_gate "GDScript status effects bar test" "[gdtest] PASS: test_status_effects_bar" res://tests/test_status_effects_bar.gd

# 2l. Boss health bar render/state test (server-independent; v53).
run_gate "GDScript boss health bar test" "[gdtest] PASS: test_boss_health_bar" res://tests/test_boss_health_bar.gd

# 2m. Delta and snapshot state-mutation unit tests (server-independent; v53).
run_gate "GDScript delta apply test" "[gdtest] PASS: test_delta_apply" res://tests/test_delta_apply.gd

if [[ "${CLIENT_UNIT_ONLY:-}" == "1" ]]; then
  echo "[client-unit] PASS"
  exit 0
fi

# 3. Headless slice smoke against the running server.
run_gate "headless slice smoke" "[smoke] PASS" res://scripts/smoke.gd
echo "[client-smoke] PASS"
