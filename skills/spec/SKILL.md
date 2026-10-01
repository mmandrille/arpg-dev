---
name: spec
description: >-
  Draft or revise one SDD slice spec under docs/specs. Use for /spec, $spec,
  or the spec stage of an approved slice session. Writes documentation only.
---

# /spec — Slice specification

Read `PROGRESS.md` current status, open gaps, and checklist, then `CLAUDE.md`, `docs/CODEMAP.md`, the accepted brief, and relevant ADRs/as-built notes. In a batch session, use the `vN`, codename, dependencies, and base commit assigned by the coordinator; do not recalculate or reserve another number. Standalone: check existing specs, plans, and the lifecycle index before choosing `vN`, and update an existing draft rather than duplicating it.

Write `docs/specs/vN_spec-<codename>.md` with `- **Status:** To Do`, date, purpose, non-goals, observable acceptance criteria, likely shared/server/client/bot/docs surfaces, focused verification, bot and real-camera proof where relevant, dependencies, integration risks, and only the open questions that affect planning. Client UI, camera, inventory, or art specs must inspect in-repo Godot code and asset manifests first and record an adopt / borrow / reject decision or require it in the plan. Prefer data-owned gameplay tuning and server-owned outcomes.

This skill does not implement or write a plan. Resolve blocking product decisions with the user; correct minor details supported by the accepted brief. In a batch session, leave the spec in the assigned worktree and proceed to `/plan` under the accepted `$autoloop` assignment once the spec review is sound. Outside a batch, report the file and stop unless the user also authorized later stages. Read [the batch workflow](../autoloop/references/batch-workflow.md) for handoff boundaries.
