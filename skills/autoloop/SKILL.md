---
name: autoloop
description: >-
  Run an autonomous SDD loop for a bounded number of continuous slices. Use when
  the user runs $autoloop N, /autoloop N, or asks Codex to repeatedly run next,
  spec, plan, execute, and finish with default decisions and committed slices.
disable-model-invocation: true
---

# $autoloop — Autonomous SDD Slice Loop

**Trigger:** `$autoloop {count}` or `/autoloop {count}`

Examples:

- `$autoloop 1` — complete one autonomous slice.
- `$autoloop 3` — complete up to three autonomous slices, stopping early on any gate.

**Announce at start:** "Using the **autoloop** skill to run a bounded autonomous SDD loop."

## Purpose

Run the normal SDD workflow repeatedly:

```text
$next -> $spec -> $plan -> $execute -> $finish
```

This skill is an explicit user authorization to continue from brief to spec, plan,
implementation, CI, and commit when all gates pass. It does **not** authorize pushing,
branch creation, bypassing CI, or guessing through blockers.

## Count handling

1. Parse `{count}` as an integer.
2. If missing, zero, negative, or not an integer, ask the user for a valid count.
3. If `{count} > 3`, run at most **3** slices and say the request was capped.
4. Stop after `{count}` completed and committed slices, or earlier on any stop condition.

## Defaults for non-blocking choices

Use these defaults whenever the repo context supports more than one viable choice and the
choice is not a true blocker:

- Choose the smallest vertical slice.
- Prefer player-visible progress.
- Prefer existing backlog, open gaps, ADR deferred work, or in-flight specs over new inventions.
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
7. Completing another slice would exceed the capped count of 3.

Do not create branches. Do not push. Do not use `--no-verify`, `--amend`, or destructive git
commands unless the user explicitly asks in a later message.

## Phase 0 — Preflight

1. Read [`CLAUDE.md`](../../CLAUDE.md), [`docs/PROGRESS.md`](../../docs/PROGRESS.md), and
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

## Per-slice loop

Repeat until the requested count is complete or a stop condition fires.

### 1. Discover

Use the **next** skill to identify the highest-value next slice from `PROGRESS.md`,
ADRs, existing specs/plans, open gaps, and natural project trajectory.

- Produce the normal next-slice brief.
- If multiple viable candidates remain, choose using the autoloop defaults.
- If the choice still requires user judgment, stop.

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
- Stop on unresolved questions.

### 4. Execute

Use the **execute** skill on the plan file.

- Run the plan review gate.
- Implement task-by-task in plan order.
- Update plan checkboxes as tasks complete.
- Run focused verification while iterating and `make ci` before claiming done.
- Stop on unresolved ambiguity, unplanned protocol/schema changes, or persistent failing tests.

### 5. Finish

Use the **finish** skill.

- Consolidate `docs/PROGRESS.md`.
- Run `make ci`.
- Stage only files belonging to the current slice.
- Commit with exactly:

```text
feat: v{N}: {title of this slice}
```

- Do not push.

### 6. Continue or stop

After each commit:

1. Run `git status --short`.
2. If only unrelated pre-existing dirty changes remain, stop and report them.
3. If the worktree is clean and the requested count is not reached, begin the next slice.

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
