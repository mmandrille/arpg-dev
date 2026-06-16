---
name: autoloop
description: >-
  Curate and execute autonomous feature/gameplay SDD slices only. Either use
  the user-provided slice idea(s) directly or, when no idea is provided, present
  a curated menu of 15 player-visible feature, game-improvement, or addition
  candidates. Batch blocking clarification questions across the selected queue,
  then execute an autonomous loop through next, spec, plan, execute, and finish
  with committed slices. Use when the user runs $autoloop, /autoloop, provides
  inline autoloop idea(s), or asks Codex to curate and run multiple feature slices.
disable-model-invocation: true
---

# $autoloop — Curated Autonomous Feature Slice Loop

**Trigger:** `$autoloop` or `/autoloop`, optionally followed by inline idea text.

Examples:

- `$autoloop` — present 15 ideas, then complete every viable slice the user selects.
- `$autoloop add a town healer` — treat the inline idea as selected and complete that slice if gates pass.
- `$autoloop idea A; idea B; ...` — treat the inline ideas as selected and complete each viable selected slice.
- `$autoloop show me options for ranged combat polish` — present focused ideas, then complete the viable ideas the user chooses in reply.

**Announce at start:** "Using the **autoloop** skill to curate feature/gameplay slice ideas, then run an autonomous SDD loop after you choose."

## Purpose

Run the normal SDD workflow repeatedly for feature/gameplay progress only. An
autoloop slice should usually be watchable end-to-end and treat backend,
shared contracts, bot proof, and client presentation as one vertical slice when
the behavior reaches the player.

If the user supplied idea text in the initial command, that text is the
selection gate. If the command has no concrete selected idea text, present a
menu and wait for the user to select:

```text
$autoloop with idea -> order inline feature idea(s) -> batch questions -> $spec -> $plan -> $execute -> $finish
$autoloop without idea -> feature idea menu -> user picks -> order picks -> batch questions -> $spec -> $plan -> $execute -> $finish
```

The initial `$autoloop` invocation authorizes preflight and idea discovery only
when no idea text is provided. When concrete idea text is provided, that same
invocation authorizes using the provided idea(s) as the selected queue. The
number of slices to execute is the number of viable ideas in the selected queue,
not a predeclared count.
It does **not** authorize architecture cleanup, documentation cleanup,
documentation reordering, scorecard paydown, repository-maintenance edits, or
refactor-only work. If such work is detected, do not treat it as a pre-task or
slice candidate; record it as input for the post-loop `$review` recommendations
and the `$refactor` pass that follows them.
After the user picks one or more ideas from the menu, that reply authorizes the
agent to order the selected ideas and run the batch clarification gate. If that
gate emits no questions, the agent may continue from brief to spec, plan,
implementation, focused verification, and commit when all gates pass. If that
gate emits questions, the user's answer authorizes continuing with those answers
applied. It does **not** authorize pushing, branch creation, bypassing the final
CI gate, or guessing through blockers.

## Selection handling

1. Do not require, request, or wait for a numeric slice count.
2. If the initial command includes concrete idea text, do **not** show the 15
   idea menu. Treat the inline idea text as selected input and proceed to
   ordering and the batch clarification gate.
3. If no concrete inline idea text is provided, show **15** slice ideas before
   execution, unless the user explicitly asks for a different menu size.
4. After inline idea parsing or user menu selection, set the execution target to
   the number of viable selected ideas.
5. If the user selects one idea, execute one slice. If the user selects many
   ideas, execute as many viable slices as they selected, in the ordered queue.
6. If a selected idea must be split into multiple slices, include every obvious
   split slice in the ordered queue and tell the user before execution. If the
   split requires product/design judgment, stop and ask.
7. If the user still provides a leading number from the old syntax, treat it as
   a menu-size hint only when there is no inline idea text after it. Do not use
   it as an execution limit. When there is inline idea text after a leading
   number, ignore the number and parse the remaining text as selected input.
8. After the selected queue is completed and committed, run the post-loop
   review/refactor handoff; stop earlier only on a hard stop condition.

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
4. A focused per-slice verification command fails after reasonable diagnosis and fixes, or the
   final post-loop `make ci` fails after reasonable diagnosis and focused fixes.
5. A decision requires product/design judgment not covered by the defaults.
6. Secrets, credentials, `.env`, or local-only artifacts appear in the diff or staged changes.
7. Completing another slice would go beyond the selected queue.
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
6. If `PROGRESS.md` says an engineering review is due, record that `$review` and `$refactor`
   should run after the selected feature loop completes. Do **not** stop the loop or replace a
   selected feature slice with review, refactor, or cleanup work solely because the review cadence
   is due.

