#!/usr/bin/env bash
# Launch the server + interactive Godot client in visual-bot mode.
# The Godot client drives the same slice flow visibly through the normal
# realtime protocol, then leaves the window open for inspection.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
# shellcheck source=quiet_helpers.sh
source "$ROOT/scripts/quiet_helpers.sh"

CLIENT_SCENARIOS_DIR="$ROOT/tools/bot/scenarios/client"
GODOT="${GODOT:-godot}"
GODOT_FLAGS="${GODOT_FLAGS:-}"
if [[ "${HEADLESS:-0}" == "1" && "$GODOT_FLAGS" != *"--headless"* ]]; then
  GODOT_FLAGS="--headless $GODOT_FLAGS"
fi
DATABASE_URL="${ARPG_DATABASE_URL:-postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable}"
ADDR="${ARPG_ADDR:-:8888}"
BASE_URL="${BASE_URL:-http://localhost:8888}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-${DEV_TOKEN:-local-dev-token}}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-${DEBUG_TOKEN:-local-debug-token}}"
PERF_DEBUG="${ARPG_PERF_DEBUG:-false}"
EMAIL="${ARPG_EMAIL:-bot@example.test}"
AUTOPLAY_STEP_DELAY="${AUTOPLAY_STEP_DELAY:-0.45}"
EXIT_ON_COMPLETE="${ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE:-1}"
RECORDING_CLEANUP_CHARACTERS="${ARPG_VISUAL_REPLAY_CLEANUP_CHARACTERS:-0}"
MANIFEST="${ARPG_VISUAL_REPLAY_MANIFEST:-$ROOT/.artifacts/bot-runs/$(date -u +%Y%m%dT%H%M%SZ)-visual.json}"
SCENARIO="${ARPG_BOT_SCENARIO:-${SCENARIO:-${scenario:-all}}}"
HEADLESS_REPLAY=0
if [[ "$GODOT_FLAGS" == *"--headless"* ]]; then
  HEADLESS_REPLAY=1
fi

if ! command -v "$GODOT" >/dev/null 2>&1; then
  echo "[bot-visual] Godot runtime '$GODOT' not found on PATH."
  echo "[bot-visual] Install Godot $(cat "$ROOT/.godot-version") and re-run, or set GODOT=/path/to/godot."
  exit 1
fi

is_client_scenario_selection() {
  local selection="$1"
  local f bn
  if [[ -f "$selection" ]]; then
    case "$selection" in
      "$CLIENT_SCENARIOS_DIR"/*.json|tools/bot/scenarios/client/*.json|*/tools/bot/scenarios/client/*.json)
        return 0
        ;;
    esac
    return 1
  fi
  while IFS= read -r -d '' f; do
    bn="$(basename "$f" .json)"
    if [[ "$bn" == "$selection" || "$bn.json" == "$selection" || "$bn" == *"_$selection" || "$bn" == "$selection"* ]]; then
      return 0
    fi
  done < <(find "$CLIENT_SCENARIOS_DIR" -maxdepth 1 -name '*.json' -print0 | sort -z)
  return 1
}

if [[ "$SCENARIO" != "all" ]] && is_client_scenario_selection "$SCENARIO"; then
  echo "[bot-visual] '$SCENARIO' is a Godot client scenario; running visible client bot instead of protocol replay recording."
  BOT_STEP_DELAY="${BOT_STEP_DELAY:-$AUTOPLAY_STEP_DELAY}" \
    GODOT="$GODOT" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" DEBUG_TOKEN="$DEBUG_TOKEN" \
    SCENARIO="$SCENARIO" HEADLESS="${HEADLESS:-0}" ./scripts/bot_client_local.sh
  exit $?
fi

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-bot-visual-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "[bot-visual] building server..."
SERVER_BIN="$(mktemp -t arpg-bot-visual-server.XXXXXX)"
"$RUN_QUIET" --label "go build arpg-server" -- bash -c "cd server && go build -o \"$SERVER_BIN\" ./cmd/arpg-server"

echo "[bot-visual] starting server on $ADDR (log: $SERVER_LOG)..."
ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
  ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
  ARPG_PERF_DEBUG="$PERF_DEBUG" \
  ARPG_RULES_DIR="$ROOT/shared/rules" \
  "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!

echo "[bot-visual] waiting for server readiness..."
for i in $(seq 1 60); do
  if curl -fsS "$BASE_URL/readyz" >/dev/null 2>&1; then break; fi
  if ! kill -0 "$SERVER_PID" >/dev/null 2>&1; then
    echo "[bot-visual] server exited early; log:"
    show_log "$SERVER_LOG" "server"
    exit 1
  fi
  sleep 1
done
curl -fsS "$BASE_URL/readyz" >/dev/null

recording_args=(
  "$ROOT/.venv/bin/python" -m tools.bot.run
  --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --debug-token "$DEBUG_TOKEN"
  --email "$EMAIL" --scenario "$SCENARIO" --write-manifest "$MANIFEST"
)
if [[ "$RECORDING_CLEANUP_CHARACTERS" == "1" ]]; then
  recording_args+=(--cleanup-characters)
fi

echo "[bot-visual] recording bot scenario selection '$SCENARIO' (manifest: $MANIFEST)..."
"$RUN_QUIET" --label "protocol bot recording ($SCENARIO)" -- "${recording_args[@]}"

"$RUN_QUIET" --label "Godot asset import" -- bash -c '"$1" --headless --path "$2/client" --import || true' _ "$GODOT" "$ROOT"

echo "[bot-visual] launching Godot visual replay playlist."
echo "[bot-visual] AUTOPLAY_STEP_DELAY=$AUTOPLAY_STEP_DELAY; EXIT_ON_COMPLETE=$EXIT_ON_COMPLETE."
if [[ "$HEADLESS_REPLAY" == "1" ]] && is_quiet_mode; then
  "$RUN_QUIET" --label "Godot visual replay" -- env \
    ARPG_BASE_URL="$BASE_URL" ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" ARPG_EMAIL="$EMAIL" \
    ARPG_PERF_DEBUG="$PERF_DEBUG" \
    ARPG_VISUAL_REPLAY_MANIFEST="$MANIFEST" ARPG_AUTOPLAY_STEP_DELAY="$AUTOPLAY_STEP_DELAY" \
    ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE="$EXIT_ON_COMPLETE" \
    "$GODOT" $GODOT_FLAGS --path "$ROOT/client"
else
  ARPG_BASE_URL="$BASE_URL" ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" ARPG_EMAIL="$EMAIL" \
    ARPG_PERF_DEBUG="$PERF_DEBUG" \
    ARPG_VISUAL_REPLAY_MANIFEST="$MANIFEST" ARPG_AUTOPLAY_STEP_DELAY="$AUTOPLAY_STEP_DELAY" \
    ARPG_VISUAL_REPLAY_EXIT_ON_COMPLETE="$EXIT_ON_COMPLETE" \
    "$GODOT" $GODOT_FLAGS --path "$ROOT/client"
fi

echo "[bot-visual] client closed; shutting down server."
