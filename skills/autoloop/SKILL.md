---
name: autoloop
description: >-
  Curate and execute autonomous feature/gameplay SDD slices only. Either use
  the user-provided slice idea(s) directly or, when no idea is provided, present
  a curated menu of 15 player-visible feature, game-improvement, or addition
  candidates. Batch blocking clarification questions across the selected queue,
  then execute an autonomous loop through next, spec, plan, execute, and finish
  with committed slices. Use when the user runs $autoloop N, /autoloop N, or asks
  Codex to curate and run multiple feature slices.
disable-model-invocation: true
---

# $autoloop — Curated Autonomous Feature Slice Loop

**Trigger:** `$autoloop {count}` or `/autoloop {count}`

Examples:

- `$autoloop 1` — present 15 ideas, then complete one selected slice after the user picks.
- `$autoloop 1 add a town healer` — treat the inline idea as selected and complete one slice if gates pass.
- `$autoloop 5` — present 15 ideas, then complete five selected slices, stopping early only on a hard stop.
- `$autoloop 12 idea A; idea B; ...` — treat the inline ideas as selected and complete up to twelve viable slices.

**Announce at start:** "Using the **autoloop** skill to curate feature/gameplay slice ideas, then run an autonomous SDD loop after you choose."

## Purpose

Run the normal SDD workflow repeatedly for feature/gameplay progress only. An
autoloop slice should usually be watchable end-to-end and treat backend,
shared contracts, bot proof, and client presentation as one vertical slice when
the behavior reaches the player.

If the user supplied idea text in the initial command, that text is the
selection gate. If the command has only a count, present a menu and wait for the
user to select:

```text
$autoloop with idea -> order inline feature idea(s) -> batch questions -> $spec -> $plan -> $execute -> $finish
$autoloop without idea -> feature idea menu -> user picks -> order picks -> batch questions -> $spec -> $plan -> $execute -> $finish
```

The initial `$autoloop` invocation authorizes preflight and idea discovery only
when no idea text is provided. When idea text is provided after the count, that
same invocation authorizes using the provided idea(s) as the selected queue,
up to the requested slice count.
It does **not** authorize architecture cleanup, documentation cleanup,
documentation reordering, scorecard paydown, repository-maintenance edits, or
refactor-only work. If such work is detected, do not treat it as a pre-task or
slice candidate; record it as input for `$refactor`, which runs before
engineering review generation.
After the user picks one or more ideas from the menu, that reply authorizes the
agent to order the selected ideas and run the batch clarification gate. If that
gate emits no questions, the agent may continue from brief to spec, plan,
implementation, CI, and commit when all gates pass. If that gate emits questions,
the user's answer authorizes continuing with those answers applied. It does
**not** authorize pushing, branch creation, bypassing CI, or guessing through
blockers.

## Count handling

1. Parse `{count}` as an integer.
2. If missing, zero, negative, or not an integer, ask the user for a valid count.
3. Do not impose a maximum cap. The requested count is the execution limit, even when it is large.
4. If the initial command includes idea text after the count, do **not** show
   the 15 idea menu. Treat the inline idea text as selected input and proceed
   to ordering and the batch clarification gate.
5. If no inline idea text is provided, show **15** slice ideas before execution,
   regardless of the execution count.
6. After inline idea parsing or user menu selection, set the execution target to the smaller of:
   - the requested count, and
   - the number of viable selected ideas.
7. If the user selects more ideas than the execution target, order all selected ideas,
   execute the first target-sized prefix, and report the rest as deferred.
8. After the execution target is completed and committed, run the post-loop refactor/review handoff; stop
   earlier only on a hard stop condition.

## Defaults for non-blocking choices

Use these defaults whenever the repo context supports more than one viable choice and the
choice is not a true blocker:

- Choose the smallest vertical slice.
- Prefer player-visible progress.
- Prefer existing backlog, open gaps, ADR deferred work, or in-flight specs over new inventions.
- Exclude architecture cleanup, documentation-only work, scorecard paydown, pure test
  reorganization, coordinator splitting, naming cleanup, lifecycle hygiene, stale-link cleanup,
  and similar maintenance unless it is necessary inside a player-visible feature slice.
- When a selected idea has both backend and client impact, keep them in the same vertical slice
  so the result is playable or watchable rather than a backend-only installment.
- Defer risky or large scope into explicit non-goals.
- Prefer server authority, deterministic sim changes, shared contracts, and bot proof over
  client-only presentation shortcuts.

## Hard stop conditions

Stop immediately and report the reason if any of these occur:

1. The starting git state has dirty changes unrelated to the command, or mixed changes that
   cannot be safely attributed to one current slice.
2. A `$next`, `$spec`, `$plan`, `$execute`, or `$finish` gate finds a blocker that cannot be
   resolved conservatively from repo context.
