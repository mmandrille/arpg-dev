#!/usr/bin/env bash
# Launch the server + interactive Godot client so a human can play the slice.
# Postgres is expected to be up already (the `make play` target depends on db-up).
# The server runs in the background; closing the Godot window tears it down.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GODOT="${GODOT:-godot}"
DATABASE_URL="${ARPG_DATABASE_URL:-postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable}"
ADDR="${ARPG_ADDR:-:8080}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-local-dev-token}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-local-debug-token}"

if ! command -v "$GODOT" >/dev/null 2>&1; then
  echo "[play] Godot runtime '$GODOT' not found on PATH."
  echo "[play] Install Godot $(cat "$ROOT/.godot-version") and re-run, or set GODOT=/path/to/godot."
  exit 1
fi

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-play-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "[play] building server..."
SERVER_BIN="$(mktemp -t arpg-play-server.XXXXXX)"
(cd server && go build -o "$SERVER_BIN" ./cmd/arpg-server)

echo "[play] starting server on $ADDR (log: $SERVER_LOG)..."
ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
  ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_RULES_DIR="$ROOT/shared/rules" \
  "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "[play] waiting for server readiness..."
for i in $(seq 1 60); do
  if curl -fsS "$BASE_URL/readyz" >/dev/null 2>&1; then break; fi
  # Surface an early server crash instead of waiting the full timeout.
  if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
    echo "[play] server exited early; log:"; cat "$SERVER_LOG"; exit 1
  fi
  sleep 1
done
curl -fsS "$BASE_URL/readyz" >/dev/null

# Import assets once so the first interactive launch is clean.
"$GODOT" --headless --path "$ROOT/client" --import >/dev/null 2>&1 || true

echo "[play] launching Godot client — close the window to stop the server."
echo "[play] controls: W/A/S/D move, LMB action, scroll zoom, I inventory."
ARPG_BASE_URL="$BASE_URL" ARPG_DEV_TOKEN="$DEV_TOKEN" \
  ARPG_WORLD_ID="dungeon_levels" ARPG_SESSION_ID="" \
  "$GODOT" --path "$ROOT/client"

echo "[play] client closed; shutting down server."
