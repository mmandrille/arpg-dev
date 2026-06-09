# Agent entrypoint

Read these **before** specs, plans, or code:

1. [`CLAUDE.md`](CLAUDE.md) — commands, architecture, invariants, SDD process.
2. [`docs/PROGRESS.md`](docs/PROGRESS.md) — current slice baseline and open gaps.
3. [`docs/researchs/godot-plugins-and-shortcuts.md`](docs/researchs/godot-plugins-and-shortcuts.md) — **check for existing Godot plugins, demos, and asset packs** before building new client UI, inventory presentation, isometric/camera tooling, or placeholder art from scratch.

When starting client-side work, run the adoption checklist in the plugins doc and record *adopt / borrow / reject* in the slice plan.

## Slash commands (cross-agent skills)

Canonical definitions live in [`skills/`](skills/README.md). Tool paths are symlinks to the same files.

| Command | Skill | What it does |
|---------|-------|--------------|
| `/next {optional idea}` | [`skills/next/SKILL.md`](skills/next/SKILL.md) | Read `PROGRESS.md` + ADRs → propose next slice options or evaluate your idea → spec-ready brief (complexity, requirements, doubts) |
| `/spec {brief_or_idea}` | [`skills/spec/SKILL.md`](skills/spec/SKILL.md) | Turn an approved brief or idea into `docs/specs/vN_spec-<codename>.md` without implementing |
| `/plan {spec_file.md}` | [`skills/plan/SKILL.md`](skills/plan/SKILL.md) | Review spec for gaps → ask questions → write `docs/plans/vN_<date>-<codename>.md` (includes bot scenarios when gameplay/protocol is in scope) |
| `/execute {plan_file.md}` | [`skills/execute/SKILL.md`](skills/execute/SKILL.md) | Review plan for gaps → ask questions → implement task-by-task until `make ci` is green |
| `/finish` | [`skills/finish/SKILL.md`](skills/finish/SKILL.md) | Consolidate `PROGRESS.md` + uncommitted changes → `make ci` green → commit `feat: v{N}: {title}` |
| `/showme {gear\|inventory\|...}` | [`skills/showme/SKILL.md`](skills/showme/SKILL.md) | Open or capture a focused Godot client preview for fast visual feedback |

Workflow: `/next` → `/spec` → `/plan` → `/execute` → `/finish`. Use `/showme` during client visual work when fast focused feedback is useful. Do not skip the review gates.

### Per-agent setup

| Agent | Discovery | Invoke |
|-------|-----------|--------|
| **Cursor** | `.cursor/skills/` → `skills/` (committed symlink) | `/next`, `/spec`, `/plan`, `/execute`, `/finish`, `/showme` |
| **Claude Code** | `.claude/skills/` → `skills/` (committed symlink) | same; `/reload-skills` after pull |
| **Codex** | `skills/` in repo + run [`scripts/link-agent-skills.sh`](scripts/link-agent-skills.sh) once for `~/.codex/skills/` | `$next`, `$spec`, `$plan`, `$execute`, `$finish`, `$showme` |

Edit skills only under `skills/` — never duplicate into `.cursor/` or `.claude/`.

## Git workflow

Do **not** create new branches. Work only on the branch already checked out — even if it is `main`. If a feature branch is needed, the user creates and checks it out before development begins.

## Development priority

While the game is still in active development, do **not** preserve backward compatibility just for its own sake. Prefer the cleanest, healthiest implementation and update contracts, fixtures, tests, tools, and docs together.
