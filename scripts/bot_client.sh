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
# shellcheck source=quiet_helpers.sh
source "$ROOT/scripts/quiet_helpers.sh"

GODOT="${GODOT:-godot}"
PYTHON="${PYTHON:-$ROOT/.venv/bin/python}"
if [[ ! -x "$PYTHON" ]]; then
  PYTHON="python3"
fi

_ts() {
  date -u +"%H:%M:%S"
}

BASE_URL="${BASE_URL:-http://localhost:8888}"
DEV_TOKEN="${DEV_TOKEN:-local-dev-token}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-${DEBUG_TOKEN:-local-debug-token}}"
GAMEPLAY_DEBUG="${ARPG_GAMEPLAY_DEBUG:-true}"
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
    if [[ "$bn" == "$SCENARIO" || "$bn.json" == "$SCENARIO" || "$bn" == *"_$SCENARIO" || "$bn" == "$SCENARIO"* ]]; then
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

VALIDATION_FAIL_COUNT=0
for f in "${SCENARIO_FILES[@]}"; do
  if ! validate_scenario_file "$f"; then
    VALIDATION_FAIL_COUNT=$((VALIDATION_FAIL_COUNT + 1))
  fi
done
if [[ "$VALIDATION_FAIL_COUNT" -gt 0 ]]; then
  echo "[bot-client] FAIL: $VALIDATION_FAIL_COUNT scenario file(s) failed validation" >&2
  exit 1
fi

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
if is_quiet_mode; then
  "$RUN_QUIET" --label "Godot asset import" -- bash -c '"$1" --headless --path "$2" --import || true' _ "$GODOT" "$CLIENT_DIR"
else
  echo "[bot-client $(_ts)] Godot asset import starting (can take 30-90s on cold cache)..."
  import_started=$SECONDS
  if "$GODOT" --headless --path "$CLIENT_DIR" --import; then
    echo "[bot-client $(_ts)] Godot asset import done elapsed=$((SECONDS - import_started))s"
  else
    echo "[bot-client $(_ts)] Godot asset import finished with warnings elapsed=$((SECONDS - import_started))s" >&2
  fi
fi

PASS_COUNT=0
FAIL_COUNT=0
PREFLIGHT_PIDS=()

cleanup_preflights() {
  local pid
  for pid in "${PREFLIGHT_PIDS[@]:-}"; do
    if kill -0 "$pid" >/dev/null 2>&1; then
      kill "$pid" >/dev/null 2>&1 || true
      wait "$pid" >/dev/null 2>&1 || true
    fi
  done
  PREFLIGHT_PIDS=()
}

trap cleanup_preflights EXIT
trap 'cleanup_preflights; exit 130' INT
trap 'cleanup_preflights; exit 143' TERM

json_field() {
  local path="$1"
  local expr="$2"
  python3 -c "import json,sys; d=json.load(open(sys.argv[1])); print($expr)" "$path"
}

metadata_field() {
  local path="$1"
  local key="$2"
  python3 -c "import json,sys; d=json.load(open(sys.argv[1])); print(d.get(sys.argv[2], ''))" "$path" "$key"
}

cleanup_account_email() {
  local email="$1"
  if [[ -z "$email" ]]; then
    return 0
  fi
  "$PYTHON" "$ROOT/tools/bot/cleanup_account.py" \
    --base-url "$BASE_URL" \
    --dev-token "$DEV_TOKEN" \
    --email "$email" >/dev/null
}

