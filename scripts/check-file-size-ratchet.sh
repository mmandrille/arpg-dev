#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASELINE="${ROOT}/.maintainability/file-size-baseline.tsv"
MAX_LINES="${MAX_LINES:-600}"
GROWTH_ALLOWANCE="${GROWTH_ALLOWANCE:-25}"

if [[ ! -f "${BASELINE}" ]]; then
  echo "missing file-size baseline: ${BASELINE}" >&2
  exit 1
fi

is_source_file() {
  local path="$1"
  case "${path}" in
    *.go|*.gd|*.py|*.sh|*.mk|Makefile) return 0 ;;
    *) return 1 ;;
  esac
}

is_exempt_path() {
  local path="$1"
  case "${path}" in
    .git/*|.venv/*|docs/*|shared/golden/*|client/.godot/*|client/imports/*) return 0 ;;
    *) return 1 ;;
  esac
}

baseline_for() {
  local path="$1"
  awk -F '\t' -v wanted="${path}" '($1 == wanted) { print $2; found = 1; exit } END { if (!found) exit 1 }' "${BASELINE}"
}

failures_file="$(mktemp)"
trap 'rm -f "${failures_file}"' EXIT

while IFS= read -r path; do
  [[ -z "${path}" ]] && continue
  is_exempt_path "${path}" && continue
  is_source_file "${path}" || continue

  full_path="${ROOT}/${path}"
  [[ -f "${full_path}" ]] || continue

  line_count="$(wc -l < "${full_path}" | tr -d ' ')"
  baseline_count="$(baseline_for "${path}" || true)"

  if [[ -n "${baseline_count}" ]]; then
    allowed=$((baseline_count + GROWTH_ALLOWANCE))
    if (( line_count > allowed )); then
      printf '%s\n' "${path}: ${line_count} lines exceeds grandfathered baseline ${baseline_count} + allowance ${GROWTH_ALLOWANCE}. Split code out or update the baseline with a documented maintenance exception." >> "${failures_file}"
    fi
  elif (( line_count > MAX_LINES )); then
    printf '%s\n' "${path}: ${line_count} lines exceeds new-file target ${MAX_LINES}. Split this file before committing." >> "${failures_file}"
  fi
done < <(cd "${ROOT}" && git ls-files)

if [[ -s "${failures_file}" ]]; then
  echo "File size ratchet failed:" >&2
  sed 's/^/  - /' "${failures_file}" >&2
  echo "" >&2
  echo "Rule: new source/test/tool files stay at or below ${MAX_LINES} lines; grandfathered files may not grow by more than ${GROWTH_ALLOWANCE} lines without an explicit documented exception." >&2
  exit 1
fi

echo "file-size ratchet passed"
