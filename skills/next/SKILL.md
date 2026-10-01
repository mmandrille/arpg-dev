---
name: next
description: >-
  Propose one or more spec-ready game slices from PROGRESS.md, ADRs, and the
  user's idea. Use for /next, a request for the next slice, or a batch of slices
  to develop in separate sessions. Discovery only; no spec or code changes.
disable-model-invocation: true
---

# /next — Slice discovery and batch offer

Read `PROGRESS.md` current status, open gaps, and checklist first. Then read `CLAUDE.md`, relevant ADRs, active drafts, and the as-built notes needed to validate the baseline. Check `docs/CODEMAP.md` before broad code searches. Do not infer the next number from stale documents.

## Offer

- If the user supplied an idea or slice list, treat it as the primary candidate. Otherwise offer a small coherent group from the backlog and current player experience.
- For each candidate give a title/codename, player or system value, scope and non-goals, verifiable acceptance criteria, likely files/contracts, focused test or bot proof, size/risk, and asset/plugin adopt-borrow-reject decision to be recorded for client work.
- Show a dependency graph and likely shared-file conflicts. Mark which sessions can implement immediately and which can only spec/plan until a prerequisite is integrated. Respect existing accepted `To Do` specs and their numbers; assign new consecutive collision-free numbers provisionally until the user accepts the batch.
- Recommend an order and ask the user to accept the whole set or name the selected subset. Say explicitly that accepting a batch creates one separate user-visible session and temporary detached worktree for each selected slice. That clear approval authorizes dispatch by `$autoloop`; do not ask again per slice.

A single-slice request produces one spec-ready brief. Do not create a session, worktree, spec, plan, or code under `/next` alone. A batch request produces the batch offer; dispatch belongs to the main `$autoloop` workflow after acceptance.

If an engineering review is already due, state that fact. In a batch, schedule `/review` then `/refactor` after integration and the combined CI gate, as described in [the batch workflow](../autoloop/references/batch-workflow.md). For standalone work, follow the current review cadence in `PROGRESS.md`.

Ask only questions that affect acceptance, ordering, or architecture. Use a sensible stated default for minor choices and do not block the offer on optional polish.