start_preflight() {
  local scenario_path="$1"
  local scenario_id="$2"
  local world_id="$3"
  local seed="$4"
  local metadata_file="$5"
  local log_file="$6"
  local preflight_type host_email pid
  preflight_type="$(json_field "$scenario_path" "d.get('preflight', {}).get('type', '')")"
  if [[ -z "$preflight_type" ]]; then
    return 0
  fi
  if [[ "$preflight_type" == "market_listing" ]]; then
    host_email="$(scenario_email "$EMAIL" "$EMAIL_RUN_ID" "${scenario_id}-seller")"
    if [[ "$(json_field "$scenario_path" "d.get('preflight', {}).get('seller_is_client', False)")" == "True" ]]; then
      host_email="$(scenario_email "$EMAIL" "$EMAIL_RUN_ID" "$scenario_id")"
    fi
    echo "[bot-client $(_ts)] starting market preflight for $scenario_id email=$host_email"
    "$PYTHON" "$ROOT/tools/bot/client_market_preflight.py" \
      --base-url "$BASE_URL" \
      --dev-token "$DEV_TOKEN" \
      --debug-token "${ARPG_DEBUG_TOKEN:-local-debug-token}" \
      --world-id "$world_id" \
      --seed "$seed" \
      --email "$host_email" \
      --character-name "Market Seller" \
      --metadata-file "$metadata_file" \
      --item-def-id "$(json_field "$scenario_path" "d.get('preflight', {}).get('item_def_id', 'cave_mail')")" \
      --price-gold "$(json_field "$scenario_path" "d.get('preflight', {}).get('price_gold', 37)")" \
      --offer-email "$(scenario_email "$EMAIL" "$EMAIL_RUN_ID" "${scenario_id}-bidder")" \
      --offer-item-def-id "$(json_field "$scenario_path" "d.get('preflight', {}).get('offer_item_def_id', '')")" \
      >"$log_file" 2>&1
    echo "[bot-client $(_ts)] market preflight ready listing=$(metadata_field "$metadata_file" listing_id)"
    return 0
  fi
  if [[ "$preflight_type" != "listed_coop_host" ]]; then
    echo "[bot-client] FAIL: $scenario_path: unsupported preflight type '$preflight_type'" >&2
    return 1
  fi
  host_email="$(scenario_email "$EMAIL" "$EMAIL_RUN_ID" "${scenario_id}-host")"
  echo "[bot-client $(_ts)] starting preflight host for $scenario_id email=$host_email"
  "$PYTHON" "$ROOT/tools/bot/client_join_preflight.py" \
    --base-url "$BASE_URL" \
    --dev-token "$DEV_TOKEN" \
    --world-id "$world_id" \
    --seed "$seed" \
    --email "$host_email" \
    --character-name "Join Host" \
    --metadata-file "$metadata_file" \
    >"$log_file" 2>&1 &
  pid=$!
  PREFLIGHT_PIDS+=("$pid")

  local deadline=$((SECONDS + 20))
  while (( SECONDS < deadline )); do
    if ! kill -0 "$pid" >/dev/null 2>&1; then
      echo "[bot-client] FAIL: preflight host exited before ready for $scenario_id" >&2
      show_log "$log_file" "$scenario_id preflight"
      return 1
    fi
    if [[ -s "$metadata_file" ]] && python3 -c "import json,sys; d=json.load(open(sys.argv[1])); raise SystemExit(0 if d.get('ready') else 1)" "$metadata_file" >/dev/null 2>&1; then
      echo "[bot-client $(_ts)] preflight ready session=$(metadata_field "$metadata_file" session_id)"
      return 0
    fi
    sleep 0.1
  done

  echo "[bot-client] FAIL: preflight host timed out for $scenario_id" >&2
  show_log "$log_file" "$scenario_id preflight"
  return 1
}

