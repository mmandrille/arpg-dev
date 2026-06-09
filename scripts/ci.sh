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

echo "== 1/8 shared schema validation =="
"$RUN_QUIET" --label validate-shared -- make validate-shared

echo "== 2/8 asset manifest + GLB validation =="
"$RUN_QUIET" --label validate-assets -- make validate-assets

echo "== 3/8 Go tests =="
"$RUN_QUIET" --label "go test ./..." -- bash -c 'cd server && go test ./...'

echo "== 4/8 Python unit checks =="
make tools >/dev/null
"$RUN_QUIET" --label "pytest tools" -- "$ROOT/.venv/bin/python" -m pytest -q tools

echo "== 5/8 start Postgres + server =="
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
  exit 1
fi

echo "== 6/8 protocol bot + replay =="
BOT_LOG="$(mktemp -t arpg-ci-bot.XXXXXX.log)"
set +e
SESSION_ID="$("$ROOT/.venv/bin/python" -m tools.bot.run \
  --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --debug-token "$DEBUG_TOKEN" \
  --print-session-id 2>"$BOT_LOG")"
bot_status=$?
set -e
if [[ "$bot_status" -ne 0 ]]; then
  echo "FAILED: protocol bot"
  show_log "$BOT_LOG" "protocol bot"
  rm -f "$BOT_LOG"
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

echo "== 7/8 Godot client bot scenarios =="
"$RUN_QUIET" --label "bot-client scenarios" -- env \
  GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" \
  SCENARIO=all HEADLESS=1 ./scripts/bot_client.sh

echo "== 8/8 Godot headless smoke (optional) =="
"$RUN_QUIET" --label "client smoke" -- env \
  GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" DEBUG_TOKEN="$DEBUG_TOKEN" \
  ./scripts/client_smoke.sh

echo "CI OK"
