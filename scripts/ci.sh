#!/usr/bin/env bash
# Local CI aggregation for the first playable vertical slice.
# Runs all steps regardless of failures, then reports a summary.
# Default ARPG_CI_SCENARIO=ci runs tools/bot/ci_pack.json (~22 protocol + ~14 client).
# Use ARPG_CI_SCENARIO=all or `make ci-full` for the full scenario matrix.
# Quiet by default — VERBOSE=1 (or V=1) for full output.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
# shellcheck source=quiet_helpers.sh
source "$ROOT/scripts/quiet_helpers.sh"
# shellcheck source=server_helpers.sh
source "$ROOT/scripts/server_helpers.sh"

# Per-checkout DB + kernel-picked port by default so concurrent worktrees never
# share bot accounts or talk to each other's server (see server_helpers.sh).
DATABASE_URL="${ARPG_DATABASE_URL:-$("$ROOT/scripts/test_db.sh" url)}"
ADDR="${ARPG_ADDR:-}"
BASE_URL="${BASE_URL:-}"
DEV_TOKEN="${ARPG_DEV_TOKEN:-local-dev-token}"
DEBUG_TOKEN="${ARPG_DEBUG_TOKEN:-local-debug-token}"
GAMEPLAY_DEBUG="${ARPG_GAMEPLAY_DEBUG:-true}"
CI_SCENARIO="${ARPG_CI_SCENARIO:-ci}"

SERVER_PID=""
SERVER_LOG="$(mktemp -t arpg-ci-server.XXXXXX.log)"
cleanup() {
  [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT
CI_STARTED_AT="$(date +%s)"

FAILED_STEPS=()
STEP_PRINTED=0
STEP_STARTED_AT=0
STEP_LABEL=""

format_duration() {
  local total="$1"
  local hours=$((total / 3600))
  local minutes=$(((total % 3600) / 60))
  local seconds=$((total % 60))
  if [[ "$hours" -gt 0 ]]; then
    printf '%dh%02dm%02ds' "$hours" "$minutes" "$seconds"
  elif [[ "$minutes" -gt 0 ]]; then
    printf '%dm%02ds' "$minutes" "$seconds"
  else
    printf '%ds' "$seconds"
  fi
}

begin_step() {
  local title="$1"
  if [[ "$STEP_PRINTED" -eq 1 ]]; then
    echo
  fi
  echo "$title"
  STEP_LABEL="${title#== }"
  STEP_LABEL="${STEP_LABEL% ==}"
  STEP_STARTED_AT="$(date +%s)"
  STEP_PRINTED=1
}

finish_step() {
  local elapsed=$(( $(date +%s) - STEP_STARTED_AT ))
  local total_elapsed=$(( $(date +%s) - CI_STARTED_AT ))
  echo "completed in $(format_duration "$elapsed") (total $(format_duration "$total_elapsed"))"
  STEP_STARTED_AT=0
  STEP_LABEL=""
}

finish_step_failed() {
  local elapsed=$(( $(date +%s) - STEP_STARTED_AT ))
  local total_elapsed=$(( $(date +%s) - CI_STARTED_AT ))
  echo "FAILED after $(format_duration "$elapsed") (total $(format_duration "$total_elapsed"))"
  STEP_STARTED_AT=0
  STEP_LABEL=""
}

summary_label() {
  local label="$1"
  label="${label#== }"
  label="${label/ == (/ (}"
  label="${label% ==}"
  printf '%s\n' "$label"
}

join_details() {
  local joined=""
  local detail
  for detail in "$@"; do
    if [[ -z "$joined" ]]; then
      joined="$detail"
    else
      joined="${joined}; ${detail}"
    fi
  done
  printf '%s\n' "$joined"
}

# Run a step: print header, run command, track failure. Never exits.
ci_step() {
  local label="$1"
  shift
  begin_step "$label"
  set +e
  "$@"
  local status=$?
  set +e
  if [[ $status -ne 0 ]]; then
    finish_step_failed
    FAILED_STEPS+=("${label}")
    return 1
  fi
  finish_step
  return 0
}

stream_bot_progress() {
  local log_path="$1"
  tee "$log_path" | awk '
    /\] scenario begin / {
      line = $0
      sub(/^.*\] scenario begin /, "", line)
      current = line
      sub(/ .*/, "", current)
      printf "RUNNING: protocol bot scenario %s\n", current
      fflush()
    }
    /\] scenario done / {
      line = $0
      sub(/^.*\] scenario done /, "", line)
      scenario = line
      sub(/ .*/, "", scenario)
      elapsed = ""
      if (match(line, /elapsed=[^ ]+/)) {
        elapsed = substr(line, RSTART, RLENGTH)
      }
      if (elapsed != "") {
        printf "OK: protocol bot scenario %s (%s)\n", scenario, elapsed
      } else {
        printf "OK: protocol bot scenario %s\n", scenario
      }
      fflush()
    }
    /\] scenario failed / {
      line = $0
      sub(/^.*\] scenario failed /, "", line)
      scenario = line
      sub(/ .*/, "", scenario)
      elapsed = ""
      if (match(line, /elapsed=[^ ]+/)) {
        elapsed = substr(line, RSTART, RLENGTH)
      }
      if (elapsed != "") {
        printf "FAIL: protocol bot scenario %s (%s)\n", scenario, elapsed
      } else {
        printf "FAIL: protocol bot scenario %s\n", scenario
      }
      fflush()
    }
  '
}

