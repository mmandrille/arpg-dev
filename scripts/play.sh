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
PLAY_CLIENTS="${PLAY_CLIENTS:-1}"

if ! [[ "$PLAY_CLIENTS" =~ ^[0-9]+$ ]] || (( PLAY_CLIENTS < 1 )); then
  echo "[play] PLAY_CLIENTS must be a positive integer; got '$PLAY_CLIENTS'."
  exit 1
fi

if ! command -v "$GODOT" >/dev/null 2>&1; then
  echo "[play] Godot runtime '$GODOT' not found on PATH."
  echo "[play] Install Godot $(cat "$ROOT/.godot-version") and re-run, or set GODOT=/path/to/godot."
  exit 1
fi

SERVER_PID=""
CLIENT_PIDS=()
SERVER_LOG="$(mktemp -t arpg-play-server.XXXXXX.log)"
cleanup() {
  for pid in "${CLIENT_PIDS[@]:-}"; do
    kill "$pid" >/dev/null 2>&1 || true
  done
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

if (( PLAY_CLIENTS == 1 )); then
  echo "[play] launching Godot client — close the window to stop the server."
  echo "[play] main menu opens first; gameplay controls: W/A/S/D move, LMB action, scroll zoom, I inventory."
  GODOT_ENV=(
    "ARPG_BASE_URL=$BASE_URL"
    "ARPG_DEV_TOKEN=$DEV_TOKEN"
  )
  [[ -n "${ARPG_AUTOSTART:-}" ]] && GODOT_ENV+=("ARPG_AUTOSTART=$ARPG_AUTOSTART")
  [[ -n "${ARPG_WORLD_ID:-}" ]] && GODOT_ENV+=("ARPG_WORLD_ID=$ARPG_WORLD_ID")
  [[ -n "${ARPG_SESSION_ID:-}" ]] && GODOT_ENV+=("ARPG_SESSION_ID=$ARPG_SESSION_ID")
  env -u ARPG_AUTOSTART -u ARPG_WORLD_ID -u ARPG_SESSION_ID \
    "${GODOT_ENV[@]}" "$GODOT" --path "$ROOT/client"
else
  echo "[play] launching Godot clients — close all windows to stop the server."
  echo "[play] co-op clients autostart in one shared session; controls: W/A/S/D move, LMB action, scroll zoom, I inventory."
  SETUP_JSON="$(mktemp -t arpg-play-coop.XXXXXX.json)"
  "$ROOT/.venv/bin/python" "$ROOT/tools/play/coop_setup.py" \
    --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --clients "$PLAY_CLIENTS" \
    --email-prefix "${ARPG_PLAY_EMAIL_PREFIX:-player}" \
    --world-id "${ARPG_WORLD_ID:-dungeon_levels}" >"$SETUP_JSON"
  SESSION_ID="$("$ROOT/.venv/bin/python" -c 'import json,sys; print(json.load(open(sys.argv[1]))["session_id"])' "$SETUP_JSON")"
  echo "[play] launching $PLAY_CLIENTS Godot clients in co-op session $SESSION_ID."
  for idx in $(seq 1 "$PLAY_CLIENTS"); do
    EMAIL="$("$ROOT/.venv/bin/python" -c 'import json,sys; print(json.load(open(sys.argv[1]))["clients"][int(sys.argv[2])-1]["email"])' "$SETUP_JSON" "$idx")"
    env -u ARPG_WORLD_ID \
      ARPG_BASE_URL="$BASE_URL" \
      ARPG_DEV_TOKEN="$DEV_TOKEN" \
      ARPG_EMAIL="$EMAIL" \
      ARPG_AUTOSTART=1 \
      ARPG_SESSION_ID="$SESSION_ID" \
      "$GODOT" --path "$ROOT/client" &
    CLIENT_PIDS+=("$!")
  done
  for pid in "${CLIENT_PIDS[@]}"; do
    wait "$pid" || true
  done
fi

echo "[play] client(s) closed; shutting down server."