## Phase 1 — Idea intake, menu, and selection

Before writing any spec, plan, or code:

1. Determine whether the initial command includes concrete idea text.
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
4. Show the ordered execution queue and the execution target, which is the count
   of viable selected ideas.
5. Run the batch clarification gate below before beginning the per-slice loop.

## Phase 2 — Batch clarification gate

Before writing specs, plans, or code for selected ideas:

1. Do a lightweight discovery pass for every viable selected idea:
   - cover all ideas in the selected execution queue,
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

Repeat over the clarified ordered selected queue until every selected viable idea is complete or a
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
- If client UI, camera, inventory presentation, or art is in scope, check existing in-repo
  Godot scripts, scenes, demos, and asset manifests before introducing dependencies, and ensure
  the spec or plan records an adopt / borrow / reject decision.

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
- Run focused verification while iterating. In autoloop batch mode, do **not** run `make ci`
  for each slice unless the slice is unusually broad, focused verification cannot cover the
  risk, or the user explicitly asks for per-slice CI.
- Choose the smallest sufficient verification command set for the touched areas, such as
  `make validate-shared`, a focused Go package test, `make client-unit`, `make client-smoke`,
  a specific `make bot scenario=...`, or the bot/client scenario named by the plan.
- Record the exact focused verification commands and results in the slice report/checkpoint.
- Stop on unresolved ambiguity, unplanned protocol/schema changes, or persistent failing tests.

### 5. Finish

Use the **finish** skill.

- Consolidate `PROGRESS.md`.
- In autoloop batch mode, use the finish skill's autoloop exception: commit each completed slice
  after focused verification, without running `make ci` per slice when focused tests are adequate.
- Do run `make ci` before an individual slice commit only when focused verification is not enough
  for the changed surface, when the slice is broad enough to justify it, or when the user explicitly
  requested per-slice CI.
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
   - completed count and selected queue size,
   - current branch,
   - last commit hash and message,
   - current `git status --short`,
   - ordered selected queue with completed, next, and deferred ideas,
   - any new hard-stop risks or deferred scope.
3. Only after the checkpoint exists, clear or compact context if the runtime exposes a safe
   explicit mechanism for doing so.
4. Do not clear context if the commit failed, focused verification did not pass, git status is
   ambiguous, or the next selected idea cannot be reconstructed from the checkpoint.
5. If no explicit context clearing mechanism is available, do manual context hygiene instead:
   treat previous slice implementation details as stale, re-read `CLAUDE.md`, `PROGRESS.md`,
   `AGENTS.md`, the remaining selected idea, and the relevant skill files before the next slice.
6. Never create repository files solely as autoloop memory or checkpoints.

### 7. Continue or stop

After the checkpoint and optional context hygiene:

1. If only unrelated pre-existing dirty changes remain, stop and report them.
2. If the worktree is clean and selected ideas remain, begin the next slice.
3. If the worktree is clean and the selected queue is complete, run the post-loop handoff below.

### 8. Post-loop review/refactor handoff

After all selected slices have completed and committed:

1. Re-read `PROGRESS.md`.
2. Run `make ci` once for the completed batch before any review/refactor handoff. If it fails,
   diagnose, fix minor bugs that belong to the completed slices, re-run focused commands as needed,
   then re-run `make ci` until it passes or a hard stop condition fires. Commit CI-fix changes as a
   small follow-up commit that clearly belongs to the autoloop batch.
3. If the engineering-review cadence is due, run `$review` first to generate the fresh review set
   from the just-completed feature baseline. Treat the autoloop command as authorization for this
   post-loop review only after the selected feature queue is committed and the worktree is clean.
4. After `$review` writes the fresh overview and companion reports, run `$refactor` pointed at that
   new review. `$refactor` must classify every recommendation as `minor-commit`, `future-slice`,
   `future-plan`, or `reject`, and land only small verified cleanup commits.
5. Stop after `$refactor` reports no more safe minor commits remain, or earlier if either skill
   hits a hard stop.
6. If no engineering review is due, stop and report completion.

## Reporting

For each completed slice, report:

- Slice number and title.
- Commit hash and commit message.
- Focused verification commands and results.
- Any deferred scope added to `PROGRESS.md`.

For the completed autoloop batch, report:

- Final `make ci` result.
- Any minor CI-fix follow-up commits made after the per-slice commits.

If the loop stops early, report:

- Completed slice count.
- Exact stop condition.
- Current git status summary.
- The next manual command the user should run after resolving the blocker.

If execution stops before all selected ideas are completed, also report the remaining ordered
ideas that were not started.
