---
name: autoloop
description: >-
  Complete one SDD slice autonomously from the current workflow position in the
  session. Resume after next/spec/plan/execute, or start from next when nothing
  is prepared. Use when the user runs $autoloop, /autoloop, or asks to finish
  the slice loop from wherever they left off.
disable-model-invocation: true
---

# $autoloop — Autonomous Single-Slice SDD Continuation

**Trigger:** `$autoloop` or `/autoloop`, optionally followed by inline idea text.

Examples:

- `$autoloop` — detect the current session position; if nothing is prepared, run **next** and wait for your answers, then continue through **spec → plan → execute → finish**.
- `$autoloop` after `/next` — you already approved a brief; continue autonomously through **spec → plan → execute → finish**.
- `$autoloop` after `/spec` — continue through **plan → execute → finish**.
- `$autoloop` after `/plan` — continue through **execute → finish**.
- `$autoloop` after `/execute` — run **finish** only.
- `$autoloop add a town healer` — treat the inline text as the slice idea, run **next** on it, wait for your answers if needed, then continue through the remaining steps.

**Announce at start:** "Using the **autoloop** skill to continue the SDD workflow from the current session position and complete one slice."

## Purpose

Run the normal SDD workflow for **exactly one slice**, starting from wherever the
session already is:

```text
next -> spec -> plan -> execute -> finish
```

`$autoloop` never runs multiple slices, never presents a multi-idea menu, and never
queues follow-up slices. One invocation completes one slice (or stops on a hard
blocker).

The invocation authorizes autonomous continuation through every remaining step
after the detected entry point. It does **not** authorize architecture cleanup,
documentation-only maintenance, scorecard paydown, refactor-only work, branch
creation, pushing, bypassing verification gates, or guessing through blockers.

## Detect the entry point

Before doing any work, determine where this session already is. Check in order:

| Evidence | Entry point | Remaining steps |
|----------|-------------|-----------------|
| Uncommitted implementation, plan checkboxes mostly `[x]`, or user just ran `/execute` | **finish** | finish |
| `docs/plans/vN_*.md` exists for the active slice and plan tasks are not all `[x]` | **execute** | execute → finish |
| `docs/specs/vN_spec-*.md` exists but no matching plan file | **plan** | plan → execute → finish |
| Approved next-slice brief in the current session (from `/next` or inline idea) but no spec file yet | **spec** | spec → plan → execute → finish |
| Inline idea in the `$autoloop` command and no approved brief yet | **next** | next → spec → plan → execute → finish |
| None of the above | **next** | next → spec → plan → execute → finish |

Rules:

1. Prefer **conversation context** over stale repo files when they disagree. If
   the user just approved a brief in chat, treat that as **spec** entry even if
   older draft specs exist for other slices.
2. If multiple in-flight slices make the entry point ambiguous, stop and ask which
   `vN` / codename to continue.
3. If dirty git changes belong to a different slice than the detected entry
   point, stop and ask.
4. Record the detected entry point and remaining steps in chat before continuing.

## Defaults for non-blocking choices

Use these whenever more than one viable choice exists and the answer is not a true blocker:

- Choose the smallest vertical slice.
- Prefer player-visible progress.
- Prefer existing backlog, open gaps, ADR deferred work, or in-flight specs over new inventions.
- Exclude architecture cleanup, documentation-only work, scorecard paydown, pure test
  reorganization, coordinator splitting, naming cleanup, lifecycle hygiene, stale-link cleanup,
  and similar maintenance unless it is necessary inside a player-visible feature slice.
- When the slice has both backend and client impact, keep them in one vertical slice so the
  result is playable or watchable.
- Defer risky or large scope into explicit non-goals.
- Prefer server authority, deterministic sim changes, shared contracts, and bot proof over
  client-only presentation shortcuts.

## Hard stop conditions

Stop immediately and report the reason if any of these occur:

1. The starting git state has dirty changes unrelated to the current slice, or mixed changes that
   cannot be safely attributed to one slice.
2. A **next**, **spec**, **plan**, **execute**, or **finish** gate finds a blocker that cannot be
   resolved conservatively from repo context.
3. A spec or plan contradiction cannot be resolved by the defaults above.
4. Focused per-slice verification fails after reasonable diagnosis and fixes, or final `make ci`
   fails after reasonable diagnosis and focused fixes at finish.
5. A decision requires product/design judgment not covered by the defaults.
6. Secrets, credentials, `.env`, or local-only artifacts appear in the diff or staged changes.
7. The slice idea is too vague, too large, or not verifiable enough for one small slice.
8. A step emitted blocking questions and the user has not answered them yet.
9. The detected work is architecture cleanup, documentation maintenance, scorecard improvement,
   or refactor-only paydown rather than feature/gameplay progress; tell the user to run `$refactor`.

