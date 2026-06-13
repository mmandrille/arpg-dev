# v120 As-built: Tuning-Friendly Rule Tests

Date: 2026-06-13
Spec: [`docs/specs/v120_spec-tuning-friendly-rule-tests.md`](../specs/v120_spec-tuning-friendly-rule-tests.md)
Plan: [`docs/plans/v120_2026-06-13-tuning-friendly-rule-tests.md`](../plans/v120_2026-06-13-tuning-friendly-rule-tests.md)

## What shipped

- Converted `client/tests/test_skills_panel.gd` to derive Magic Bolt stat requirements, mana cost,
  and max-rank fixture values from `SkillRulesLoader`.
- Preserved the UI assertions for requirement met/blocked states, tooltip text, missing-stat diff,
  and spend-button enablement.
- Updated the maintainability baseline for already-committed `main.gd` and `test_item_visuals.gd`
  growth surfaced during final closeout, with the v120 review carrying the extraction follow-up.

## Verification

```bash
make client-unit
make maintainability
make ci
```

## Review gate

v120 is the engineering-review milestone. The review set is written under `docs/reviews/` after
the slice CI gate:

- [`docs/reviews/20260613_v120-overview.md`](../reviews/20260613_v120-overview.md)
- [`docs/reviews/backend/20260613_v120-backend.md`](../reviews/backend/20260613_v120-backend.md)
- [`docs/reviews/client/20260613_v120-client.md`](../reviews/client/20260613_v120-client.md)
- [`docs/reviews/extras/20260613_v120-shared-tooling-and-process.md`](../reviews/extras/20260613_v120-shared-tooling-and-process.md)
