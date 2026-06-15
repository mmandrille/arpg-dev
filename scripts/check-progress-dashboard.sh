#!/usr/bin/env bash
# Fail CI when PROGRESS.md grows back into a monolithic changelog/dashboard hybrid.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROGRESS="${ROOT}/PROGRESS.md"
MAX_LINES=250

if [[ ! -f "${PROGRESS}" ]]; then
  echo "check-progress-dashboard: missing ${PROGRESS}" >&2
  exit 1
fi

line_count="$(wc -l < "${PROGRESS}" | tr -d ' ')"
if (( line_count > MAX_LINES )); then
  echo "check-progress-dashboard: PROGRESS.md has ${line_count} lines (max ${MAX_LINES})." >&2
  echo "Move history to docs/progress/ and keep PROGRESS.md as a dashboard." >&2
  exit 1
fi

if grep -q '^### Recently closed' "${PROGRESS}"; then
  echo "check-progress-dashboard: PROGRESS.md must not contain '### Recently closed'." >&2
  echo "Write shipped proof to docs/as-built/ instead." >&2
  exit 1
fi

required=(
  "docs/progress/slice-lifecycle.md"
  "docs/progress/slice-codename-index.md"
  "docs/progress/scenario-catalog.md"
)
for path in "${required[@]}"; do
  if [[ ! -f "${ROOT}/${path}" ]]; then
    echo "check-progress-dashboard: missing required archive ${path}" >&2
    exit 1
  fi
done

echo "check-progress-dashboard: ok (${line_count}/${MAX_LINES} lines)"