3. A spec or plan contradiction cannot be resolved by the defaults above.
4. CI fails after reasonable diagnosis and focused fixes.
5. A decision requires product/design judgment not covered by the defaults.
6. Secrets, credentials, `.env`, or local-only artifacts appear in the diff or staged changes.
7. Completing another slice would exceed the requested execution target.
8. No inline idea was provided and the user has not yet selected ideas from the generated menu.
9. A selected idea is too vague, too large, or not verifiable enough to turn into a small slice.
10. Batch clarification questions were emitted and the user has not answered them yet.
11. The next available work is architecture cleanup, documentation maintenance, scorecard
    improvement, or refactor-only paydown rather than feature/gameplay progress; stop and tell the
    user to run `$refactor` if they want autonomous quality paydown.

Do not create branches. Do not push. Do not use `--no-verify`, `--amend`, or destructive git
commands unless the user explicitly asks in a later message.

## Phase 0 — Preflight

1. Read [`CLAUDE.md`](../../CLAUDE.md), [`PROGRESS.md`](../../PROGRESS.md), and
   [`AGENTS.md`](../../AGENTS.md).
2. Run `git status --short`.
3. If dirty changes exist before the loop starts, inspect enough to determine whether they are
   intentionally part of a single in-flight slice. If ambiguous or unrelated, stop and ask.
4. Record the current branch with `git branch --show-current`; stay on it for the whole loop.
5. Load the current repo skill files as needed:
   - [`skills/next/SKILL.md`](../next/SKILL.md)
   - [`skills/spec/SKILL.md`](../spec/SKILL.md)
   - [`skills/plan/SKILL.md`](../plan/SKILL.md)
   - [`skills/execute/SKILL.md`](../execute/SKILL.md)
   - [`skills/finish/SKILL.md`](../finish/SKILL.md)
6. If `PROGRESS.md` says an engineering review is due, record that `$refactor` and `$review`
   should run after the requested feature loop completes. Do **not** stop the loop or replace a
   selected feature slice with review, refactor, or cleanup work solely because the review cadence
   is due.

## Phase 1 — Idea intake, menu, and selection

Before writing any spec, plan, or code:

1. Determine whether the initial command includes idea text after the count.
   - If yes, treat that text as the selected idea input. If it clearly contains
     multiple ideas, split only on explicit separators such as numbered lines,
     semicolons, or separate paragraphs.
   - If no, use the **next** skill discovery inputs to gather candidate slices from `PROGRESS.md`,
   ADRs, existing specs/plans, open gaps, bot gaps, and natural project trajectory.
2. Ignore repository-maintenance and architecture-cleanup work for autoloop selection. This
   includes documentation cleaning, updating, ordering, index repair, lifecycle-table correction,
   stale-link cleanup, backlog/review hygiene, file-size paydown, coordinator splitting, and
   scorecard improvement whose purpose is codebase health rather than a new gameplay/system proof.
   Mention notable cleanup signals only as `$refactor` input; do not ask permission to perform
   them during `$autoloop`.
4. If inline idea text was provided, skip menu generation and continue to the
   "selected ideas" validation and ordering steps below.
5. Present **15** slice ideas when no inline idea text was provided.
   Slice ideas must be features, game improvements, player-facing additions, or the minimal
   supporting system/tooling needed to make those additions playable and verifiable. They must not
   be cleanup-only, architecture-only, documentation-only, test-only, or scorecard-paydown work.
6. Keep each menu idea compact and selection-friendly:
   - stable number or short code the user can choose,
   - codename,
   - one-line player/system value,
   - size `S | M | L | XL`,
   - touch surfaces,
   - main risk or dependency.
7. When no inline idea text was provided, do **not** choose for the user during
   the first autoloop response. Ask them to pick the idea numbers/codenames they like.
8. If the user asks for fewer than 15 ideas, honor that explicit request; otherwise the
   default menu remains 15 ideas.
9. If fewer than 15 credible candidates exist, show the credible candidates and say why the
   menu is shorter.
10. Do not write or modify files during the idea menu phase.

When selected ideas come from inline text or from a later menu reply:

1. Validate that each selected idea maps to a menu candidate, a clearly equivalent
   user idea, or a concrete inline idea from the initial command.
2. Drop or split ideas that are too large for one slice; if the split is not obvious, stop.
3. Order the selected ideas using:
   - dependency order first,
   - smallest vertical proof next,
   - player-visible progress next,
   - bot/test proof clarity as a tie-breaker.
4. Show the ordered execution queue and the execution target.
5. Run the batch clarification gate below before beginning the per-slice loop.

## Phase 2 — Batch clarification gate

Before writing specs, plans, or code for selected ideas:

1. Do a lightweight discovery pass for every viable selected idea:
   - cover all ideas in the execution target,
   - also cover deferred selected ideas when a question affects splitting, ordering, or future viability,
   - re-check `PROGRESS.md`, ADRs, existing specs/plans, open gaps, and bot/test proof needs.
2. Anticipate the questions that would otherwise be raised later by `$next`, `$spec`, `$plan`,
   or `$execute` gates.
