---
name: execute
description: >-
  Implement one approved slice plan and verify its behavior. Use for /execute,
  requests to implement a plan, or the implementation stage in a slice session.
  Batch sessions use focused checks and hand off to the coordinator.
disable-model-invocation: true
---

# /execute — Implement one approved plan

Read `PROGRESS.md` current status, open gaps, and checklist, then `CLAUDE.md`, `docs/CODEMAP.md`, the plan/spec, and relevant ADRs. Review the plan before code: every acceptance criterion needs an owned task and a runnable proof; dependencies and the assigned base must still hold. Ask about material scope or contract changes that the approved plan does not cover. Continue with conservative fixes for minor drift already implied by the spec.

Implement in task order, normally shared contracts → authoritative server → bot proof → client presentation → as-built. Keep gameplay tuning in schema-backed shared data; preserve deterministic sim and server authority. Mark plan checkboxes as work is actually completed. New bot scenarios must be registered in the movement audit and CI pack only when merge-blocking coverage justifies it; update CODEMAP and asset manifests for new files. Capture the actual player camera for visual claims and matched measurements for performance claims.

## Verification mode

- **Batch slice session:** Run focused Go/Python/Godot tests and named bot or visual scenarios covering the changed behavior, plus relevant `make validate-shared`, `make validate-assets`, and `make maintainability`. Inspect the captures. Do not run `make ci`, `make ci-full`, `/finish`, commit, push, or edit the coordinator's `main`. Write `docs/as-built/vN_<codename>.md` with results and limits, then send the complete handoff defined in [the batch workflow](../autoloop/references/batch-workflow.md).
- **Standalone execution:** Run the plan's focused checks and the appropriate final gate. If the user asked only for implementation, leave commit to `/finish`; report exactly what passed. Do not claim a failed or skipped check passed.

If a check fails, diagnose and rerun the smallest relevant check while iterating. Keep the slice open when acceptance remains unproved, report the blocker, and let other batch sessions proceed. Do not turn a screenshot into a performance claim or mark checkboxes complete from static inspection alone.
