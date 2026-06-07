#!/usr/bin/env bash
# Local protocol bot wrapper.
# Starts a temporary server, runs Python bot scenarios, then tears the server
# down. Postgres is expected to be up before this script is called.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

DATABASE_URL="${ARPG_DATABASE_URL:-postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable}"
if [[ -n "${ARPG_ADDR:-}" ]]; then
  ADDR="$ARPG_ADDR"
else
  ADDR=":0"
fi
BASE_URL="${BASE_URL:-}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-${DEV_TOKEN:-local-dev-token}}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-${DEBUG_TOKEN:-local-debug-token}}"
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
(cd server && go build -o "$SERVER_BIN" ./cmd/arpg-server)

echo "[bot-local] starting server on $ADDR (log: $SERVER_LOG)..."
ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
  ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_RULES_DIR="$ROOT/shared/rules" \
  "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "[bot-local] waiting for server readiness..."
if [[ -z "$BASE_URL" && "$ADDR" == ":0" ]]; then
  for i in $(seq 1 60); do
    if [[ -s "$SERVER_LOG" ]]; then
      PORT="$(python3 - "$SERVER_LOG" <<'PY'
import json
import re
import sys
path = sys.argv[1]
for line in open(path, encoding="utf-8"):
    try:
        data = json.loads(line)
    except json.JSONDecodeError:
        continue
    if data.get("message") != "server listening":
        continue
    addr = str(data.get("addr", ""))
    match = re.search(r":([0-9]+)$", addr)
    if match:
        print(match.group(1))
        raise SystemExit(0)
raise SystemExit(1)
PY
)" && break
    fi
    if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
      echo "[bot-local] server exited early; log:"; cat "$SERVER_LOG"; exit 1
    fi
    sleep 0.1
  done
  BASE_URL="http://localhost:${PORT:?}"
elif [[ -z "$BASE_URL" ]]; then
  BASE_URL="http://localhost:${ADDR#:}"
fi
for i in $(seq 1 60); do
  if curl -fsS "${BASE_URL%/}/readyz" >/dev/null 2>&1; then break; fi
  if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
    echo "[bot-local] server exited early; log:"; cat "$SERVER_LOG"; exit 1
  fi
  sleep 1
done
curl -fsS "${BASE_URL%/}/readyz" >/dev/null
if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
  echo "[bot-local] server exited before bot could start; log:"; cat "$SERVER_LOG"; exit 1
fi

echo "[bot-local] running protocol bot scenario selection '$SCENARIO'..."
"$ROOT/.venv/bin/python" -m tools.bot.run \
  --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --debug-token "$DEBUG_TOKEN" \
  --email "$EMAIL" --scenario "$SCENARIO"

echo "[bot-local] scenarios complete; shutting down server."
