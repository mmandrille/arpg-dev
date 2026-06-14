# Agent entrypoint

Read these **before** specs, plans, or code:

1. [`PROGRESS.md`](PROGRESS.md) — **start here** for where the project stands: latest completed slice, active branch, open gaps, deferred backlog, engineering-review cadence, and the agent checklist. Do not rely on stale slice numbers in other docs — `PROGRESS.md` is the canonical baseline. When **Next engineering review** is due, read [`docs/reviews/`](docs/reviews/) and write a fresh review set before the next batch of slices.
2. [`CLAUDE.md`](CLAUDE.md) — commands, architecture, invariants, SDD process.
3. [`docs/CODEMAP.md`](docs/CODEMAP.md) — domain → files index. Use it to decide which files to load before grepping broad coordinators.
4. [`docs/researchs/godot-plugins-and-shortcuts.md`](docs/researchs/godot-plugins-and-shortcuts.md) — **check for existing Godot plugins, demos, and asset packs** before building new client UI, inventory presentation, isometric/camera tooling, or placeholder art from scratch.

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
| `/review {vN?}` | [`skills/review/SKILL.md`](skills/review/SKILL.md) | Analyze the full repo → write overview at `docs/reviews/YYYYMMDD_vN-overview.md` plus companion reports under `docs/reviews/{backend,client,extras}/` |
| `/showme {gear\|inventory\|...}` | [`skills/showme/SKILL.md`](skills/showme/SKILL.md) | Open or capture a focused Godot client preview for fast visual feedback |
| `/autoloop {count}` | [`skills/autoloop/SKILL.md`](skills/autoloop/SKILL.md) | Repeat `/next` → `/spec` → `/plan` → `/execute` → `/finish` for the requested number of autonomous committed slices, then run any due engineering review |

Workflow: `/next` → `/spec` → `/plan` → `/execute` → `/finish`. Use `/review` when the engineering-review cadence is due, and `/showme` during client visual work when fast focused feedback is useful. Do not skip the review gates.

### Per-agent setup

| Agent | Discovery | Invoke |
|-------|-----------|--------|
| **Cursor** | `.cursor/skills/` → `skills/` (committed symlink) | `/next`, `/spec`, `/plan`, `/execute`, `/finish`, `/review`, `/showme`, `/autoloop` |
| **Claude Code** | `.claude/skills/` → `skills/` (committed symlink) | same; `/reload-skills` after pull |
| **Codex** | `skills/` in repo + run [`scripts/link-agent-skills.sh`](scripts/link-agent-skills.sh) once for `~/.codex/skills/` | `$next`, `$spec`, `$plan`, `$execute`, `$finish`, `$review`, `$showme`, `$autoloop` |

Edit skills only under `skills/` — never duplicate into `.cursor/` or `.claude/`.

## Git workflow

Do **not** create new branches. Work only on the branch already checked out — even if it is `main`. If a feature branch is needed, the user creates and checks it out before development begins.

### Worktree isolation

Prefer isolating agent implementation work in a separate Git worktree until the slice has passed its targeted verification. This is valid when the user has already provided a worktree/branch for the task, or when the user explicitly approves creating a temporary worktree and branch. Agents must not create that branch unprompted, because Git cannot check out the same branch in two worktrees and this repo's default rule is still "no new branches."

When worktree isolation is used:

1. Keep exploratory edits, generated files, and focused test iterations inside the isolated worktree.
2. Do not commit from the isolated worktree unless the user explicitly asks.
3. After verification, transfer the complete tested change set back to `main` and run `/finish` there.
4. Let `/finish` perform the final `PROGRESS.md` consolidation, `make ci` gate, staging, and single `feat: vN: ...` commit on `main`.

## Testing discipline

Prefer targeted verification while iterating. Run the smallest command or scenario that covers the files and behavior you changed, such as a focused Go package test, `make validate-shared`, `make client-unit`, a single `make bot scenario=...`, or one client bot scenario.

When working on features or changes that involve visual effects and a client bot scenario exists, always tell the user the exact scenario name and command they can run for visual verification, for example: `make bot-visual scenario=blablabla`.

Do **not** repeatedly run the full suite by default. Reserve `make ci` for the final pre-commit proof when the change is broad enough to justify it, when targeted tests leave meaningful integration risk, or when the user explicitly asks for full CI.

## Data-driven gameplay and tests

Gameplay tuning belongs in shared data, not hardcoded implementation constants. Before adding or changing balance-sensitive behavior, check whether the value should live in `shared/rules/main_config.v0.json`, another `shared/rules/*.json` catalog, or a schema-backed content file. Values such as attack speed, movement speed, drop chance, loot weights, monster stats, class stats, skill costs/cooldowns, shop pricing, XP curves, and generated-content budgets must be configurable unless the slice explicitly documents why code ownership is required.

Tests must preserve that configurability. Do not copy current tuning values into unrelated assertions. Prefer rule-derived expectations, semantic/range/eventual assertions, or focused temp-rule fixtures that change only the relevant shared JSON and prove gameplay follows it. Exact numeric assertions are acceptable only for protocol/schema contracts, deterministic goldens, evaluator parity, or a test whose stated purpose is to own that formula.

## Development priority

While the game is still in active development, do **not** preserve backward compatibility just for its own sake. Prefer the cleanest, healthiest implementation and update contracts, fixtures, tests, tools, and docs together.
