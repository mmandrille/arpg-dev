---
name: plan
description: >-
  Review an approved game slice spec and write its implementation plan under
  docs/plans. Use for /plan or the plan stage of an approved slice session.
  Documentation only; execution requires a passing spec review.
disable-model-invocation: true
---

# /plan — Spec review and implementation plan

Read `PROGRESS.md` current status, open gaps, and checklist, then `CLAUDE.md`, `docs/CODEMAP.md`, the assigned spec, and relevant ADRs/as-built code. In a batch session, keep the coordinator's reserved `vN` and recorded base; check dependency drift against the latest integrated state before implementation.

## Review gate

Check scope and non-goals; acceptance-to-test mapping; protocol/schema/golden changes; Go determinism; shared data ownership; server authority; replay; world presets; client asset/plugin adopt-borrow-reject; maintainability ratchet; and likely overlap with sibling slices. If a material spec gap or product decision remains, ask the user and pause this slice. Fix minor inconsistencies that follow directly from the accepted brief, then record the correction. Do not write a plan that silently changes the accepted scope.

## Plan

Write or update `docs/plans/vN_YYYY-MM-DD-<codename>.md` using the user's local date. Include:

- goal, baseline commit, prerequisite slices, and which tasks can proceed before they integrate;
- file map, ownership boundaries, shared-file conflicts, and integration/handoff notes;
- ordered tasks with checkboxes, paths, and the smallest runnable check for each;
- bot scenarios for gameplay/protocol/world/inventory/movement/combat/replay work, unless the spec explains a narrow deferral;
- real-renderer visual capture and performance comparison when presentation or smoothness is claimed;
- maintainability decision for touched over-limit files, data-driven tuning ownership, and final focused verification commands;
- as-built evidence and the complete changed/untracked/ignored-evidence handoff to the coordinator.

For an approved batch, the plan's final gate is **focused slice verification**. State that the coordinator runs combined `make ci` only after every accepted slice is integrated; do not put a per-slice `make ci` checkbox in the worker plan. For standalone work, retain `make ci` as the pre-commit gate when the change warrants it. Do not implement code in `/plan`.

After the review passes, the batch session continues to `/execute` under its accepted assignment. Otherwise report the plan path and wait for implementation authorization. See [the batch workflow](../autoloop/references/batch-workflow.md).
