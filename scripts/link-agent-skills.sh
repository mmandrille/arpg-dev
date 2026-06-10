#!/usr/bin/env bash
# Link repo skills into Codex user skill dir (~/.codex/skills).
# Cursor and Claude Code use committed symlinks under .cursor/ and .claude/.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CODEX_SKILLS="${CODEX_HOME:-$HOME/.codex}/skills"

mkdir -p "$CODEX_SKILLS"

for name in next spec plan execute finish showme autoloop; do
  target="$CODEX_SKILLS/$name"
  source="$REPO_ROOT/skills/$name"
  if [[ ! -d "$source" ]]; then
    echo "error: missing $source" >&2
    exit 1
  fi
  ln -sfn "$source" "$target"
  echo "linked $target -> $source"
done

echo ""
echo "Codex: restart the session (or start a new one) to pick up skills."
echo "Invoke with \$next / \$spec / \$plan / \$execute / \$finish / \$showme / \$autoloop or ask per AGENTS.md slash commands."
