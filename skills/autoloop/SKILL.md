---
name: autoloop
description: >-
  Complete every eligible ERP SDD roadmap step in the next incomplete phase by
  repeatedly running /next no-human. Use when the user runs $autoloop or
  /autoloop to advance a whole phase autonomously.
disable-model-invocation: true
---

# $autoloop — Autonomous ERP Phase Loop

**Trigger:** `$autoloop` or `/autoloop`.

**Announce at start:** "Uso `$autoloop` para completar los pasos elegibles de la próxima fase incompleta mediante `/next no-human`, sin cruzar a otra fase."

## Purpose

Complete the **next incomplete phase** in `erp-base`, in its roadmap order.
For each eligible roadmap step in that phase, run:

```text
/next S<N> no-human
```

`/next no-human` remains the source of truth for the full lifecycle of one
step: scope, specification, plan, tests and fixtures, implementation, final
validation, atomic commits, and `STATUS.md` closeout. `$autoloop` only selects
the phase, repeats that one-step operation, and enforces the phase boundary.

Do not use the former `next -> spec -> plan -> execute -> finish` workflow.
Do not invoke the ARPG `next`, `spec`, `plan`, `execute`, or `finish` skills.

## Phase selection

Before the first invocation:

1. Read the ERP repository's `AGENTS.md`, `docs/sdd/STATUS.md`, and
   `docs/architecture.md`.
2. Inspect `git status --short` and the current branch. Follow `/next` dirty
   worktree safety; do not switch branches, create branches, push, or create a
   pull request.
3. In `STATUS.md`, select the first phase in document order that contains at
   least one row whose state is not `hecho`.
4. Record the selected phase, its roadmap rows, and which rows are currently
   eligible. A blocked row is part of the selected phase even if earlier rows
   are already complete.

The command is intentionally phase-scoped: it never starts a row in a later
phase, even when that row would otherwise be eligible.

## Execution loop

Repeat until the selected phase has no non-`hecho` rows:

1. Re-read `docs/sdd/STATUS.md`; do not rely on stale state.
2. Within the selected phase, choose the first row in roadmap order that is not
   `hecho` and whose prerequisites are all `hecho` and which is not blocked.
3. If no such row exists, stop. Report the first remaining row and its unmet
   prerequisite or pending architecture decision. Do not skip it or begin a
   later phase.
4. Run `/next S<N> no-human` for that row, following the current
   `/Users/mmandrille/git/erp-base/.codex/skills/next/SKILL.md` instructions
   exactly.
5. Continue only if `/next no-human` reports `PASS`, the selected row is now
   `hecho`, and its required atomic commit completed. Otherwise stop and report
   its exact `BLOCKED`, `PARTIAL`, or `FAIL` evidence.

Before every iteration, preserve unrelated worktree changes. If changes cannot
be safely attributed to the active roadmap step, stop rather than commit or
modify them.

## Stop conditions

Stop immediately and leave later work untouched when:

- a pending architecture/product decision blocks a row;
- `/next no-human` reports `BLOCKED`, `PARTIAL`, or `FAIL`;
- a test, lint, coverage, acceptance criterion, or atomic commit fails;
- the worktree is dirty with unrelated or inseparable changes; or
- all steps in the selected phase are `hecho`.

Do not invent decisions, weaken validation, bypass `/next` safeguards, retry a
failing step indefinitely, or automatically start the next phase.

## Reporting

After each `/next no-human`, report the step identifier, resulting state,
commit, and concise validation result. At the end, report one of:

- `PASS`: every row in the selected phase is `hecho`, with the completed steps
  and validation evidence; or
- `BLOCKED`, `PARTIAL`, or `FAIL`: the stopped step, exact evidence, current
  git-status summary, and the concrete action needed before rerunning
  `$autoloop`.

When a phase completes, state that the next `$autoloop` invocation may select
the next incomplete phase; do not start it automatically.
