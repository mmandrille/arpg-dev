# Agent entrypoint

Read these **before** specs, plans, or code:

1. [`PROGRESS.md`](PROGRESS.md) — **start here** for where the project stands: read **Current status**,
   **Open gaps**, and **Agent checklist** (not the full file unless needed). Slice history lives in
   [`docs/progress/slice-lifecycle.md`](docs/progress/slice-lifecycle.md). Per-slice proof lives in
   [`docs/as-built/`](docs/as-built/). Do not rely on stale slice numbers in other docs —
   `PROGRESS.md` **Current status** is the canonical baseline.
2. [`CLAUDE.md`](CLAUDE.md) — commands, architecture, invariants, SDD process.
3. [`docs/CODEMAP.md`](docs/CODEMAP.md) — domain → files index. Use it to decide which files to load before grepping broad coordinators.
4. For client UI, inventory presentation, isometric/camera tooling, or placeholder art, first check existing in-repo Godot scripts, scenes, demos, and asset manifests before introducing new dependencies or asset pipelines.

When starting client-side work that could use outside assets or plugins, record an *adopt / borrow / reject* decision in the slice spec or plan. If external adoption needs deeper research, add or update a focused note under `docs/researchs/` as part of that planning work.

## Timestamped task updates

During any multi-step task, prefix every intermediary user-facing progress update with the local
time in `HH:MM:SS` format:

```text
[HH:MM:SS] <message>
```

Apply this to progress/update messages printed while working, not final answers, code blocks, file
contents, commit messages, quoted output, or tool output summaries.

## Slash commands (cross-agent skills)

Canonical definitions live in [`skills/`](skills/README.md). Tool paths are symlinks to the same files.

| Command | Skill | What it does |
|---------|-------|--------------|
| `/next {optional idea}` | [`skills/next/SKILL.md`](skills/next/SKILL.md) | Offer one slice or a dependency-aware batch of spec-ready slices for acceptance |
| `/spec {brief_or_idea}` | [`skills/spec/SKILL.md`](skills/spec/SKILL.md) | Turn an approved brief or idea into `docs/specs/vN_spec-<codename>.md` without implementing |
| `/plan {spec_file.md}` | [`skills/plan/SKILL.md`](skills/plan/SKILL.md) | Review spec for gaps → ask questions → write `docs/plans/vN_<date>-<codename>.md` (includes bot scenarios when gameplay/protocol is in scope) |
| `/execute {plan_file.md}` | [`skills/execute/SKILL.md`](skills/execute/SKILL.md) | Implement one approved plan; focused verification in a batch slice session |
| `/finish` | [`skills/finish/SKILL.md`](skills/finish/SKILL.md) | Coordinator closes an integrated batch with combined CI, lifecycle docs, and commits; standalone slices retain their own gate |
| `/review {vN?}` | [`skills/review/SKILL.md`](skills/review/SKILL.md) | Analyze the full repo → write overview at `docs/reviews/YYYYMMDD_vN-overview.md` plus companion reports under `docs/reviews/{backend,client,extras}/` |
| `/showme {gear\|inventory\|...}` | [`skills/showme/SKILL.md`](skills/showme/SKILL.md) | Capture one focused Godot visual, or run `make regen-screenshots` for batch regression |
| `$3dmodel {model task}` | [`skills/3dmodel/SKILL.md`](skills/3dmodel/SKILL.md) | Integrate supplied GLB/glTF models into the Godot client presentation path |
| `/autoloop` | [`skills/autoloop/SKILL.md`](skills/autoloop/SKILL.md) | Main chat offers a batch, dispatches one detached-worktree session per accepted slice, integrates ready work, then runs CI, review, and refactor |
| `/refactor` | [`skills/refactor/SKILL.md`](skills/refactor/SKILL.md) | Read the latest review scorecard → make small verified cleanup commits until scorecard areas are 9+ or only major work remains |

Batch workflow: main chat runs `/next` and waits for acceptance; each accepted slice session runs `/spec` → `/plan` → `/execute` with focused checks in its own detached worktree. Main integrates ready slices as their dependencies permit, runs combined `make ci` after all are integrated, closes them with `/finish`, cleans up batch worktrees, then runs `/review` → `/refactor` on the new baseline. See [`skills/autoloop/references/batch-workflow.md`](skills/autoloop/references/batch-workflow.md). A directly requested standalone slice still uses `/next` → `/spec` → `/plan` → `/execute` → `/finish`. Use `/showme` during visual work. Do not skip spec, plan, or review gates.

### Per-agent setup

