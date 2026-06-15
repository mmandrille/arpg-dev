#!/usr/bin/env bash
# Local CI aggregation for the first playable vertical slice.
# Runs all steps regardless of failures, then reports a summary.
# Quiet by default — VERBOSE=1 (or V=1) for full output.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
# shellcheck source=quiet_helpers.sh
source "$ROOT/scripts/quiet_helpers.sh"

DATABASE_URL="${ARPG_DATABASE_URL:-postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable}"
ADDR="${ARPG_ADDR:-:8888}"
BASE_URL="${BASE_URL:-http://localhost:8888}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-local-dev-token}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-local-debug-token}"
GAMEPLAY_DEBUG="${ARPG_GAMEPLAY_DEBUG:-true}"

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-ci-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT
CI_STARTED_AT="$(date +%s)"

FAILED_STEPS=()
STEP_PRINTED=0
STEP_STARTED_AT=0
STEP_LABEL=""

format_duration() {
  local total="$1"
  local hours=$((total / 3600))
  local minutes=$(((total % 3600) / 60))
  local seconds=$((total % 60))
  if [[ "$hours" -gt 0 ]]; then
    printf '%dh%02dm%02ds' "$hours" "$minutes" "$seconds"
  elif [[ "$minutes" -gt 0 ]]; then
    printf '%dm%02ds' "$minutes" "$seconds"
  else
    printf '%ds' "$seconds"
  fi
}

begin_step() {
  local title="$1"
  if [[ "$STEP_PRINTED" -eq 1 ]]; then
    echo
  fi
  echo "$title"
  STEP_LABEL="${title#== }"
  STEP_LABEL="${STEP_LABEL% ==}"
  STEP_STARTED_AT="$(date +%s)"
  STEP_PRINTED=1
}

finish_step() {
  local elapsed=$(( $(date +%s) - STEP_STARTED_AT ))
  local total_elapsed=$(( $(date +%s) - CI_STARTED_AT ))
  echo "completed in $(format_duration "$elapsed") (total $(format_duration "$total_elapsed"))"
  STEP_STARTED_AT=0
  STEP_LABEL=""
}

finish_step_failed() {
  local elapsed=$(( $(date +%s) - STEP_STARTED_AT ))
  local total_elapsed=$(( $(date +%s) - CI_STARTED_AT ))
  echo "FAILED after $(format_duration "$elapsed") (total $(format_duration "$total_elapsed"))"
  STEP_STARTED_AT=0
  STEP_LABEL=""
}

# Run a step: print header, run command, track failure. Never exits.
ci_step() {
  local label="$1"
  shift
  begin_step "$label"
  set +e
  "$@"
  local status=$?
  set -e
  if [[ $status -ne 0 ]]; then
    finish_step_failed
    FAILED_STEPS+=("${label}")
    return 1
  fi
  finish_step
  return 0
}

stream_bot_progress() {
  local log_path="$1"
  tee "$log_path" | awk '
    /\] scenario done / {
      line = $0
      sub(/^.*\] scenario done /, "", line)
      scenario = line
      sub(/ .*/, "", scenario)
      elapsed = ""
      if (match(line, /elapsed=[^ ]+/)) {
        elapsed = substr(line, RSTART, RLENGTH)
      }
      if (elapsed != "") {
        printf "OK: protocol bot scenario %s (%s)\n", scenario, elapsed
      } else {
        printf "OK: protocol bot scenario %s\n", scenario
      }
      fflush()
    }
  '
}

# ── Maintainability ratchets (were a Makefile prereq) ────────────────────────

ci_step "== 1/11 file-size ratchet ==" \
  "$RUN_QUIET" --label "file-size-ratchet" -- ./scripts/check-file-size-ratchet.sh

ci_step "== 2/11 extraction coupling ratchet ==" \
  "$RUN_QUIET" --label "extraction-coupling-ratchet" -- \
    python3 ./scripts/check-extraction-coupling-ratchet.py

# ── Independent checks ────────────────────────────────────────────────────────

ci_step "== 3/11 shared schema validation ==" \
  "$RUN_QUIET" --label validate-shared -- make validate-shared

ci_step "== 4/11 asset manifest + GLB validation ==" \
  "$RUN_QUIET" --label validate-assets -- make validate-assets

ci_step "== 5/11 determinism lint ==" \
  "$RUN_QUIET" --label "determinism-lint" -- make lint-determinism

ci_step "== 6/11 Go tests ==" \
  "$RUN_QUIET" --label "go test ./..." -- bash -c 'cd server && go test ./...'

ci_step "== 7/11 Python unit checks ==" \
  bash -c "make tools >/dev/null && \"$RUN_QUIET\" --label 'pytest tools' -- \"$ROOT/.venv/bin/python\" -m pytest -q tools"

# ── Server-dependent steps ────────────────────────────────────────────────────

