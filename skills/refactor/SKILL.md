---
name: refactor
description: >-
  Run autonomous scorecard-driven repository improvement before engineering
  review generation. Read the latest engineering review overview and companion
  reports, identify minor architecture, maintainability, test, documentation,
  and process paydown tasks that can improve every scorecard area toward 9+,
  implement them as small verified commits without consuming slice numbers, and
  stop when only feature work or major design decisions remain. Use when the
  user runs $refactor, /refactor, asks to improve review scorecard areas, or
  when a review is due before generating the fresh review set.
disable-model-invocation: true
---

# $refactor — Scorecard Paydown Before Review

**Trigger:** `$refactor` or `/refactor`, optionally with a target such as `9+`, `v180`, or a
specific scorecard area.

Examples:

- `$refactor` — read the latest review set, then make autonomous minor cleanup commits before the
  next engineering review.
- `$refactor 9+` — keep improving review scorecard areas until every area appears plausibly at 9+
  or only major/product work remains.
- `$refactor maintainability tests` — focus on those scorecard areas first, while still avoiding
  regressions in the rest.

**Announce at start:** "Using the **refactor** skill to pay down review scorecard gaps with small verified commits before review generation."

## Purpose

Improve the engineering review scorecard without creating feature slices. This command is the
autonomous quality-paydown step that runs before a fresh `$review` when review cadence is due.

```text
$refactor -> latest review overview + companions -> scorecard gap queue -> minor fix commit(s) -> stop -> $review
```

The command targets the stable review scorecard areas:

- Architecture (boundaries)
- Architecture (internal cohesion)
- Technical correctness
- Test quality
- Maintainability
- Documentation
- Process / SDD
- Overall

The goal is to push every area toward **9+**. Do not fake or document scores upward; make concrete
repo improvements that a later `$review` can verify from evidence.

## Scope

Allowed work:

- Small architecture/internal-cohesion cleanup with no behavior change.
- File-size and coordinator paydown that follows the maintainability ratchet.
- Test quality improvements, focused fixtures, clearer assertions, and missing regression tests.
- Documentation/process repairs tied to review findings, stale instructions, command contracts,
  codemap drift, or review handoff clarity.
- Tooling hygiene that improves validation, CI reliability, determinism checks, or review evidence.
- Low-risk correctness fixes found while addressing a review recommendation.

Disallowed work:

- New gameplay features, player-facing additions, balance changes, or content expansion.
- Public protocol/schema changes unless the latest review explicitly identifies a minor bug and the
  change is fully coordinated and verified.
- Large rewrites, speculative architecture projects, or work that needs product/design judgment.
- SDD slice creation, `docs/specs/vN_*`, `docs/plans/vN_*`, `docs/as-built/vN_*`, or `feat: vN:`
  commits.
- Review generation itself. `$refactor` prepares the repo; `$review` writes the new review set.

## Commit Policy

Use minor commits, not slice commits:

- `refactor: ...` for structure-only code movement or simplification.
- `test: ...` for test-only improvements.
- `docs: ...` for documentation/process repair.
- `fix: ...` for small correctness fixes.
- `chore: ...` for tooling/build hygiene.

Keep each commit focused and independently verified. Do not push. Do not create branches. Do not use
`--no-verify`, `--amend`, or destructive git commands unless the user explicitly asks later.

## Hard Stop Conditions

Stop immediately and report the reason if any of these occur:

1. The starting git state has dirty changes unrelated to the command, or mixed changes that cannot
   be safely attributed to this refactor run.
2. No prior engineering review overview and companion reports can be found.
3. A needed improvement requires feature design, product judgment, gameplay tuning, or a new SDD
   slice.
4. The next useful task is larger than a minor commit or cannot be verified with focused checks.
5. A change would alter public protocol/schema contracts without an explicit, narrow review-backed
   reason and coordinated tests.
6. CI or focused verification fails after reasonable diagnosis and focused fixes.
7. Secrets, credentials, `.env`, or local-only artifacts appear in the diff or staged changes.
8. The scorecard appears plausibly at 9+ in all areas, or the remaining gaps are major work that
   should become future feature slices or explicit plans.

## Phase 0 — Baseline

1. Read [`PROGRESS.md`](../../PROGRESS.md), [`CLAUDE.md`](../../CLAUDE.md), [`AGENTS.md`](../../AGENTS.md),
   and [`docs/CODEMAP.md`](../../docs/CODEMAP.md).
2. Run `git status --short`; stop on ambiguous unrelated dirt.
3. Record the current branch with `git branch --show-current`; stay on it.
4. Locate the latest overview from `PROGRESS.md` → **Last engineering review**. If absent, choose the
   newest `docs/reviews/*_v*-overview.md`.
5. Read that overview plus its companion backend, client, and extras reports.
6. Extract current scorecard areas, scores, top recommendations, and unresolved prior-review
   findings.

## Phase 1 — Build The Paydown Queue

Create an internal queue of minor tasks from the latest review evidence.

Prioritize:

1. Lowest scorecard areas first.
2. Cross-cutting changes that improve more than one area.
3. Tasks with focused verification and low blast radius.
4. Documentation/process repairs that prevent future agents from choosing the wrong workflow.
5. Maintainability-ratchet reductions where the target file is already implicated by the review.

For each candidate, classify it:

- `minor-commit`: can be done now under `$refactor`.
- `future-slice`: player-facing or feature/system work for `$autoloop`.
- `future-plan`: too broad or design-heavy for autonomous paydown.
- `reject`: stale, already fixed, or no longer supported by current evidence.

Do not ask the user to approve each minor task. Ask only if a decision would change product
behavior, public contracts, or scope boundaries.

## Phase 2 — Execute Minor Commits

Repeat until the queue is exhausted, all areas are plausibly 9+, or a hard stop fires.

For each `minor-commit` task:

1. Re-read the specific files named by the review and `docs/CODEMAP.md`.
2. Make the smallest coherent edit.
3. Run focused verification that covers the changed files.
4. If the change is broad or touches core checks, run `make ci` before committing.
5. Stage only files belonging to the task.
6. Commit with the minor commit policy above.
7. Run `git status --short` and `git log -1 --oneline`.
8. Report a checkpoint:
   - completed refactor commit count,
   - current branch,
   - last commit hash and message,
   - verification run,
   - scorecard areas helped,
   - remaining queue summary.

Do not update `PROGRESS.md` review cadence during `$refactor` unless the task is specifically a
documentation/process repair and the existing cadence text is wrong. `$review` owns Last/Next
engineering review updates.

## Phase 3 — Stop And Handoff

When no more safe minor commits remain:

1. Run `git status --short`.
2. If the final worktree is clean, report that `$refactor` is complete and the next step is
   `$review`.
3. If the final worktree is dirty because verification failed or a hard stop fired, report the exact
   files and stop condition.
4. Summarize remaining `future-slice` and `future-plan` items so `$autoloop` or a later explicit
   plan can pick them up.

## Reporting

For each completed minor commit, report:

- Commit hash and message.
- Scorecard area(s) targeted.
- Verification command and outcome.

If no commits were made, report whether the repo already appears at the target or which hard stop
prevented autonomous paydown.
