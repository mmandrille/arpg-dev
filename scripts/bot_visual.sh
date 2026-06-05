#!/usr/bin/env bash
# Launch the server + interactive Godot client in visual-bot mode.
# The Godot client drives the same slice flow visibly through the normal
# realtime protocol, then leaves the window open for inspection.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GODOT="${GODOT:-godot}"
GODOT_FLAGS="${GODOT_FLAGS:-}"
DATABASE_URL="${ARPG_DATABASE_URL:-postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable}"
ADDR="${ARPG_ADDR:-:8080}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-${DEV_TOKEN:-local-dev-token}}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-${DEBUG_TOKEN:-local-debug-token}}"
AUTOPLAY_STEP_DELAY="${AUTOPLAY_STEP_DELAY:-0.45}"

if ! command -v "$GODOT" >/dev/null 2>&1; then
  echo "[bot-visual] Godot runtime '$GODOT' not found on PATH."
  echo "[bot-visual] Install Godot $(cat "$ROOT/.godot-version") and re-run, or set GODOT=/path/to/godot."
  exit 1
fi

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-bot-visual-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "[bot-visual] building server..."
SERVER_BIN="$(mktemp -t arpg-bot-visual-server.XXXXXX)"
(cd server && go build -o "$SERVER_BIN" ./cmd/arpg-server)

echo "[bot-visual] starting server on $ADDR (log: $SERVER_LOG)..."
ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
  ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_RULES_DIR="$ROOT/shared/rules" \
  "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "[bot-visual] waiting for server readiness..."
for i in $(seq 1 60); do
  if curl -fsS "$BASE_URL/readyz" >/dev/null 2>&1; then break; fi
  if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
    echo "[bot-visual] server exited early; log:"; cat "$SERVER_LOG"; exit 1
  fi
  sleep 1
done
curl -fsS "$BASE_URL/readyz" >/dev/null

"$GODOT" --headless --path "$ROOT/client" --import >/dev/null 2>&1 || true

echo "[bot-visual] launching Godot visual bot."
echo "[bot-visual] AUTOPLAY_STEP_DELAY=$AUTOPLAY_STEP_DELAY; close the window to stop the server."
ARPG_BASE_URL="$BASE_URL" ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_AUTOPLAY=1 ARPG_AUTOPLAY_STEP_DELAY="$AUTOPLAY_STEP_DELAY" \
  "$GODOT" $GODOT_FLAGS --path "$ROOT/client"

echo "[bot-visual] client closed; shutting down server."
