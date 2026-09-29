#!/usr/bin/env bash
# Local protocol bot wrapper.
# Starts a temporary server, runs Python bot scenarios, then tears the server
# down. Postgres is expected to be up before this script is called.
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
PERF_DEBUG="${ARPG_PERF_DEBUG:-false}"
EMAIL="${ARPG_EMAIL:-bot@example.test}"
SCENARIO="${ARPG_BOT_SCENARIO:-${SCENARIO:-${scenario:-all}}}"

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-bot-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "[bot-local] building server..."
SERVER_BIN="$(mktemp -t arpg-bot-server.XXXXXX)"
"$RUN_QUIET" --label "go build arpg-server" -- bash -c "cd server && go build -o \"$SERVER_BIN\" ./cmd/arpg-server"

"$ROOT/scripts/test_db.sh" ensure "$DATABASE_URL"
arpg_resolve_server_addr
echo "[bot-local] starting server on $ADDR db=$(arpg_db_name "$DATABASE_URL") (log: $SERVER_LOG)..."
ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
  ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_GAMEPLAY_DEBUG="$GAMEPLAY_DEBUG" \
  ARPG_PERF_DEBUG="$PERF_DEBUG" \
  ARPG_RULES_DIR="$ROOT/shared/rules" \
  "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "[bot-local] waiting for server readiness..."
arpg_wait_own_server "bot-local"

echo "[bot-local] running protocol bot scenario selection '$SCENARIO'..."
export ARPG_BOT_SERVER_LOG="$SERVER_LOG"
"$RUN_QUIET" --label "protocol bot ($SCENARIO)" -- \
  "$ROOT/.venv/bin/python" -m tools.bot.run \
  --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --debug-token "$DEBUG_TOKEN" \
  --email "$EMAIL" --scenario "$SCENARIO" --cleanup-characters

echo "[bot-local] scenarios complete; shutting down server."
