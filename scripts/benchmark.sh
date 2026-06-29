#!/usr/bin/env bash
# Live concurrent benchmark: protocol bot + Godot client on the same live session.
#
# Architecture:
#   1. Start server once (ARPG_PERF_DEBUG=1) — kept running for all scenarios.
#   2. For each benchmark scenario:
#      a. Run the bot in the background with --write-session-id $SID_FILE.
#         The bot creates a listed-coop session and writes the session ID to the
#         file before driving the scenario, then runs the full scenario.
#      b. Wait for the session ID file to appear (bot has created the session).
#      c. Wait 3 seconds for the session to stabilise.
#      d. Launch Godot (full visual, no --headless) with ARPG_JOIN_SESSION_ID=$SID
#         so it joins the live session and renders real-time gameplay.
#      e. Wait for the bot to finish driving the scenario.
#      f. Kill Godot, collect per-scenario logs.
#   3. Combine all per-scenario client logs into a single client.log.
#   4. Generate a unified report via tools/bot/benchmark_report.py.
#
# Godot is optional — if not found the script degrades gracefully to bot-only
# (no visual, no client log) and still generates the report.
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

# Bot uses a dedicated account; Godot observer uses a separate account so both
# can be authenticated concurrently against the same server.
BOT_EMAIL="benchmark-bot@example.test"
OBSERVER_EMAIL="benchmark-observer@example.test"

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
ARTIFACTS_DIR="$ROOT/.artifacts/benchmark-runs/$TIMESTAMP"
mkdir -p "$ARTIFACTS_DIR"
SERVER_LOG="$ARTIFACTS_DIR/server.log"
BOT_LOG="$ARTIFACTS_DIR/bot.log"
CLIENT_LOG="$ARTIFACTS_DIR/client.log"   # combined after all scenarios
if [[ -z "$BENCHMARK_OUT" ]]; then
  BENCHMARK_OUT="$ARTIFACTS_DIR/report.txt"
fi

# Temp dir for per-scenario session-ID handoff files
SID_DIR="$(mktemp -d -t arpg-benchmark-sid.XXXXXX)"
cleanup_sid_dir() { rm -rf "$SID_DIR"; }
trap cleanup_sid_dir EXIT

# ── Godot availability check ──────────────────────────────────────────────────

HAS_GODOT=0
if command -v "$GODOT" >/dev/null 2>&1; then
  HAS_GODOT=1
else
  echo "[benchmark] WARNING: Godot runtime '$GODOT' not found — will run bot-only (no visual client)."
  echo "[benchmark]           Set GODOT=/path/to/godot to enable the visual client."
fi

# ── Process tracking ──────────────────────────────────────────────────────────

SERVER_PID=""
GODOT_PID=""
BOT_PID=""

