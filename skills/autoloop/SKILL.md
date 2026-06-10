---
name: autoloop
description: >-
  Present a curated menu of 5-10 possible SDD slices, wait for the user to pick
  the ideas they like, then order and execute a bounded autonomous loop through
  next, spec, plan, execute, and finish with committed slices. Use when the user
  runs $autoloop N, /autoloop N, or asks Codex to curate and run multiple SDD slices.
disable-model-invocation: true
---

# $autoloop — Curated Autonomous SDD Slice Loop

**Trigger:** `$autoloop {count}` or `/autoloop {count}`

Examples:

- `$autoloop 1` — present 5-10 ideas, then complete one selected slice.
- `$autoloop 3` — present 5-10 ideas, then complete up to three selected slices, stopping early on any gate.

**Announce at start:** "Using the **autoloop** skill to curate slice ideas, then run a bounded autonomous SDD loop after you choose."

## Purpose

Run the normal SDD workflow repeatedly after one explicit user selection gate:

```text
$autoloop -> idea menu -> user picks -> agent orders picks -> $spec -> $plan -> $execute -> $finish
```

The initial `$autoloop` invocation authorizes preflight and idea discovery only.
After the user picks one or more ideas from the menu, that reply authorizes the
agent to order the selected ideas and continue from brief to spec, plan,
implementation, CI, and commit when all gates pass. It does **not** authorize
pushing, branch creation, bypassing CI, or guessing through blockers.

## Count handling

1. Parse `{count}` as an integer.
2. If missing, zero, negative, or not an integer, ask the user for a valid count.
3. If `{count} > 3`, run at most **3** slices and say the execution request was capped.
4. Always show **5-10** slice ideas before execution, regardless of the execution count.
5. After the user picks ideas, set the execution target to the smaller of:
   - the capped count, and
   - the number of viable selected ideas.
6. If the user selects more ideas than the execution target, order all selected ideas,
   execute the first target-sized prefix, and report the rest as deferred.
7. Stop after the execution target is completed and committed, or earlier on any stop condition.

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
8. The user has not yet selected ideas from the generated menu.
9. A selected idea is too vague, too large, or not verifiable enough to turn into a small slice.

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

## Phase 1 — Idea menu and selection

Before writing any spec, plan, or code:

1. Use the **next** skill discovery inputs to gather candidate slices from `PROGRESS.md`,
   ADRs, existing specs/plans, open gaps, bot gaps, and natural project trajectory.
2. Present **5-10** slice ideas. Prefer 8 when the backlog has enough credible options.
3. Keep each idea compact and selection-friendly:
   - stable number or short code the user can choose,
   - codename,
   - one-line player/system value,
   - size `S | M | L | XL`,
   - touch surfaces,
   - main risk or dependency.
4. Do **not** choose for the user during the first autoloop response. Ask them to pick the
   idea numbers/codenames they like.
5. If the user asks for fewer than 5 ideas, honor that explicit request; otherwise the
   default menu remains 5-10 ideas.
6. If fewer than 5 credible candidates exist, show the credible candidates and say why the
   menu is shorter.
7. Do not write or modify files during the idea menu phase unless needed to inspect repo state.

When the user replies with selected ideas:

1. Validate that each selected idea maps to a menu candidate or a clearly equivalent user idea.
2. Drop or split ideas that are too large for one slice; if the split is not obvious, stop.
3. Order the selected ideas using:
   - dependency order first,
   - smallest vertical proof next,
   - player-visible progress next,
   - bot/test proof clarity as a tie-breaker.
4. Show the ordered execution queue and the capped execution target.
5. Begin the per-slice loop without asking for another confirmation, unless a hard stop
   condition applies.

## Per-slice loop

Repeat over the ordered selected queue until the execution target is complete or a stop
condition fires.

### 1. Discover

Use the selected menu idea as the next-slice brief. Re-check `PROGRESS.md`, ADRs,
existing specs/plans, open gaps, and natural project trajectory before committing to it.

- Produce the normal next-slice brief for the selected idea.
- If implementation order needs adjustment after a completed slice, reorder remaining selected
  ideas using the ordering rules above and report the adjustment.
- If the selected idea still requires user judgment, stop.

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

If execution stops before all selected ideas are completed, also report the remaining ordered
ideas that were not started.
