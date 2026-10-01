# Project agent skills

Canonical skill definitions for this repo. **Edit files here only** — tool-specific paths are symlinks.

| Skill | Purpose |
|-------|---------|
| [`next/`](next/SKILL.md) | `/next {idea?}` → offer one slice or a dependency-aware batch for acceptance |
| [`spec/`](spec/SKILL.md) | `/spec {brief_or_idea}` → draft `docs/specs/vN_spec-*.md` |
| [`plan/`](plan/SKILL.md) | `/plan {spec}` → review spec, write `docs/plans/` |
| [`execute/`](execute/SKILL.md) | `/execute {plan}` → implement one slice with focused checks in batch mode |
| [`finish/`](finish/SKILL.md) | `/finish` → coordinator closes integrated batch with CI and commits; standalone closeout remains supported |
| [`review/`](review/SKILL.md) | `/review` or `$review` → write repo-wide engineering review docs |
| [`showme/`](showme/SKILL.md) | `/showme` or `$showme` → focused screenshot/live preview; `--refresh` hot-reloads gear JSON while tuning; `make regen-screenshots` for batch visual regression |
| [`3dmodel/`](3dmodel/SKILL.md) | `$3dmodel` → integrate supplied GLB/glTF models into the Godot client |
| [`autoloop/`](autoloop/SKILL.md) | `$autoloop` → coordinate accepted slice sessions, integration, final CI, review, and refactor |
| [`refactor/`](refactor/SKILL.md) | `$refactor` → scorecard-driven minor cleanup commits after a fresh review |

For a batch, read the [shared batch workflow](autoloop/references/batch-workflow.md). The main
chat dispatches one user-visible detached-worktree session per accepted slice. Slice sessions
spec, plan, implement, and run focused checks; the main chat integrates ready slices, runs the
combined `make ci`, closes the batch, then runs `$review` and `$refactor`. `/execute` is the
implementation command; “implement” in a request refers to that skill.

## Discovery paths

| Agent | Project path | How to invoke |
|-------|--------------|---------------|
| **Cursor** | `.cursor/skills/*` → `skills/` | `/next`, `/spec`, `/plan`, `/execute`, `/finish`, `/review`, `/showme`, `/autoloop`, `/refactor` |
| **Claude Code** | `.claude/skills/*` → `skills/` | same slash commands |
| **Codex** | `skills/` (repo) + optional `~/.codex/skills/` symlink | `$next`, `$spec`, `$plan`, `$execute`, `$finish`, `$review`, `$showme`, `$3dmodel`, `$autoloop`, `$refactor` |

Run once per machine for Codex user-level discovery:

```bash
./scripts/link-agent-skills.sh
```

Then restart Codex (or run `/reload-skills` in Claude Code after pulling).

The linker discovers every `skills/*/SKILL.md` directory, so newly added repo
skills are picked up without editing a fixed allowlist.
