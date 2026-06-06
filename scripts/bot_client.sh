#!/usr/bin/env bash
# Godot client bot runner.
# Discovers client scenario JSON files, validates each, and launches one fresh
# Godot headless process per scenario. Requires a live DB + server (same as
# `make bot`). Fails hard if Godot is unavailable or any scenario exits non-zero
# or omits the expected [bot-client] PASS sentinel.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLIENT_DIR="$ROOT/client"
SCENARIOS_DIR="$ROOT/tools/bot/scenarios/client"

GODOT="${GODOT:-godot}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
DEV_TOKEN="${DEV_TOKEN:-local-dev-token}"
SCENARIO="${SCENARIO:-all}"

if ! command -v "$GODOT" >/dev/null 2>&1; then
  echo "[bot-client] FAIL: Godot runtime '$GODOT' not found on PATH." >&2
  echo "[bot-client] Install Godot $(cat "$ROOT/.godot-version" 2>/dev/null || echo '?') and set GODOT=/path/to/godot." >&2
  exit 1
fi

echo "[bot-client] Using Godot: $("$GODOT" --version 2>/dev/null | tail -1)"

# Collect scenario files to run.
declare -a SCENARIO_FILES=()
if [[ "$SCENARIO" == "all" ]]; then
  while IFS= read -r -d '' f; do
    SCENARIO_FILES+=("$f")
  done < <(find "$SCENARIOS_DIR" -maxdepth 1 -name '*.json' -print0 | sort -z)
elif [[ -f "$SCENARIO" ]]; then
  SCENARIO_FILES=("$SCENARIO")
else
  # Treat as a scenario id: match against file basenames.
  while IFS= read -r -d '' f; do
    bn="$(basename "$f" .json)"
    if [[ "$bn" == "$SCENARIO" || "$bn" == *"_$SCENARIO" || "$bn" == "$SCENARIO"* ]]; then
      SCENARIO_FILES+=("$f")
    fi
  done < <(find "$SCENARIOS_DIR" -maxdepth 1 -name '*.json' -print0 | sort -z)
fi

if [[ "${#SCENARIO_FILES[@]}" -eq 0 ]]; then
  echo "[bot-client] FAIL: no scenarios found matching '$SCENARIO' in $SCENARIOS_DIR" >&2
  exit 1
fi

# Validate runner + world_id + client_steps for each file before launching.
validate_scenario_file() {
  local path="$1"
  local runner world_id steps_len
  runner="$(python3 -c "import json,sys; d=json.load(open('$path')); print(d.get('runner',''))" 2>/dev/null || echo "")"
  if [[ "$runner" != "godot_client" ]]; then
    echo "[bot-client] FAIL: $path: runner must be 'godot_client', got '$runner'" >&2
    return 1
  fi
  world_id="$(python3 -c "import json,sys; d=json.load(open('$path')); print(d.get('world_id',''))" 2>/dev/null || echo "")"
  if [[ -z "$world_id" ]]; then
    echo "[bot-client] FAIL: $path: world_id is missing or empty" >&2
    return 1
  fi
  steps_len="$(python3 -c "import json,sys; d=json.load(open('$path')); print(len(d.get('client_steps',[])))" 2>/dev/null || echo "0")"
  if [[ "$steps_len" -eq 0 ]]; then
    echo "[bot-client] FAIL: $path: client_steps is missing or empty" >&2
    return 1
  fi
}

for f in "${SCENARIO_FILES[@]}"; do
  validate_scenario_file "$f"
done

# Import once before the scenario loop so headless --path runs cleanly.
"$GODOT" --headless --path "$CLIENT_DIR" --import >/dev/null 2>&1 || true

PASS_COUNT=0
FAIL_COUNT=0

run_scenario() {
  local scenario_path="$1"
  local scenario_id world_id out exit_code
  scenario_id="$(python3 -c "import json; d=json.load(open('$scenario_path')); print(d.get('id','unknown'))")"
  world_id="$(python3 -c "import json; d=json.load(open('$scenario_path')); print(d.get('world_id',''))")"

  echo "[bot-client] --- running scenario: $scenario_id (world=$world_id)"
  exit_code=0
  out="$(ARPG_BOT_CLIENT=1 \
    ARPG_BOT_SCENARIO="$scenario_path" \
    ARPG_WORLD_ID="$world_id" \
    ARPG_BASE_URL="$BASE_URL" \
    ARPG_DEV_TOKEN="$DEV_TOKEN" \
    "$GODOT" --headless --resolution 1280x720 --path "$CLIENT_DIR" 2>&1)" \
    || exit_code=$?

  printf '%s\n' "$out"

  if [[ $exit_code -ne 0 ]]; then
    echo "[bot-client] FAIL $scenario_id -- exited with code $exit_code" >&2
    return 1
  fi

  if ! grep -qF "[bot-client] PASS $scenario_id" <<<"$out"; then
    echo "[bot-client] FAIL $scenario_id -- PASS sentinel not found in output" >&2
    return 1
  fi

  echo "[bot-client] OK $scenario_id"
}

for f in "${SCENARIO_FILES[@]}"; do
  if run_scenario "$f"; then
    PASS_COUNT=$((PASS_COUNT + 1))
  else
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi
done

echo "[bot-client] Results: $PASS_COUNT passed, $FAIL_COUNT failed"

if [[ $FAIL_COUNT -gt 0 ]]; then
  exit 1
fi
exit 0
