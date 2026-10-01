#!/usr/bin/env bash
# Local convenience wrapper for Godot client bot scenarios.
# Starts a temporary server, runs the low-level bot client runner, then tears the
# server down. Postgres is expected to be up before this script is called.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
# shellcheck source=quiet_helpers.sh
source "$ROOT/scripts/quiet_helpers.sh"

# shellcheck source=server_helpers.sh
source "$ROOT/scripts/server_helpers.sh"

DATABASE_URL="${ARPG_DATABASE_URL:-$("$ROOT/scripts/test_db.sh" url)}"
ADDR="${ARPG_ADDR:-}"
BASE_URL="${BASE_URL:-}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-${DEV_TOKEN:-local-dev-token}}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-${DEBUG_TOKEN:-local-debug-token}}"
GAMEPLAY_DEBUG="${ARPG_GAMEPLAY_DEBUG:-true}"

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-bot-client-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
  [[ -n "$SERVER_PID" ]] && wait "$SERVER_PID" >/dev/null 2>&1 || true
  if [[ -n "${BOT_CLIENT_LOG_DIR:-}" ]]; then
    mkdir -p "$BOT_CLIENT_LOG_DIR"
    grep -F 'backend_perf' "$SERVER_LOG" > "$BOT_CLIENT_LOG_DIR/server.log" || true
    echo "[bot-client-local] retained server performance counters: $BOT_CLIENT_LOG_DIR/server.log"
  fi
}
trap cleanup EXIT

echo "[bot-client-local] building server..."
SERVER_BIN="$(mktemp -t arpg-bot-client-server.XXXXXX)"
"$RUN_QUIET" --label "go build arpg-server" -- bash -c "cd server && go build -o \"$SERVER_BIN\" ./cmd/arpg-server"

"$ROOT/scripts/test_db.sh" ensure "$DATABASE_URL"
arpg_resolve_server_addr
echo "[bot-client-local] starting server on $ADDR db=$(arpg_db_name "$DATABASE_URL") (log: $SERVER_LOG)..."
ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
  ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_GAMEPLAY_DEBUG="$GAMEPLAY_DEBUG" \
  ARPG_RULES_DIR="$ROOT/shared/rules" \
  "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "[bot-client-local] waiting for server readiness..."
arpg_wait_own_server "bot-client-local"

GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" \
  SCENARIO="${SCENARIO:-all}" HEADLESS="${HEADLESS:-0}" ./scripts/bot_client.sh

echo "[bot-client-local] scenarios complete; shutting down server."