start_server() {
  begin_step "== 8/11 start Postgres + server =="
  set +e

  make db-up
  local db_status=$?
  if [[ $db_status -ne 0 ]]; then
    echo "FAILED: make db-up"
    finish_step_failed
    set -e
    FAILED_STEPS+=("== 8/11 start Postgres + server ==")
    return 1
  fi

  SERVER_BIN="$(mktemp -t arpg-ci-server.XXXXXX)"
  "$RUN_QUIET" --label "go build arpg-server" -- bash -c \
    "cd server && go build -o \"$SERVER_BIN\" ./cmd/arpg-server"
  local build_status=$?
  if [[ $build_status -ne 0 ]]; then
    finish_step_failed
    set -e
    FAILED_STEPS+=("== 8/11 start Postgres + server ==")
    return 1
  fi

  ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
    ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
    ARPG_GAMEPLAY_DEBUG="$GAMEPLAY_DEBUG" \
    ARPG_RULES_DIR="$ROOT/shared/rules" \
    "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
  SERVER_PID=$!
  echo "server pid=$SERVER_PID (log: $SERVER_LOG); waiting for readiness..."
  for i in $(seq 1 60); do
    if curl -fsS "$BASE_URL/readyz" >/dev/null 2>&1; then break; fi
    sleep 1
  done

  if ! curl -fsS "$BASE_URL/readyz" >/dev/null 2>&1; then
    echo "server failed readiness check; log:"
    show_log "$SERVER_LOG" "server"
    finish_step_failed
    set -e
    FAILED_STEPS+=("== 8/11 start Postgres + server ==")
    return 1
  fi

  set -e
  finish_step
  return 0
}

SERVER_AVAILABLE=0
if start_server; then
  SERVER_AVAILABLE=1
fi

if [[ "$SERVER_AVAILABLE" -eq 1 ]]; then
  # Step 9: protocol bot + replay
  begin_step "== 9/11 protocol bot + replay =="
  BOT_LOG="$(mktemp -t arpg-ci-bot.XXXXXX.log)"
  set +e
  SESSION_ID="$("$ROOT/.venv/bin/python" -m tools.bot.run \
    --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --debug-token "$DEBUG_TOKEN" \
    --print-session-id 2> >(stream_bot_progress "$BOT_LOG" >&2))"
  bot_status=$?
  set -e
  if [[ "$bot_status" -ne 0 ]]; then
    echo "FAILED: protocol bot"
    show_log "$BOT_LOG" "protocol bot"
    rm -f "$BOT_LOG"
    finish_step_failed
    FAILED_STEPS+=("== 9/11 protocol bot + replay ==")
  else
    if [[ "${ARPG_VERBOSE:-0}" == "1" ]]; then
      cat "$BOT_LOG"
    else
      echo "OK: protocol bot"
    fi
    rm -f "$BOT_LOG"
    echo "bot completed session: $SESSION_ID"

    set +e
    "$RUN_QUIET" --label "arpg-replay" -- bash -c \
      "cd server && ARPG_DATABASE_URL=\"$DATABASE_URL\" ARPG_GAMEPLAY_DEBUG=\"$GAMEPLAY_DEBUG\" \
       go run ./cmd/arpg-replay --session-id \"$SESSION_ID\""
    replay_status=$?
    set -e
    if [[ $replay_status -ne 0 ]]; then
      finish_step_failed
      FAILED_STEPS+=("== 9/11 protocol bot + replay ==")
    else
      finish_step
    fi
  fi

  # Steps 10-11: Godot
  ci_step "== 10/11 Godot client bot scenarios ==" \
    env GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" \
      SCENARIO=all HEADLESS=1 ./scripts/bot_client.sh

  ci_step "== 11/11 Godot headless smoke (optional) ==" \
    env GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" \
      DEBUG_TOKEN="$DEBUG_TOKEN" ./scripts/client_smoke.sh
else
  echo
  echo "SKIPPED: steps 9-11 require a running server (step 8 failed)"
  FAILED_STEPS+=("== 9/11 protocol bot + replay == (skipped: server failed)")
  FAILED_STEPS+=("== 10/11 Godot client bot scenarios == (skipped: server failed)")
  FAILED_STEPS+=("== 11/11 Godot headless smoke == (skipped: server failed)")
fi

# ── Summary ───────────────────────────────────────────────────────────────────

total_elapsed=$(( $(date +%s) - CI_STARTED_AT ))
echo

if [[ ${#FAILED_STEPS[@]} -gt 0 ]]; then
  echo "CI FAILED in $(format_duration $total_elapsed) — ${#FAILED_STEPS[@]} step(s) failed:"
  for step in "${FAILED_STEPS[@]}"; do
    echo "  ✗  ${step#== }"
  done
  echo "(scroll up for each step's output)"
  exit 1
fi

echo "CI OK in $(format_duration $total_elapsed)"