3. Use the defaults for non-blocking choices instead of asking. Only ask when the answer requires
   product/design judgment, changes slice boundaries, affects protocol/schema contracts, or could
   make implementation unsafe to start while the user is away.
4. If any blocking questions remain, present them all at once, grouped by selected idea. Keep them
   answerable in a single reply:
   - label each question as `Required for this run` or `Deferred/future`,
   - include a short reason the answer blocks autonomous execution,
   - include a conservative suggested default only when the repo context supports one,
   - state that unanswered required questions will stop the loop.
5. Stop and wait for the user's answers if any required questions were emitted.
6. If there are no required questions, say that the selected queue has no blocking questions and
   begin the per-slice loop without asking for another confirmation.
7. After the user answers, apply the answers consistently across all selected slices. If an answer
   changes ordering, scope, or viability, update the queue and report the adjustment before execution.

## Per-slice loop

Repeat over the clarified ordered selected queue until the execution target is complete or a
stop condition fires.

### 1. Discover

Use the selected menu idea as the next-slice brief. Re-check `PROGRESS.md`, ADRs,
existing specs/plans, open gaps, and natural project trajectory before committing to it.

- Produce the normal next-slice brief for the selected idea.
- If implementation order needs adjustment after a completed slice, reorder remaining selected
  ideas using the ordering rules above and report the adjustment.
- If the selected idea still requires user judgment not covered by the batch clarification gate,
  stop and ask only that newly discovered blocking question.

### 2. Spec

Use the **spec** skill to create or update `docs/specs/vN_spec-<codename>.md`
from the selected brief.

- This autoloop invocation counts as the user's approval to write the spec.
- Keep the slice small and verifiable.
- If client UI, camera, inventory presentation, or art is in scope, consult
  [`docs/researchs/godot-plugins-and-shortcuts.md`](../../docs/researchs/godot-plugins-and-shortcuts.md)
  and ensure the spec or plan records an adopt / borrow / reject decision.

### 3. Plan

Use the **plan** skill on the spec file.

- Run the spec review gate.
- Fix only minor spec gaps that are clearly implied by the brief and defaults.
- Write or update `docs/plans/vN_YYYY-MM-DD-<codename>.md`.
- Stop on unresolved questions not already covered by the batch clarification answers.

### 4. Execute

Use the **execute** skill on the plan file.

- Run the plan review gate.
- Implement task-by-task in plan order.
- Update plan checkboxes as tasks complete.
- Run focused verification while iterating and `make ci` before claiming done.
- Stop on unresolved ambiguity, unplanned protocol/schema changes, or persistent failing tests.

### 5. Finish

Use the **finish** skill.

- Consolidate `PROGRESS.md`.
- Run `make ci`.
- Stage only files belonging to the current slice.
- Commit with exactly:

```text
feat: v{N}: {title of this slice}
```

- Do not push.

### 6. Post-commit checkpoint and context hygiene

After each commit:

1. Run `git status --short` and `git log -1 --oneline`.
2. Write a concise continuation checkpoint in chat before any context clearing attempt:
   - completed count and execution target,
   - current branch,
   - last commit hash and message,
   - current `git status --short`,
   - ordered selected queue with completed, next, and deferred ideas,
   - any new hard-stop risks or deferred scope.
3. Only after the checkpoint exists, clear or compact context if the runtime exposes a safe
   explicit mechanism for doing so.
4. Do not clear context if the commit failed, `make ci` did not pass, git status is ambiguous,
   or the next selected idea cannot be reconstructed from the checkpoint.
5. If no explicit context clearing mechanism is available, do manual context hygiene instead:
   treat previous slice implementation details as stale, re-read `CLAUDE.md`, `PROGRESS.md`,
   `AGENTS.md`, the remaining selected idea, and the relevant skill files before the next slice.
6. Never create repository files solely as autoloop memory or checkpoints.

### 7. Continue or stop

After the checkpoint and optional context hygiene:

1. If only unrelated pre-existing dirty changes remain, stop and report them.
2. If the worktree is clean and the execution target is not reached, begin the next slice.
3. If the worktree is clean and the execution target is reached, run the post-loop handoff below.

### 8. Post-loop refactor/review handoff

After all requested slices have completed and committed:

1. Re-read `PROGRESS.md`.
2. If the engineering-review cadence is due, do **not** generate the review from `$autoloop`.
   Report that the requested feature slices are complete and that the next autonomous step is
   `$refactor`, followed by `$review`.
3. `$refactor` is responsible for scorecard-driven minor cleanup commits before the fresh
   engineering review is written.
4. If no engineering review is due, stop and report completion.

## Reporting

For each completed slice, report:

- Slice number and title.
- Commit hash and commit message.
- `make ci` result.
- Any deferred scope added to `PROGRESS.md`.

If the loop stops early, report:

- Completed slice count.
- Exact stop condition.
- Current git status summary.
- The next manual command the user should run after resolving the blocker.

If execution stops before all selected ideas are completed, also report the remaining ordered
ideas that were not started.
