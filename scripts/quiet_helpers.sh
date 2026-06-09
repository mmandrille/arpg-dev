#!/usr/bin/env bash
# Shared helpers for agent-friendly quiet output. Source from other scripts.
if [[ -z "${ROOT:-}" ]]; then
  ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

RUN_QUIET="${RUN_QUIET:-$ROOT/scripts/run_quiet.sh}"
TAIL_LINES="${ARPG_QUIET_TAIL_LINES:-100}"

show_log() {
  local log="$1"
  local label="${2:-log}"

  if [[ "${ARPG_VERBOSE:-0}" == "1" ]]; then
    cat "$log"
  else
    echo "--- last ${TAIL_LINES} lines of ${label} ---"
    tail -n "$TAIL_LINES" "$log"
    echo "--- rerun with VERBOSE=1 for full output ---"
  fi
}

is_quiet_mode() {
  [[ "${ARPG_VERBOSE:-0}" != "1" ]]
}
