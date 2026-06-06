#!/usr/bin/env bash
# Local CI aggregation for the first playable vertical slice.
# Runs shared schema validation, Go tests, Python unit checks, the end-to-end
# bot + replay flow against a throwaway Postgres + server, and the Godot
# headless smoke when the runtime is available.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DATABASE_URL="${ARPG_DATABASE_URL:-postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable}"
ADDR="${ARPG_ADDR:-:8080}"
BASE_URL="${BASE_URL:-http://localhost:8080}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-local-dev-token}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-local-debug-token}"

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-ci-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "== 1/7 shared schema validation =="
make validate-shared

echo "== 2/7 asset manifest + GLB validation =="
make validate-assets

echo "== 3/7 Go tests =="
(cd server && go test ./...)

echo "== 4/7 Python unit checks =="
make tools >/dev/null
"$ROOT/.venv/bin/python" -m pytest -q tools || { echo "pytest failed"; exit 1; }

echo "== 5/7 start Postgres + server =="
make db-up
# Build a binary and run it directly (not via `go run`, whose child binary
# would survive the cleanup kill and, if stdout were piped, hold the pipe open).
SERVER_BIN="$(mktemp -t arpg-ci-server.XXXXXX)"
(cd server && go build -o "$SERVER_BIN" ./cmd/arpg-server)
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
curl -fsS "$BASE_URL/readyz" >/dev/null

echo "== 6/8 protocol bot + replay =="
SESSION_ID="$("$ROOT/.venv/bin/python" -m tools.bot.run \
  --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --debug-token "$DEBUG_TOKEN" \
  --print-session-id)"
echo "bot completed session: $SESSION_ID"
(cd server && ARPG_DATABASE_URL="$DATABASE_URL" go run ./cmd/arpg-replay --session-id "$SESSION_ID")

echo "== 7/8 Godot client bot scenarios =="
GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" \
  SCENARIO=all HEADLESS=1 ./scripts/bot_client.sh

echo "== 8/8 Godot headless smoke (optional) =="
GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" DEBUG_TOKEN="$DEBUG_TOKEN" \
  ./scripts/client_smoke.sh

echo "CI OK"
