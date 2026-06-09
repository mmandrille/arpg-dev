#!/usr/bin/env bash
# Run a command with agent-friendly output.
# Success (non-verbose): one OK line. Failure (non-verbose): last N log lines.
# Full output: ARPG_VERBOSE=1, or pass --verbose / -v.
set -euo pipefail

TAIL_LINES="${ARPG_QUIET_TAIL_LINES:-100}"
VERBOSE="${ARPG_VERBOSE:-0}"
LABEL=""

usage() {
  echo "usage: run_quiet.sh [--verbose|-v] [--label NAME] -- COMMAND [ARGS...]" >&2
  exit 2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -v|--verbose)
      VERBOSE=1
      shift
      ;;
    --label)
      [[ $# -ge 2 ]] || usage
      LABEL="$2"
      shift 2
      ;;
    --)
      shift
      break
      ;;
    -h|--help)
      usage
      ;;
    *)
      break
      ;;
  esac
done

[[ $# -gt 0 ]] || usage

if [[ -z "$LABEL" ]]; then
  LABEL="$*"
fi

tmp="$(mktemp -t arpg-quiet.XXXXXX.log)"
cleanup() { rm -f "$tmp"; }
trap cleanup EXIT

set +e
"$@" >"$tmp" 2>&1
status=$?
set -e

if [[ "$status" -eq 0 ]]; then
  if [[ "$VERBOSE" == "1" ]]; then
    cat "$tmp"
  else
    printf 'OK: %s\n' "$LABEL"
  fi

  exit 0
fi

if [[ "$VERBOSE" == "1" ]]; then
  cat "$tmp"
else
  printf 'FAILED: %s\n' "$LABEL"
  printf -- '--- last %s lines ---\n' "$TAIL_LINES"
  tail -n "$TAIL_LINES" "$tmp"
  printf '%s\n' '--- rerun with VERBOSE=1 (or V=1) for full output ---'
fi

exit "$status"
