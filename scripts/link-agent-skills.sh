#!/usr/bin/env bash
# Link repo skills into Codex user skill dir (~/.codex/skills).
# Cursor and Claude Code use committed symlinks under .cursor/ and .claude/.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CODEX_SKILLS="${CODEX_HOME:-$HOME/.codex}/skills"

mkdir -p "$CODEX_SKILLS"

skill_names=()
while IFS= read -r skill_file; do
  skill_names+=("$(basename "$(dirname "$skill_file")")")
done < <(find "$REPO_ROOT/skills" -mindepth 2 -maxdepth 2 -name SKILL.md -print | sort)

if [[ ${#skill_names[@]} -eq 0 ]]; then
  echo "error: no skills found under $REPO_ROOT/skills" >&2
  exit 1
fi

for name in "${skill_names[@]}"; do
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
printf 'Invoke with '
printf '$%s' "${skill_names[0]}"
for name in "${skill_names[@]:1}"; do
  printf ' / $%s' "$name"
done
echo " or ask per AGENTS.md slash commands."