protocol_bot_failure_detail() {
  local log_path="$1"
  awk '
    /\] scenario begin / {
      total++
    }
    /\] scenario failed / {
      line = $0
      sub(/^.*\] scenario failed /, "", line)
      scenario = line
      sub(/ .*/, "", scenario)
      if (!seen[scenario]++) {
        failed++
        if (scenarios != "") {
          scenarios = scenarios ", " scenario
        } else {
          scenarios = scenario
        }
      }
    }
    END {
      if (failed > 0 && total > 0) {
        printf "%d of %d failed: %s", failed, total, scenarios
      } else if (failed > 0) {
        printf "%d failed: %s", failed, scenarios
      } else {
        printf "protocol bot failed"
      }
    }
  ' "$log_path"
}

client_bot_failure_detail() {
  local log_path="$1"
  awk '
    /^RUNNING: client bot scenario / {
      total++
    }
    /^\[bot-client .*\] running scenario: / {
      total++
    }
    /^\[bot-client\] FAIL [^ ]+ --/ {
      scenario = $3
      if (!seen[scenario]++) {
        failed++
        if (scenarios != "") {
          scenarios = scenarios ", " scenario
        } else {
          scenarios = scenario
        }
      }
    }
    /^\[bot-client\] FAIL: [0-9]+ scenario file\(s\) failed validation/ {
      validation = $0
      sub(/^\[bot-client\] FAIL: /, "", validation)
    }
    /^\[bot-client\] FAIL: .*: (runner|world_id|client_steps)/ {
      validation = "scenario file validation failed"
    }
    END {
      if (failed > 0 && total > 0) {
        printf "%d of %d failed: %s", failed, total, scenarios
      } else if (failed > 0) {
        printf "%d failed: %s", failed, scenarios
      } else if (validation != "") {
        printf "%s", validation
      } else {
        printf "client bot runner failed"
      }
    }
  ' "$log_path"
}

