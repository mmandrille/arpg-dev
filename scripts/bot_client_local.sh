#!/usr/bin/env bash
# Local convenience wrapper for Godot client bot scenarios.
# Starts a temporary server, runs the low-level bot client runner, then tears the
# server down. Postgres is expected to be up before this script is called.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DATABASE_URL="${ARPG_DATABASE_URL:-postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable}"
ADDR="${ARPG_ADDR:-:8080}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-${DEV_TOKEN:-local-dev-token}}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-${DEBUG_TOKEN:-local-debug-token}}"

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-bot-client-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "[bot-client-local] building server..."
SERVER_BIN="$(mktemp -t arpg-bot-client-server.XXXXXX)"
(cd server && go build -o "$SERVER_BIN" ./cmd/arpg-server)

echo "[bot-client-local] starting server on $ADDR (log: $SERVER_LOG)..."
ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
  ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_RULES_DIR="$ROOT/shared/rules" \
  "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "[bot-client-local] waiting for server readiness..."
for i in $(seq 1 60); do
  if curl -fsS "${BASE_URL%/}/readyz" >/dev/null 2>&1; then break; fi
  if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
    echo "[bot-client-local] server exited early; log:"; cat "$SERVER_LOG"; exit 1
  fi
  sleep 1
done
curl -fsS "${BASE_URL%/}/readyz" >/dev/null
if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
  echo "[bot-client-local] server exited before bot-client could start; log:"; cat "$SERVER_LOG"; exit 1
fi

GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" \
  SCENARIO="${SCENARIO:-all}" HEADLESS="${HEADLESS:-0}" ./scripts/bot_client.sh

echo "[bot-client-local] scenarios complete; shutting down server."
