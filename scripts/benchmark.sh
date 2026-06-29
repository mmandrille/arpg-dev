#!/usr/bin/env bash
# Run all scenarios marked ci_tier=benchmark with ARPG_PERF_DEBUG=1.
#
# Flow:
#   1. Start server with ARPG_PERF_DEBUG=1 (server log kept as artifact)
#   2. Protocol bot records the scenarios (writes replay manifest)
#   3. Godot opens in visual mode — watch the replay with the Performance
#      status overlay visible in the top-right corner
#   4. After Godot closes, generate a perf report from the server log
#
# Godot is optional: if not found the script falls back to protocol-only
# mode and still generates the report.
#
# Usage:
#   make benchmark
#   make benchmark BENCHMARK_OUT=docs/performance/reports/myrun.txt
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
# shellcheck source=quiet_helpers.sh
source "$ROOT/scripts/quiet_helpers.sh"
# shellcheck source=godot_ci_flags.sh
source "$ROOT/scripts/godot_ci_flags.sh"

DATABASE_URL="${ARPG_DATABASE_URL:-postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable}"
if [[ -n "${ARPG_ADDR:-}" ]]; then
  ADDR="$ARPG_ADDR"
else
  ADDR=":0"
fi
BASE_URL="${BASE_URL:-}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-${DEV_TOKEN:-local-dev-token}}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-${DEBUG_TOKEN:-local-debug-token}}"
GODOT="${GODOT:-godot}"
BENCHMARK_OUT="${BENCHMARK_OUT:-}"
AUTOPLAY_STEP_DELAY="${AUTOPLAY_STEP_DELAY:-0.45}"

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARTIFACTS_DIR="$ROOT/.artifacts/benchmark-runs"
mkdir -p "$ARTIFACTS_DIR"
SERVER_LOG="$ARTIFACTS_DIR/${TIMESTAMP}-server.log"
BOT_LOG="$ARTIFACTS_DIR/${TIMESTAMP}-bot.log"
MANIFEST="$ARTIFACTS_DIR/${TIMESTAMP}-manifest.json"
if [[ -z "$BENCHMARK_OUT" ]]; then
  BENCHMARK_OUT="$ARTIFACTS_DIR/${TIMESTAMP}-report.txt"
fi

# Check for Godot — degrade gracefully to protocol-only if missing
HAS_GODOT=0
if command -v "$GODOT" >/dev/null 2>&1; then
  HAS_GODOT=1
else
  echo "[benchmark] WARNING: Godot runtime '$GODOT' not found — will skip visual replay."
  echo "[benchmark]           Set GODOT=/path/to/godot to enable the visual client."
fi

SERVER_PID=""
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT

# ── 1. Build and start server ─────────────────────────────────────────────────

echo "[benchmark] building server..."
SERVER_BIN="$(mktemp -t arpg-benchmark-server.XXXXXX)"
"$RUN_QUIET" --label "go build arpg-server" -- bash -c "cd server && go build -o \"$SERVER_BIN\" ./cmd/arpg-server"

echo "[benchmark] starting server with ARPG_PERF_DEBUG=1 (log: $SERVER_LOG)..."
ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
  ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_GAMEPLAY_DEBUG=true \
  ARPG_PERF_DEBUG=1 \
  ARPG_RULES_DIR="$ROOT/shared/rules" \
  "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "[benchmark] waiting for server readiness..."
if [[ -z "$BASE_URL" && "$ADDR" == ":0" ]]; then
  for i in $(seq 1 60); do
    if [[ -s "$SERVER_LOG" ]]; then
      PORT="$(python3 - "$SERVER_LOG" <<'PY'
import json, re, sys
for line in open(sys.argv[1], encoding="utf-8"):
    try:
        data = json.loads(line)
    except json.JSONDecodeError:
        continue
    if data.get("message") != "server listening":
        continue
    m = re.search(r":([0-9]+)$", str(data.get("addr", "")))
    if m:
        print(m.group(1))
        raise SystemExit(0)
raise SystemExit(1)
PY
)" && break
    fi
    if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
      echo "[benchmark] server exited early; log:"
      show_log "$SERVER_LOG" "server"
      exit 1
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
    echo "[benchmark] server exited early; log:"
    show_log "$SERVER_LOG" "server"
    exit 1
  fi
  sleep 1
done
curl -fsS "${BASE_URL%/}/readyz" >/dev/null

# ── 2. Protocol bot — records all benchmark scenarios ─────────────────────────

echo "[benchmark] recording benchmark scenarios (manifest: $MANIFEST)..."
"$ROOT/.venv/bin/python" -m tools.bot.run \
  --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --debug-token "$DEBUG_TOKEN" \
  --email "benchmark@example.test" --scenario benchmark \
  --write-manifest "$MANIFEST" \
  2>"$BOT_LOG" || { echo "[benchmark] bot recording failed; check $BOT_LOG"; cat "$BOT_LOG"; exit 1; }

# ── 3. Godot visual replay ────────────────────────────────────────────────────

if [[ "$HAS_GODOT" -eq 1 ]]; then
  echo "[benchmark] importing Godot assets..."
  "$RUN_QUIET" --label "Godot asset import" -- bash -c \
    'source "$0" && "$1" $GODOT_HEADLESS_FLAGS --path "$2/client" --import || true' \
    "$ROOT/scripts/godot_ci_flags.sh" "$GODOT" "$ROOT"

  echo "[benchmark] launching Godot visual replay — watch the Performance status overlay (top-right)."
  echo "[benchmark]   ARPG_PERF_DEBUG=1  |  status_text=on by default  |  step_delay=${AUTOPLAY_STEP_DELAY}s"
  ARPG_BASE_URL="$BASE_URL" \
    ARPG_DEV_TOKEN="$DEV_TOKEN" \
    ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
    ARPG_EMAIL="benchmark@example.test" \
    ARPG_PERF_DEBUG=1 \
    ARPG_VISUAL_REPLAY_MANIFEST="$MANIFEST" \
    ARPG_AUTOPLAY_STEP_DELAY="$AUTOPLAY_STEP_DELAY" \
    ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE=1 \
    "$GODOT" --path "$ROOT/client"
  echo "[benchmark] Godot closed."
else
  echo "[benchmark] skipping visual replay (Godot not available)."
fi

# ── 4. Shut down server + generate report ─────────────────────────────────────

echo "[benchmark] shutting down server..."
kill "$SERVER_PID" >/dev/null 2>&1 || true
SERVER_PID=""
sleep 1

echo "[benchmark] generating report..."
"$ROOT/.venv/bin/python" -m tools.bot.benchmark_report \
  --server-log "$SERVER_LOG" \
  --bot-log "$BOT_LOG" \
  --out "$BENCHMARK_OUT"

echo ""
echo "[benchmark] artifacts saved:"
echo "  server log : $SERVER_LOG"
echo "  bot log    : $BOT_LOG"
echo "  manifest   : $MANIFEST"
echo "  report     : $BENCHMARK_OUT"
