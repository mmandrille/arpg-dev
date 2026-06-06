# Project agent skills

Canonical skill definitions for this repo. **Edit files here only** — tool-specific paths are symlinks.

| Skill | Purpose |
|-------|---------|
| [`next/`](next/SKILL.md) | `/next {idea?}` → propose next slice, spec-ready brief |
| [`plan/`](plan/SKILL.md) | `/plan {spec}` → review spec, write `docs/plans/` |
| [`execute/`](execute/SKILL.md) | `/execute {plan}` → implement until `make ci` green |
| [`finish/`](finish/SKILL.md) | `/finish` → consolidate PROGRESS, CI, `feat: vN:` commit |

## Discovery paths

| Agent | Project path | How to invoke |
|-------|--------------|---------------|
| **Cursor** | `.cursor/skills/*` → `skills/` | `/next`, `/plan`, `/execute`, `/finish` |
| **Claude Code** | `.claude/skills/*` → `skills/` | same slash commands |
| **Codex** | `skills/` (repo) + optional `~/.codex/skills/` symlink | `$next`, `$plan`, `$execute`, `$finish` |

Run once per machine for Codex user-level discovery:

```bash
./scripts/link-agent-skills.sh
```

Then restart Codex (or run `/reload-skills` in Claude Code after pulling).
