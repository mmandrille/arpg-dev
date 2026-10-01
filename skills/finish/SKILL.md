---
name: finish
description: >-
  Close a verified standalone slice or an integrated slice batch: reconcile
  lifecycle docs, run the appropriate CI gate, and commit on the current branch.
  In a batch, only the main coordinator invokes /finish after all handoffs.
disable-model-invocation: true
---

# /finish — Verified closeout

Read `PROGRESS.md` current status, open gaps, and checklist; inspect Git state, specs, plans, as-built notes, and actual verification evidence. Never create a branch, push without a request, bypass hooks, or commit secrets/local-only artifacts.

## Batch coordinator

Use [the batch workflow](../autoloop/references/batch-workflow.md). Confirm every accepted `vN` is represented in `main`, including changed/untracked files, tests, assets/licenses, and referenced ignored evidence. Resolve overlap, reconcile CODEMAP/registries, run focused integration checks, and verify no accepted slice remains blocked. Confirm the coordinator's combined `make ci` passed on this exact integrated code state; run it here only if the gate has not yet run or the code changed afterward. Fix failures and rerun until the final state is green. Do not run per-slice CI.

After the green gate, mark each spec and plan accurately, write/update every as-built, update `PROGRESS.md` and `docs/progress/slice-lifecycle.md`, inspect staged changes and `git diff --check`, and commit logically grouped slice or batch changes on the already checked-out `main`. Prefer `feat: vN: <title>` for a slice commit; a batch commit must name the covered range. Documentation-only closeout edits do not require another full CI; code or contract edits do. Preserve the review/refactor handoff rather than claiming it was already run.

Archive or remove only batch-created worktrees after comparing each child change set to the integrated result and preserving necessary ignored evidence. Confirm `git worktree list`, a clean `main`, commit hashes, and no batch server/process left running. Then return control to `$autoloop` for `/review` followed by `/refactor`; do not stop at a green CI summary.

## Standalone slice

Verify the spec's acceptance criteria and plan's completed tasks, update as-built/PROGRESS/lifecycle, and run focused checks plus `make ci` when needed for the final changed surface. Fix and rerun a failing gate before committing. Stage only this slice and use `feat: vN: <title>`. Report the commit and exact checks. If the user requested implementation only, do not infer permission to push.

When work is incomplete or mixed with unrelated dirty files, preserve it and report the exact blocker rather than making a partial closeout commit.