Do not create branches. Do not push. Do not use `--no-verify`, `--amend`, or destructive git
commands unless the user explicitly asks in a later message.

## Phase 0 — Preflight

1. Read [`CLAUDE.md`](../../CLAUDE.md), [`PROGRESS.md`](../../PROGRESS.md), and
   [`AGENTS.md`](../../AGENTS.md).
2. Run `git status --short`.
3. If dirty changes exist, inspect enough to determine whether they belong to the detected slice.
   If ambiguous or unrelated, stop and ask.
4. Record the current branch with `git branch --show-current`; stay on it for the whole run.
5. Detect the entry point using the table above and state the remaining steps.
6. Load the skill files needed for the remaining steps:
   - [`skills/next/SKILL.md`](../next/SKILL.md)
   - [`skills/spec/SKILL.md`](../spec/SKILL.md)
   - [`skills/plan/SKILL.md`](../plan/SKILL.md)
   - [`skills/execute/SKILL.md`](../execute/SKILL.md)
   - [`skills/finish/SKILL.md`](../finish/SKILL.md)
7. If `PROGRESS.md` says an engineering review is due, record that `$review` and `$refactor`
   may run after this slice completes. Do **not** replace the current slice with review or
   refactor work solely because the review cadence is due.

## Phase 1 — Next (only when entry point is **next**)

Use the **next** skill.

1. If inline idea text was provided in the `$autoloop` command, pass it to **next**.
2. Otherwise run **next** with no idea and propose candidates from backlog and trajectory.
3. Follow the **next** skill gates. **Stop and wait** for the user's answers when **next**
   would normally ask — slice priority, idea approval, or brief confirmation.
4. Do not write the spec file during **next**; autoloop continues to **spec** only after the
   brief is approved in chat.
5. Apply the defaults for non-blocking choices instead of asking when the repo context supports one.
6. Once the user approves the brief (or confirms the recommended slice), continue without asking
   for another confirmation to start **spec**.

## Phase 2 — Spec (when entry point is **spec** or later steps not yet done)

Skip this phase if entry point is **plan**, **execute**, or **finish**.

Use the **spec** skill on the approved brief.

- This `$autoloop` invocation counts as approval to write the spec.
- Keep the slice small and verifiable.
- If client UI, camera, inventory presentation, or art is in scope, check existing in-repo
  Godot scripts, scenes, demos, and asset manifests before introducing dependencies, and ensure
  the spec or plan records an adopt / borrow / reject decision.
- Stop on unresolved questions that require product/design judgment not already answered in the session.

## Phase 3 — Plan (when entry point is **plan** or later steps not yet done)

Skip this phase if entry point is **execute** or **finish**.

Use the **plan** skill on the spec file.

- Run the spec review gate.
- Fix only minor spec gaps that are clearly implied by the brief and defaults.
- Write or update `docs/plans/vN_YYYY-MM-DD-<codename>.md`.
- Stop on unresolved questions not already answered in the session.

## Phase 4 — Execute (when entry point is **execute** or **finish** not yet done)

Skip this phase if entry point is **finish** only.

Use the **execute** skill on the plan file.

- Run the plan review gate.
- Implement task-by-task in plan order.
- Update plan checkboxes as tasks complete.
- In `$autoloop` mode, use focused per-slice verification instead of `make ci` unless the slice is
  unusually broad, focused verification cannot cover the risk, or the user explicitly asked for
  per-slice CI.
- Choose the smallest sufficient verification command set for the touched areas, such as
  `make validate-shared`, a focused Go package test, `make client-unit`, `make client-smoke`,
  a specific `make bot scenario=...`, or the bot/client scenario named by the plan.
- Record the exact focused verification commands and results.
- Stop on unresolved ambiguity, unplanned protocol/schema changes, or persistent failing tests.

## Phase 5 — Finish

Use the **finish** skill.

- Consolidate `PROGRESS.md`.
- Run `make ci` before commit unless focused verification already covered the changed surface and
  close-out edits introduce no new risk; when in doubt, run `make ci`.
- Stage only files belonging to this slice.
- Commit with exactly:

```text
feat: v{N}: {title of this slice}
```

- Do not push.

## Phase 6 — Post-slice handoff (optional)

After the slice is committed and the worktree is clean:

1. Re-read `PROGRESS.md`.
2. If the engineering-review cadence is due, tell the user they may run `$review` then `$refactor`
   on the fresh baseline. Do **not** start another slice automatically.
3. Report completion (see Reporting below).

## Reporting

Report:

- Detected entry point and steps completed.
- Slice number, codename, and title.
- Commit hash and commit message.
- Verification commands and results (`make ci` or focused commands).
- Any deferred scope added to `PROGRESS.md`.
- Suggested next manual command (`/next`, `/showme`, `$review`, etc.).

If the run stops early, report:

- Exact stop condition.
- Current git status summary.
- The next manual command after resolving the blocker.
