#!/usr/bin/env bash
# Local CI aggregation for the first playable vertical slice.
# Runs shared schema validation, Go tests, Python unit checks, the end-to-end
# bot + replay flow against a throwaway Postgres + server, and the Godot
# headless smoke when the runtime is available.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
# shellcheck source=quiet_helpers.sh
source "$ROOT/scripts/quiet_helpers.sh"

DATABASE_URL="${ARPG_DATABASE_URL:-postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable}"
ADDR="${ARPG_ADDR:-:8888}"
BASE_URL="${BASE_URL:-http://localhost:8888}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-local-dev-token}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-local-debug-token}"

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-ci-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT
CI_STARTED_AT="$(date +%s)"

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
  if [[ "$STEP_STARTED_AT" -gt 0 && -n "$STEP_LABEL" ]]; then
    local elapsed=$(( $(date +%s) - STEP_STARTED_AT ))
    local total_elapsed=$(( $(date +%s) - CI_STARTED_AT ))
    echo "failed after $(format_duration "$elapsed") (total $(format_duration "$total_elapsed"))" >&2
    STEP_STARTED_AT=0
    STEP_LABEL=""
  fi
}

report_step_failure() {
  local status=$?
  finish_step_failed
  exit "$status"
}

trap report_step_failure ERR

begin_step "== 1/9 shared schema validation =="
"$RUN_QUIET" --label validate-shared -- make validate-shared
finish_step

begin_step "== 2/9 asset manifest + GLB validation =="
"$RUN_QUIET" --label validate-assets -- make validate-assets
finish_step

begin_step "== 3/9 determinism lint =="
"$RUN_QUIET" --label "determinism-lint" -- make lint-determinism
finish_step

begin_step "== 4/9 Go tests =="
"$RUN_QUIET" --label "go test ./..." -- bash -c 'cd server && go test ./...'
finish_step

begin_step "== 5/9 Python unit checks =="
make tools >/dev/null
"$RUN_QUIET" --label "pytest tools" -- "$ROOT/.venv/bin/python" -m pytest -q tools
finish_step

begin_step "== 6/9 start Postgres + server =="
make db-up
# Build a binary and run it directly (not via `go run`, whose child binary
# would survive the cleanup kill and, if stdout were piped, hold the pipe open).
SERVER_BIN="$(mktemp -t arpg-ci-server.XXXXXX)"
"$RUN_QUIET" --label "go build arpg-server" -- bash -c "cd server && go build -o \"$SERVER_BIN\" ./cmd/arpg-server"
ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
  ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_RULES_DIR="$ROOT/shared/rules" \
  "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!
echo "server pid=$SERVER_PID (log: $SERVER_LOG); waiting for readiness..."
for i in $(seq 1 60); do
  if curl -fsS "$BASE_URL/readyz" >/dev/null 2>&1; then break; fi
  sleep 1
done
if ! curl -fsS "$BASE_URL/readyz" >/dev/null; then
  echo "server failed readiness check; log:"
  show_log "$SERVER_LOG" "server"
  finish_step_failed
  exit 1
fi
finish_step

begin_step "== 7/9 protocol bot + replay =="
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
  exit 1
fi
if [[ "${ARPG_VERBOSE:-0}" == "1" ]]; then
  cat "$BOT_LOG"
else
  echo "OK: protocol bot"
fi
rm -f "$BOT_LOG"
echo "bot completed session: $SESSION_ID"
"$RUN_QUIET" --label "arpg-replay" -- bash -c \
  "cd server && ARPG_DATABASE_URL=\"$DATABASE_URL\" go run ./cmd/arpg-replay --session-id \"$SESSION_ID\""
finish_step

begin_step "== 8/9 Godot client bot scenarios =="
env \
  GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" \
  SCENARIO=all HEADLESS=1 ./scripts/bot_client.sh
finish_step

begin_step "== 9/9 Godot headless smoke (optional) =="
env \
  GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" DEBUG_TOKEN="$DEBUG_TOKEN" \
  ./scripts/client_smoke.sh
finish_step

echo "CI OK"