cleanup() {
  [[ -n "$BOT_PID" ]]    && kill "$BOT_PID"    >/dev/null 2>&1 || true
  [[ -n "$GODOT_PID" ]]  && kill "$GODOT_PID"  >/dev/null 2>&1 || true
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

# ── 2. Enumerate benchmark scenarios ─────────────────────────────────────────

echo "[benchmark] enumerating benchmark scenarios..."
SCENARIO_IDS="$("$ROOT/.venv/bin/python" -c \
  "from tools.bot.run import load_scenarios,select_scenarios; print(' '.join(s.id for s in select_scenarios(load_scenarios(),'benchmark')))")"

if [[ -z "$SCENARIO_IDS" ]]; then
  echo "[benchmark] no benchmark scenarios found — nothing to do."
  exit 0
fi

echo "[benchmark] scenarios: $SCENARIO_IDS"

# Optionally import Godot assets once before the scenario loop
if [[ "$HAS_GODOT" -eq 1 ]]; then
  echo "[benchmark] importing Godot assets..."
  "$RUN_QUIET" --label "Godot asset import" -- bash -c \
    'source "$0" && "$1" $GODOT_HEADLESS_FLAGS --path "$2/client" --import || true' \
    "$ROOT/scripts/godot_ci_flags.sh" "$GODOT" "$ROOT"
fi

# ── 3. Per-scenario concurrent run ───────────────────────────────────────────

# Accumulate per-scenario client logs; combined into $CLIENT_LOG at the end.
SCENARIO_CLIENT_LOGS=()

for SCENARIO_ID in $SCENARIO_IDS; do
  echo ""
  echo "[benchmark] ── scenario: $SCENARIO_ID ──────────────────────────────"

  SCENARIO_BOT_LOG="$ARTIFACTS_DIR/${SCENARIO_ID}-bot.log"
  SCENARIO_CLIENT_LOG="$ARTIFACTS_DIR/${SCENARIO_ID}-client.log"
  SID_FILE="$SID_DIR/${SCENARIO_ID}.sid"

  # Check if this scenario requires a solo session (e.g. multi-level dungeon worlds
  # that don't spawn monsters in coop mode). Search by id field, not filename,
  # because scenario files are prefixed with numbers.
  SOLO_SESSION="$("$ROOT/.venv/bin/python" -c "
import json, pathlib
for p in sorted(pathlib.Path('tools/bot/scenarios').glob('*.json')):
    try:
        d = json.loads(p.read_text())
        if d.get('id') == '${SCENARIO_ID}':
            print('1' if d.get('benchmark_solo_session') else '0')
            break
    except Exception:
        pass
else:
    print('0')
")"

  # 3a. Launch bot in background.
  #   - Coop scenarios: creates a listed-coop session and writes the session ID
  #     so Godot can join as an observer.
  #   - Solo scenarios (benchmark_solo_session=true): creates a solo session for
  #     correct world initialization; Godot is skipped for this scenario.
  echo "[benchmark]   starting bot for $SCENARIO_ID (email: $BOT_EMAIL, solo=$SOLO_SESSION)..."
  if [[ "$SOLO_SESSION" -eq 0 ]]; then
    ARPG_PERF_DEBUG=1 \
      "$ROOT/.venv/bin/python" -m tools.bot.run \
        --base-url "$BASE_URL" \
        --dev-token "$DEV_TOKEN" \
        --debug-token "$DEBUG_TOKEN" \
        --email "$BOT_EMAIL" \
        --scenario "$SCENARIO_ID" \
        --write-session-id "$SID_FILE" \
        --skip-replay \
        >>"$SCENARIO_BOT_LOG" 2>&1 &
  else
    # Solo scenarios use a unique email per run to avoid character state
    # contamination from previously-run coop scenarios on the shared server.
    ARPG_PERF_DEBUG=1 \
      "$ROOT/.venv/bin/python" -m tools.bot.run \
        --base-url "$BASE_URL" \
        --dev-token "$DEV_TOKEN" \
        --debug-token "$DEBUG_TOKEN" \
        --email "benchmark-solo-${SCENARIO_ID}@example.test" \
        --scenario "$SCENARIO_ID" \
        --skip-replay \
        --cleanup-characters \
        >>"$SCENARIO_BOT_LOG" 2>&1 &
  fi
  BOT_PID=$!

  # 3b. Wait for the session ID file to appear (bot has created the session).
  echo "[benchmark]   waiting for session ID from bot..."
  SID_WAIT=0
  SID_FOUND=0
  while [[ $SID_WAIT -lt 30 ]]; do
    if [[ -s "$SID_FILE" ]]; then
      SID_FOUND=1
      break
    fi
    if ! kill -0 "$BOT_PID" >/dev/null 2>&1; then
      echo "[benchmark]   WARNING: bot exited before writing session ID — skipping Godot for $SCENARIO_ID."
      break
    fi
    sleep 0.5
    SID_WAIT=$((SID_WAIT + 1))
  done

  if [[ "$SOLO_SESSION" -eq 1 ]]; then
    echo "[benchmark]   solo session — skipping Godot observer for $SCENARIO_ID (server metrics only)."
    if wait "$BOT_PID"; then
      echo "[benchmark]   bot finished $SCENARIO_ID successfully."
    else
      echo "[benchmark]   bot FAILED for $SCENARIO_ID — check $SCENARIO_BOT_LOG"
    fi
    BOT_PID=""
    SCENARIO_CLIENT_LOGS+=("$SCENARIO_CLIENT_LOG")  # empty file, still track
    continue
  fi

  if [[ "$SID_FOUND" -eq 0 ]]; then
    echo "[benchmark]   skipping Godot for $SCENARIO_ID (no session ID)."
    wait "$BOT_PID" || echo "[benchmark]   bot FAILED for $SCENARIO_ID — check $SCENARIO_BOT_LOG"
    BOT_PID=""
    continue
  fi

  SESSION_ID="$(cat "$SID_FILE")"
  echo "[benchmark]   session ID: $SESSION_ID"

  # 3c. Wait 3 seconds for the session to stabilise before Godot joins.
  echo "[benchmark]   waiting 3s for session to stabilise..."
  sleep 3

  # 3d. Launch Godot (full visual, no --headless) to join the live session.
  if [[ "$HAS_GODOT" -eq 1 ]]; then
    echo "[benchmark]   launching Godot observer (email: $OBSERVER_EMAIL, session: $SESSION_ID)..."
    ARPG_BASE_URL="$BASE_URL" \
      ARPG_DEV_TOKEN="$DEV_TOKEN" \
      ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
      ARPG_EMAIL="$OBSERVER_EMAIL" \
      ARPG_JOIN_SESSION_ID="$SESSION_ID" \
      ARPG_PERF_DEBUG=1 \
      "$GODOT" --path "$ROOT/client" 2>&1 | tee "$SCENARIO_CLIENT_LOG" &
    GODOT_PID=$!
  fi

  # 3e. Wait for bot to finish driving the full scenario.
  echo "[benchmark]   waiting for bot to complete $SCENARIO_ID..."
  if wait "$BOT_PID"; then
    echo "[benchmark]   bot finished $SCENARIO_ID successfully."
  else
    echo "[benchmark]   bot FAILED for $SCENARIO_ID — check $SCENARIO_BOT_LOG"
  fi
  BOT_PID=""

  # 3f. Kill Godot now that the bot is done.
  if [[ -n "$GODOT_PID" ]]; then
    echo "[benchmark]   closing Godot..."
    kill "$GODOT_PID" >/dev/null 2>&1 || true
    wait "$GODOT_PID" 2>/dev/null || true
    GODOT_PID=""
    echo "[benchmark]   Godot closed."
  fi

  # Accumulate client log for this scenario (may be empty if no Godot).
  if [[ -s "$SCENARIO_CLIENT_LOG" ]]; then
    SCENARIO_CLIENT_LOGS+=("$SCENARIO_CLIENT_LOG")
  fi
done

# ── 4. Combine bot logs ───────────────────────────────────────────────────────

# Merge all per-scenario bot logs into the single $BOT_LOG artifact.
echo "" > "$BOT_LOG"
for SCENARIO_ID in $SCENARIO_IDS; do
  SCENARIO_BOT_LOG="$ARTIFACTS_DIR/${SCENARIO_ID}-bot.log"
  if [[ -s "$SCENARIO_BOT_LOG" ]]; then
    cat "$SCENARIO_BOT_LOG" >> "$BOT_LOG"
  fi
done

# Combine per-scenario client logs into unified client.log.
if [[ "${#SCENARIO_CLIENT_LOGS[@]}" -gt 0 ]]; then
  cat "${SCENARIO_CLIENT_LOGS[@]}" > "$CLIENT_LOG"
fi

# ── 5. Shut down server ───────────────────────────────────────────────────────

echo ""
echo "[benchmark] shutting down server..."
kill "$SERVER_PID" >/dev/null 2>&1 || true
SERVER_PID=""
sleep 1

# ── 6. Generate unified report ────────────────────────────────────────────────

echo "[benchmark] generating report..."
CLIENT_LOG_ARG=""
[[ -s "$CLIENT_LOG" ]] && CLIENT_LOG_ARG="--client-log $CLIENT_LOG"
# shellcheck disable=SC2086
"$ROOT/.venv/bin/python" -m tools.bot.benchmark_report \
  --server-log "$SERVER_LOG" \
  --bot-log "$BOT_LOG" \
  $CLIENT_LOG_ARG \
  --out "$BENCHMARK_OUT"

echo ""
echo "[benchmark] artifacts saved under $ARTIFACTS_DIR:"
echo "  server log  : $SERVER_LOG"
echo "  bot log     : $BOT_LOG"
[[ -s "$CLIENT_LOG" ]] && echo "  client log  : $CLIENT_LOG"
for SCENARIO_ID in $SCENARIO_IDS; do
  CL="$ARTIFACTS_DIR/${SCENARIO_ID}-client.log"
  [[ -s "$CL" ]] && echo "  client log  : $CL  (scenario: $SCENARIO_ID)"
done
echo "  report      : $BENCHMARK_OUT"
