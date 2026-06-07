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

_ts() {
  date -u +"%H:%M:%S"
}

BASE_URL="${BASE_URL:-http://localhost:8080}"
DEV_TOKEN="${DEV_TOKEN:-local-dev-token}"
EMAIL="${ARPG_EMAIL:-client-bot@example.test}"
EMAIL_RUN_ID="${ARPG_BOT_CLIENT_RUN_ID:-$(date -u +%Y%m%d%H%M%S)-$$}"
SCENARIO="${SCENARIO:-all}"
# Set HEADLESS=1 for CI. Default is 0 (windowed) so you can watch the bot act.
HEADLESS="${HEADLESS:-0}"
# When windowed, pause this many seconds between steps so the action is visible.
BOT_STEP_DELAY="${BOT_STEP_DELAY:-$([ "$HEADLESS" == "1" ] && echo 0.0 || echo 0.25)}"

if ! command -v "$GODOT" >/dev/null 2>&1; then
  echo "[bot-client] FAIL: Godot runtime '$GODOT' not found on PATH." >&2
  echo "[bot-client] Install Godot $(cat "$ROOT/.godot-version" 2>/dev/null || echo '?') and set GODOT=/path/to/godot." >&2
  exit 1
fi

echo "[bot-client $(_ts)] Using Godot: $("$GODOT" --version 2>/dev/null | tail -1)"
echo "[bot-client $(_ts)] SCENARIO=$SCENARIO HEADLESS=$HEADLESS BOT_STEP_DELAY=$BOT_STEP_DELAY"

scenario_email() {
  local base_email="$1"
  local run_id="$2"
  local scenario_id="$3"
  if [[ "$base_email" != *"@"* ]]; then
    printf '%s\n' "$base_email"
    return
  fi
  local local_part="${base_email%@*}"
  local domain="${base_email#*@}"
  local safe_id
  safe_id="$(printf '%s-%s' "$run_id" "$scenario_id" | tr -c '[:alnum:]' '-')"
  printf '%s+%s@%s\n' "$local_part" "$safe_id" "$domain"
}

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

READY_URL="${BASE_URL%/}/readyz"
if ! server_error="$(curl --max-time 2 -fsS "$READY_URL" 2>&1)"; then
  echo "[bot-client] FAIL: server is not reachable at $READY_URL." >&2
  echo "[bot-client] Start local dependencies and the server first:" >&2
  echo "[bot-client]   make db-up" >&2
  echo "[bot-client]   make server" >&2
  echo "[bot-client] Or use make bot-visual, which starts its own temporary server." >&2
  echo "[bot-client] Details: $server_error" >&2
  exit 1
fi

# Import once before the scenario loop so headless --path runs cleanly.
echo "[bot-client $(_ts)] Godot asset import starting (can take 30-90s on cold cache)..."
import_started=$SECONDS
if "$GODOT" --headless --path "$CLIENT_DIR" --import; then
  echo "[bot-client $(_ts)] Godot asset import done elapsed=$((SECONDS - import_started))s"
else
  echo "[bot-client $(_ts)] Godot asset import finished with warnings elapsed=$((SECONDS - import_started))s" >&2
fi

PASS_COUNT=0
FAIL_COUNT=0

run_scenario() {
  local scenario_path="$1"
  local scenario_id world_id exit_code started_ts tmpfile
  scenario_id="$(python3 -c "import json; d=json.load(open('$scenario_path')); print(d.get('id','unknown'))")"
  world_id="$(python3 -c "import json; d=json.load(open('$scenario_path')); print(d.get('world_id',''))")"
  started_ts="$SECONDS"
  tmpfile="$(mktemp)"

  echo "[bot-client $(_ts)] --- running scenario: $scenario_id (world=$world_id file=$(basename "$scenario_path"))"
  exit_code=0
  local godot_flags="--resolution 1280x720"
  local email
  email="$(scenario_email "$EMAIL" "$EMAIL_RUN_ID" "$scenario_id")"
  [[ "$HEADLESS" == "1" ]] && godot_flags="--headless $godot_flags"
  echo "[bot-client $(_ts)] launching Godot for $scenario_id..."
  local launch_started=$SECONDS
  ARPG_BOT_CLIENT=1 \
    ARPG_BOT_SCENARIO="$scenario_path" \
    ARPG_WORLD_ID="$world_id" \
    ARPG_BASE_URL="$BASE_URL" \
    ARPG_DEV_TOKEN="$DEV_TOKEN" \
    ARPG_EMAIL="$email" \
    ARPG_BOT_STEP_DELAY="$BOT_STEP_DELAY" \
    "$GODOT" $godot_flags --path "$CLIENT_DIR" 2>&1 | tee "$tmpfile"
  exit_code=${PIPESTATUS[0]}

  echo "[bot-client $(_ts)] Godot process exited code=$exit_code launch_elapsed=$((SECONDS - launch_started))s"

  if [[ $exit_code -ne 0 ]]; then
    rm -f "$tmpfile"
    echo "[bot-client] FAIL $scenario_id -- exited with code $exit_code" >&2
    return 1
  fi

  if ! grep -qF "[bot-client] PASS $scenario_id" "$tmpfile"; then
    rm -f "$tmpfile"
    echo "[bot-client] FAIL $scenario_id -- PASS sentinel not found in output" >&2
    return 1
  fi

  rm -f "$tmpfile"
  local elapsed=$((SECONDS - started_ts))
  echo "[bot-client $(_ts)] OK $scenario_id elapsed=${elapsed}s"
}

for f in "${SCENARIO_FILES[@]}"; do
  if run_scenario "$f"; then
    PASS_COUNT=$((PASS_COUNT + 1))
  else
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi
done

echo "[bot-client $(_ts)] Results: $PASS_COUNT passed, $FAIL_COUNT failed"

if [[ $FAIL_COUNT -gt 0 ]]; then
  exit 1
fi
exit 0