| Agent | Discovery | Invoke |
|-------|-----------|--------|
| **Cursor** | `.cursor/skills/` → `skills/` (committed symlink) | `/next`, `/spec`, `/plan`, `/execute`, `/finish`, `/review`, `/showme`, `/autoloop`, `/refactor` |
| **Claude Code** | `.claude/skills/` → `skills/` (committed symlink) | same; `/reload-skills` after pull |
| **Codex** | `skills/` in repo + run [`scripts/link-agent-skills.sh`](scripts/link-agent-skills.sh) once for `~/.codex/skills/` | `$next`, `$spec`, `$plan`, `$execute`, `$finish`, `$review`, `$showme`, `$3dmodel`, `$autoloop`, `$refactor` |

Edit skills only under `skills/` — never duplicate into `.cursor/` or `.claude/`.

## Git workflow

Do **not** create new branches. The coordinator works on the already checked-out branch (normally `main`); batch slice sessions use detached worktrees based on the coordinator's recorded commit. If a feature branch is needed, the user creates and checks it out before development begins.

### Worktree isolation

An accepted slice batch authorizes one temporary **detached** worktree and one user-visible session per slice. Standalone work may use a user-provided worktree or a separately approved detached worktree. Do not check out the coordinator's branch in a second worktree or create a branch on an agent's initiative.

When worktree isolation is used:

1. Keep spec, plan, implementation, exploratory edits, and focused test iterations inside the assigned worktree.
2. Do not run `make ci`, commit, push, or modify `main` from a batch slice session. Report the complete changed/untracked file list and ignored evidence to the coordinator.
3. The coordinator integrates each ready slice into `main` after its dependencies, resolves overlapping files, and runs focused integration checks. Compare every child's changed path and intended behavior with the integrated result before cleanup.
4. After all accepted slices are integrated, the coordinator runs the combined `make ci`, fixes and reruns failures, performs `/finish` closeout and commits, preserves referenced evidence, and removes only batch-created worktrees. Then run `/review` and `/refactor`.

## Testing discipline

Prefer targeted verification while iterating. Run the smallest command or scenario that covers the files and behavior you changed, such as a focused Go package test, `make validate-shared`, `make client-unit`, a single `make bot scenario=...`, or one client bot scenario.

When working on features or changes that involve visual effects and a client bot scenario exists, always tell the user the exact scenario name and command they can run for visual verification, for example: `make bot-visual scenario=blablabla`.

**Batch visual regression:** after presentation changes (gear fit, skeleton sockets, item/skill
icons, floor loot models, class bodies), run `make regen-screenshots` with the relevant
`SUITE=` filter and inspect PNGs under `.artifacts/screenshots/latest/`. Use
`make regen-screenshots-list` to see suites; `DRY_RUN=1` to preview jobs without Godot.
Single-element captures still use `/showme` or `python3 skills/showme/scripts/render_focus.py`.
Full catalog: [`skills/showme/SKILL.md`](skills/showme/SKILL.md).

Do **not** run `make ci` in every batch slice session. Reserve it for the combined integrated state after all accepted slices are in `main`; rerun after fixes until the final state is green. Standalone work may run it for its final pre-commit gate when the change warrants it. A post-review refactor that changes code beyond focused coverage may require one additional final gate.

**CI pack curation:** `make ci` runs only `tools/bot/ci_pack.json`. New scenarios default to
`"ci_tier": "extended"`; add to the pack only for merge-blocking coverage not already gated elsewhere,
and keep pack size stable by demoting redundant or slow scenarios when promoting new ones. See
`.cursor/rules/ci-pack-maintenance.mdc`.

## Data-driven gameplay and tests

Gameplay and presentation tuning belongs in shared data, not hardcoded implementation constants. Before adding or changing any tunable value, check whether it can live in `shared/rules/main_config.v0.json`, another `shared/rules/*.json` catalog, or a schema-backed content/asset file. Values such as attack speed, movement speed, animation duration/speed/height/radius, drop chance, loot weights, monster stats, class stats, skill costs/cooldowns, shop pricing, XP curves, and generated-content budgets must be configurable unless the slice explicitly documents why code ownership is required. Never write a literal tuning value into code when an existing or new data-driven field can own it.

Tests must preserve that configurability. Do not copy current tuning values into unrelated assertions. Prefer rule-derived expectations, semantic/range/eventual assertions, or focused temp-rule fixtures that change only the relevant shared JSON and prove gameplay follows it. Exact numeric assertions are acceptable only for protocol/schema contracts, deterministic goldens, evaluator parity, or a test whose stated purpose is to own that formula.

## Development priority

While the game is still in active development, do **not** preserve backward compatibility just for its own sake. Prefer the cleanest, healthiest implementation and update contracts, fixtures, tests, tools, and docs together.
