# Project agent skills

Canonical skill definitions for this repo. **Edit files here only** — tool-specific paths are symlinks.

| Skill | Purpose |
|-------|---------|
| [`next/`](next/SKILL.md) | `/next {idea?}` → propose next slice, spec-ready brief |
| [`spec/`](spec/SKILL.md) | `/spec {brief_or_idea}` → draft `docs/specs/vN_spec-*.md` |
| [`plan/`](plan/SKILL.md) | `/plan {spec}` → review spec, write `docs/plans/` |
| [`execute/`](execute/SKILL.md) | `/execute {plan}` → implement until `make ci` green |
| [`finish/`](finish/SKILL.md) | `/finish` → consolidate PROGRESS, CI, `feat: vN:` commit |
| [`review/`](review/SKILL.md) | `/review` or `$review` → write repo-wide engineering review docs |
| [`showme/`](showme/SKILL.md) | `/showme` or `$showme` → focused Godot screenshot/live preview for visual tuning |
| [`autoloop/`](autoloop/SKILL.md) | `$autoloop {count}` → repeat next/spec/plan/execute/finish for up to 3 committed slices |

## Discovery paths

| Agent | Project path | How to invoke |
|-------|--------------|---------------|
| **Cursor** | `.cursor/skills/*` → `skills/` | `/next`, `/spec`, `/plan`, `/execute`, `/finish`, `/review`, `/showme`, `/autoloop` |
| **Claude Code** | `.claude/skills/*` → `skills/` | same slash commands |
| **Codex** | `skills/` (repo) + optional `~/.codex/skills/` symlink | `$next`, `$spec`, `$plan`, `$execute`, `$finish`, `$review`, `$showme`, `$autoloop` |

Run once per machine for Codex user-level discovery:

```bash
./scripts/link-agent-skills.sh
```

Then restart Codex (or run `/reload-skills` in Claude Code after pulling).
