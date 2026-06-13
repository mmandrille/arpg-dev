# v126 As-Built - Skill Validation Split

Date: 2026-06-13
Spec: [`docs/specs/v126_spec-skill-validation-split.md`](../specs/v126_spec-skill-validation-split.md)
Plan: [`docs/plans/v126_2026-06-13-skill-validation-split.md`](../plans/v126_2026-06-13-skill-validation-split.md)

## What Shipped

- Added `tools/validate_skills.py` with `validate_skill_catalogs`.
- Moved skill class ownership, Magic Bolt tuning, skill presentation coverage, prerequisite, and
  skill golden parity checks out of `tools/validate_shared.py`.
- Added a focused regression proving the extracted validator reports unknown skill classes.

## Proof

- `.venv/bin/pytest tools/test_validate_shared.py -q`
- `make validate-shared`
- `make ci`
