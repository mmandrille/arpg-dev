# v120 Plan — Tuning-Friendly Rule Tests

Status: Complete
Goal: Convert a focused gameplay-tuning-sensitive test to derive expectations from shared rules.
Architecture: Keep production code unchanged. Use `SkillRulesLoader` in the Godot test to derive
the same skill requirements and presentation-adjacent values the UI consumes.
Tech stack: Godot headless unit tests, SDD docs, review close-out.

## Task 1 — Skills Panel Tuning Lock

Files:
- Modify: `client/tests/test_skills_panel.gd`

- [x] Step 1.1: Load skill rules through `SkillRulesLoader` in the test.
- [x] Step 1.2: Replace hardcoded Magic Bolt requirement/cost/max-rank expectations with helpers
  derived from `shared/rules/skills.v0.json`.
- [x] Step 1.3: Preserve UI assertions for tooltip text, spend enablement, and rank requirement
  blocking.

```bash
make client-unit
```

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/tests/test_skills_panel.gd`

Decision:
- [x] Defer extraction with rationale: this is a narrow test-only change that adds small local
  helpers to remove duplicated tuning values. Extracting a shared GDScript test utility would be
  larger than the slice; record a baseline update if the ratchet requires it.
- [x] Documented maintenance exception: the final ratchet also surfaced already-committed growth in
  `client/scripts/main.gd` and `client/tests/test_item_visuals.gd` from the preceding UI/loot work.
  v120 updates those grandfathered baselines to the current line counts so future growth is again
  constrained, and the v120 review records `main.gd` service-panel extraction as a next-batch input.

## Task 2 — Lifecycle Docs and Review Gate

Files:
- Modify: `docs/specs/v120_spec-tuning-friendly-rule-tests.md`
- Modify: `docs/plans/v120_2026-06-13-tuning-friendly-rule-tests.md`
- Create: `docs/as-built/v120_tuning-friendly-rule-tests.md`
- Modify: `PROGRESS.md`
- Create: `docs/reviews/20260613_v120-overview.md`
- Create: `docs/reviews/backend/20260613_v120-backend.md`
- Create: `docs/reviews/client/20260613_v120-client.md`
- Create: `docs/reviews/extras/20260613_v120-shared-tooling-and-process.md`

- [x] Step 2.1: Run focused and full verification.
- [x] Step 2.2: Write v120 as-built and update lifecycle status.
- [x] Step 2.3: Write the v120 engineering review set after CI is green.
- [x] Step 2.4: Update `PROGRESS.md` review pointers to v120.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`