client_smoke_failure_detail() {
  local log_path="$1"
  awk '
    /^FAILED: / {
      gate = $0
      sub(/^FAILED: /, "", gate)
      sub(/ \(.*$/, "", gate)
      if (!seen[gate]++) {
        if (gates != "") {
          gates = gates ", " gate
        } else {
          gates = gate
        }
      }
    }
    END {
      if (gates != "") {
        printf "GDScript gate(s) failed: %s", gates
      } else {
        printf "Godot headless smoke failed"
      }
    }
  ' "$log_path"
}

ci_captured_step() {
  local label="$1"
  local detail_parser="$2"
  shift 2
  local log_path status detail
  log_path="$(mktemp -t arpg-ci-step.XXXXXX.log)"

  begin_step "$label"
  set +e
  "$@" 2>&1 | tee "$log_path"
  status=${PIPESTATUS[0]}
  set +e

  if [[ "$status" -ne 0 ]]; then
    detail="$("$detail_parser" "$log_path")"
    finish_step_failed
    if [[ -n "$detail" ]]; then
      FAILED_STEPS+=("${label} (${detail})")
    else
      FAILED_STEPS+=("${label}")
    fi
    rm -f "$log_path"
    return 1
  fi

  rm -f "$log_path"
  finish_step
  return 0
}

# ── Maintainability ratchets (were a Makefile prereq) ────────────────────────

ci_step "== 1/11 file-size ratchet ==" \
  "$RUN_QUIET" --label "file-size-ratchet" -- ./scripts/check-file-size-ratchet.sh

ci_step "== 2/11 extraction coupling ratchet ==" \
  "$RUN_QUIET" --label "extraction-coupling-ratchet" -- \
    python3 ./scripts/check-extraction-coupling-ratchet.py

# ── Independent checks ────────────────────────────────────────────────────────

ci_step "== 3/11 shared schema validation ==" \
  "$RUN_QUIET" --label validate-shared -- make validate-shared

ci_step "== 4/11 asset manifest + GLB validation ==" \
  "$RUN_QUIET" --label validate-assets -- make validate-assets

ci_step "== 5/11 determinism lint ==" \
  "$RUN_QUIET" --label "determinism-lint" -- make lint-determinism

# DB-backed Go tests (store, http) use ARPG_DATABASE_URL, and fail rather than skip when it
# is set but unreachable (internal/testdb). Export it only when this checkout's test DB is
# ready now; otherwise they skip loudly (Postgres is started later, in step 8).
if "$ROOT/scripts/test_db.sh" ensure "$DATABASE_URL" >/dev/null 2>&1; then
  export ARPG_DATABASE_URL="$DATABASE_URL"
else
  unset ARPG_DATABASE_URL ARPG_TEST_DATABASE_URL
  echo "[ci] Postgres not reachable before step 6: DB-backed Go tests will SKIP (run make db-up first for full coverage)"
fi

# The race detector covers the concurrent realtime hub/session loop (~30s). internal/http
# exceeds the 10m test timeout under -race, so it is not included yet.
ci_step "== 6/11 Go fmt + tests + race + vet ==" \
  "$RUN_QUIET" --label "gofmt -l && go test ./... && go test -race ./internal/realtime/... && go vet ./..." -- \
  bash -c 'make fmt-check-go && cd server && go test ./... && go test -race ./internal/realtime/... && go vet ./...'

ci_step "== 7/11 Python unit checks ==" \
  bash -c "make tools >/dev/null && \"$RUN_QUIET\" --label 'pytest tools' -- \"$ROOT/.venv/bin/python\" -m pytest -q tools && \"$ROOT/.venv/bin/python\" -c 'from tools.bot.ci_pack import validate_ci_pack; validate_ci_pack()'"

# ── Server-dependent steps ────────────────────────────────────────────────────

start_server() {
  begin_step "== 8/11 start Postgres + server =="
  set +e

  make db-up && "$ROOT/scripts/test_db.sh" ensure "$DATABASE_URL"
  local db_status=$?
  if [[ $db_status -ne 0 ]]; then
    echo "FAILED: make db-up / test database"
    finish_step_failed
    set +e
    FAILED_STEPS+=("== 8/11 start Postgres + server ==")
    return 1
  fi

  SERVER_BIN="$(mktemp -t arpg-ci-server.XXXXXX)"
  "$RUN_QUIET" --label "go build arpg-server" -- bash -c \
    "cd server && go build -o \"$SERVER_BIN\" ./cmd/arpg-server"
  local build_status=$?
  if [[ $build_status -ne 0 ]]; then
    finish_step_failed
    set +e
    FAILED_STEPS+=("== 8/11 start Postgres + server ==")
    return 1
  fi

  if ! arpg_resolve_server_addr; then
    finish_step_failed
    FAILED_STEPS+=("== 8/11 start Postgres + server ==")
    return 1
  fi
  ARPG_DATABASE_URL="$DATABASE_URL" ARPG_ADDR="$ADDR" \
    ARPG_DEV_TOKEN="$DEV_TOKEN" ARPG_DEBUG_TOKEN="$DEBUG_TOKEN" \
    ARPG_GAMEPLAY_DEBUG="$GAMEPLAY_DEBUG" \
    ARPG_RULES_DIR="$ROOT/shared/rules" \
    "$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
  SERVER_PID=$!
  echo "server pid=$SERVER_PID addr=$ADDR db=$(arpg_db_name "$DATABASE_URL") (log: $SERVER_LOG); waiting for readiness..."
  if ! arpg_wait_own_server "ci" 90; then
    finish_step_failed
    set +e
    FAILED_STEPS+=("== 8/11 start Postgres + server ==")
    return 1
  fi

  set +e
  finish_step
  return 0
}

SERVER_AVAILABLE=0
if start_server; then
  SERVER_AVAILABLE=1
fi

if [[ "$SERVER_AVAILABLE" -eq 1 ]]; then
  # Step 9: protocol bot + replay
  begin_step "== 9/11 protocol bot + replay (SCENARIO=$CI_SCENARIO) =="
  BOT_LOG="$(mktemp -t arpg-ci-bot.XXXXXX.log)"
  step9_failed=0
  step9_failure_details=()
  set +e
  # Live wire-contract gate (v486): every received snapshot/state_delta payload must
  # validate against the v8 schemas. Benchmark probes below stay unvalidated (perf runs).
  SESSION_ID="$(ARPG_BOT_SCHEMA_VALIDATION="${ARPG_BOT_SCHEMA_VALIDATION:-strict}" \
    "$ROOT/.venv/bin/python" -m tools.bot.run \
    --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --debug-token "$DEBUG_TOKEN" \
    --scenario "$CI_SCENARIO" \
    --print-session-id 2> >(stream_bot_progress "$BOT_LOG" >&2))"
  bot_status=$?
  set +e
  if [[ "$bot_status" -ne 0 ]]; then
    echo "FAILED: protocol bot"
    show_log "$BOT_LOG" "protocol bot"
    bot_failure_detail="$(protocol_bot_failure_detail "$BOT_LOG")"
    rm -f "$BOT_LOG"
    step9_failed=1
    step9_failure_details+=("$bot_failure_detail")
  else
    if [[ "${ARPG_VERBOSE:-0}" == "1" ]]; then
      cat "$BOT_LOG"
    else
      echo "OK: protocol bot"
    fi
    rm -f "$BOT_LOG"
    echo "bot completed session: $SESSION_ID"
  fi

  if [[ -n "$SESSION_ID" ]]; then
    echo "RUNNING: arpg-replay session=$SESSION_ID"
    set +e
    "$RUN_QUIET" --label "arpg-replay" -- bash -c \
      "cd server && ARPG_DATABASE_URL=\"$DATABASE_URL\" ARPG_GAMEPLAY_DEBUG=\"$GAMEPLAY_DEBUG\" \
       go run ./cmd/arpg-replay --session-id \"$SESSION_ID\""
    replay_status=$?
    set +e
    if [[ $replay_status -ne 0 ]]; then
      step9_failed=1
      step9_failure_details+=("arpg-replay failed for session $SESSION_ID")
    else
      echo "OK: arpg-replay session=$SESSION_ID"
    fi
  else
    echo "SKIPPED: replay (protocol bot did not report a session id)"
    if [[ "$bot_status" -eq 0 ]]; then
      step9_failed=1
      step9_failure_details+=("replay skipped: protocol bot did not report a session id")
    fi
  fi

  # ci-full also drives the ci_tier=benchmark perf probes protocol-only (no
  # Godot observer), so they cannot rot outside `make benchmark`. Each probe's
  # declared max_elapsed_s is its budget. The probes also gate /state + replay
  # (benchmark_mixed_arena replays deterministically since v476 recorded load shed).
  if [[ "$CI_SCENARIO" == "all" ]]; then
    BENCH_LOG="$(mktemp -t arpg-ci-benchmark.XXXXXX.log)"
    echo "RUNNING: benchmark scenarios (protocol-only)"
    set +e
    "$ROOT/.venv/bin/python" -m tools.bot.run \
      --base-url "$BASE_URL" --dev-token "$DEV_TOKEN" --debug-token "$DEBUG_TOKEN" \
      --scenario benchmark --cleanup-characters \
      2>&1 >/dev/null | stream_bot_progress "$BENCH_LOG"
    bench_status=${PIPESTATUS[0]}
    set +e
    if [[ "$bench_status" -ne 0 ]]; then
      echo "FAILED: benchmark scenarios (protocol-only)"
      show_log "$BENCH_LOG" "benchmark scenarios"
      step9_failed=1
      step9_failure_details+=("benchmark: $(protocol_bot_failure_detail "$BENCH_LOG")")
    else
      echo "OK: benchmark scenarios (protocol-only)"
    fi
    rm -f "$BENCH_LOG"
  fi

  if [[ "$step9_failed" -ne 0 ]]; then
    finish_step_failed
    if [[ "${#step9_failure_details[@]}" -gt 0 ]]; then
      FAILED_STEPS+=("== 9/11 protocol bot + replay == ($(join_details "${step9_failure_details[@]}"))")
    else
      FAILED_STEPS+=("== 9/11 protocol bot + replay ==")
    fi
  else
    finish_step
  fi

  # Steps 10-11: Godot
  ci_captured_step "== 10/11 Godot client bot scenarios ==" client_bot_failure_detail \
    env GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" \
      SCENARIO="$CI_SCENARIO" HEADLESS=1 ./scripts/bot_client.sh

  ci_captured_step "== 11/11 Godot headless smoke ==" client_smoke_failure_detail \
    env GODOT="${GODOT:-godot}" BASE_URL="$BASE_URL" DEV_TOKEN="$DEV_TOKEN" \
      DEBUG_TOKEN="$DEBUG_TOKEN" ./scripts/client_smoke.sh
else
  echo
  echo "SKIPPED: steps 9-11 require a running server (step 8 failed)"
  FAILED_STEPS+=("== 9/11 protocol bot + replay == (skipped: server failed)")
  FAILED_STEPS+=("== 10/11 Godot client bot scenarios == (skipped: server failed)")
  FAILED_STEPS+=("== 11/11 Godot headless smoke == (skipped: server failed)")
fi

# ── Summary ───────────────────────────────────────────────────────────────────

total_elapsed=$(( $(date +%s) - CI_STARTED_AT ))
echo

if [[ ${#FAILED_STEPS[@]} -gt 0 ]]; then
  echo "CI FAILED in $(format_duration $total_elapsed) — ${#FAILED_STEPS[@]} step(s) failed:"
  for step in "${FAILED_STEPS[@]}"; do
    echo "  ✗  $(summary_label "$step")"
  done
  echo "(scroll up for each step's output)"
  exit 1
fi

echo "CI OK in $(format_duration $total_elapsed)"