run_scenario() {
  local scenario_path="$1"
  local scenario_id world_id seed debug_gold exit_code started_ts tmpfile preflight_metadata preflight_log expected_join_session
  scenario_id="$(python3 -c "import json; d=json.load(open('$scenario_path')); print(d.get('id','unknown'))")"
  world_id="$(python3 -c "import json; d=json.load(open('$scenario_path')); print(d.get('world_id',''))")"
  seed="$(python3 -c "import json; d=json.load(open('$scenario_path')); print(d.get('seed',''))")"
  debug_gold="$(python3 -c "import json; d=json.load(open('$scenario_path')); print(d.get('debug_progression', {}).get('gold', ''))")"
  started_ts="$(python3 -c 'import time; print(time.monotonic())')"
  tmpfile="$(mktemp)"
  preflight_metadata="$(mktemp)"
  preflight_log="$(mktemp)"

  if is_quiet_mode && [[ "$HEADLESS" == "1" ]]; then
    echo "RUNNING: client bot scenario $scenario_id"
  else
    echo "[bot-client $(_ts)] running scenario: $scenario_id (world=$world_id file=$(basename "$scenario_path"))"
  fi
  exit_code=0
  local godot_flags="--resolution 1280x720"
  local email
  email="$(scenario_email "$EMAIL" "$EMAIL_RUN_ID" "$scenario_id")"
  expected_join_session=""
  if ! start_preflight "$scenario_path" "$scenario_id" "$world_id" "$seed" "$preflight_metadata" "$preflight_log"; then
    rm -f "$tmpfile" "$preflight_metadata" "$preflight_log"
    cleanup_preflights
    return 1
  fi
  if [[ -s "$preflight_metadata" ]]; then
    expected_join_session="$(metadata_field "$preflight_metadata" session_id)"
  fi
  [[ "$HEADLESS" == "1" ]] && godot_flags="--headless $godot_flags"
  if is_quiet_mode && [[ "$HEADLESS" == "1" ]]; then
    ARPG_BOT_CLIENT=1 \
      ARPG_BOT_SCENARIO="$scenario_path" \
      ARPG_WORLD_ID="$world_id" \
      ARPG_SEED="$seed" \
      ARPG_BASE_URL="$BASE_URL" \
      ARPG_DEV_TOKEN="$DEV_TOKEN" \
      ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
      ARPG_GAMEPLAY_DEBUG="$GAMEPLAY_DEBUG" \
      ARPG_BOT_DEBUG_GOLD="$debug_gold" \
      ARPG_EMAIL="$email" \
      ARPG_EXPECTED_JOIN_SESSION_ID="$expected_join_session" \
      ARPG_BOT_STEP_DELAY="$BOT_STEP_DELAY" \
      "$GODOT" $godot_flags --path "$CLIENT_DIR" >"$tmpfile" 2>&1
    exit_code=$?
  else
    echo "[bot-client $(_ts)] launching Godot for $scenario_id..."
    local launch_started=$SECONDS
    ARPG_BOT_CLIENT=1 \
      ARPG_BOT_SCENARIO="$scenario_path" \
      ARPG_WORLD_ID="$world_id" \
      ARPG_SEED="$seed" \
      ARPG_BASE_URL="$BASE_URL" \
      ARPG_DEV_TOKEN="$DEV_TOKEN" \
      ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
      ARPG_GAMEPLAY_DEBUG="$GAMEPLAY_DEBUG" \
      ARPG_BOT_DEBUG_GOLD="$debug_gold" \
      ARPG_EMAIL="$email" \
      ARPG_EXPECTED_JOIN_SESSION_ID="$expected_join_session" \
      ARPG_BOT_STEP_DELAY="$BOT_STEP_DELAY" \
      "$GODOT" $godot_flags --path "$CLIENT_DIR" 2>&1 | tee "$tmpfile"
    exit_code=${PIPESTATUS[0]}
    echo "[bot-client $(_ts)] Godot process exited code=$exit_code launch_elapsed=$((SECONDS - launch_started))s"
  fi

  if [[ $exit_code -ne 0 ]]; then
    echo "[bot-client] FAIL $scenario_id -- exited with code $exit_code" >&2
    if is_quiet_mode && [[ "$HEADLESS" == "1" ]]; then
      show_log "$tmpfile" "$scenario_id"
    fi
    rm -f "$tmpfile" "$preflight_metadata" "$preflight_log"
    cleanup_preflights
    return 1
  fi

  if ! grep -qF "[bot-client] PASS $scenario_id" "$tmpfile"; then
    echo "[bot-client] FAIL $scenario_id -- PASS sentinel not found in output" >&2
    if is_quiet_mode && [[ "$HEADLESS" == "1" ]]; then
      show_log "$tmpfile" "$scenario_id"
    fi
    rm -f "$tmpfile" "$preflight_metadata" "$preflight_log"
    cleanup_preflights
    return 1
  fi

  cleanup_preflights
  cleanup_account_email "$email"
  if [[ -s "$preflight_metadata" ]]; then
    cleanup_account_email "$(metadata_field "$preflight_metadata" host_email)"
  fi
  rm -f "$tmpfile" "$preflight_metadata" "$preflight_log"
  local elapsed
  elapsed="$(python3 -c 'import sys,time; print(f"{time.monotonic() - float(sys.argv[1]):.2f}s")' "$started_ts")"
  if is_quiet_mode && [[ "$HEADLESS" == "1" ]]; then
    echo "OK: client bot scenario $scenario_id (elapsed=$elapsed)"
  else
    echo "[bot-client $(_ts)] OK $scenario_id elapsed=${elapsed}"
  fi
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
