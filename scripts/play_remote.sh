#!/usr/bin/env bash
# Launch one or more Godot menu clients against an already running backend.
# This intentionally does not start Postgres, run migrations, or build/start the
# local Go server.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

GODOT="${GODOT:-godot}"
BASE_URL="${BASE_URL:-http://localhost:8888}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-local-dev-token}"
PLAY_CLIENTS="${PLAY_CLIENTS:-1}"

if ! [[ "$PLAY_CLIENTS" =~ ^[0-9]+$ ]] || (( PLAY_CLIENTS < 1 )); then
  echo "[play-remote] PLAY_CLIENTS must be a positive integer; got '$PLAY_CLIENTS'."
  exit 1
fi

if ! command -v "$GODOT" >/dev/null 2>&1; then
  echo "[play-remote] Godot runtime '$GODOT' not found on PATH."
  echo "[play-remote] Install Godot $(cat "$ROOT/.godot-version") and re-run, or set GODOT=/path/to/godot."
  exit 1
fi

READY_URL="${BASE_URL%/}/readyz"
echo "[play-remote] probing $READY_URL..."
if ! curl -fsS "$READY_URL" >/dev/null; then
  echo "[play-remote] backend is not ready at $READY_URL."
  exit 1
fi

"$GODOT" --headless --path "$ROOT/client" --import >/dev/null 2>&1 || true

CLIENT_PIDS=()
cleanup() {
  for pid in "${CLIENT_PIDS[@]:-}"; do
    kill "$pid" >/dev/null 2>&1 || true
  done
}
trap cleanup EXIT

echo "[play-remote] launching $PLAY_CLIENTS Godot client(s) against $BASE_URL."
for idx in $(seq 1 "$PLAY_CLIENTS"); do
  EMAIL="${ARPG_PLAY_EMAIL_PREFIX:-remote-player}${idx}+play-$(date +%s)@example.test"
  env -u ARPG_AUTOSTART -u ARPG_WORLD_ID -u ARPG_SESSION_ID \
    ARPG_BASE_URL="$BASE_URL" \
    ARPG_DEV_TOKEN="$DEV_TOKEN" \
    ARPG_EMAIL="$EMAIL" \
    "$GODOT" --path "$ROOT/client" &
  CLIENT_PIDS+=("$!")
done

for pid in "${CLIENT_PIDS[@]}"; do
  wait "$pid" || true
done

echo "[play-remote] client(s) closed."
