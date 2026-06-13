# v126 Spec - Skill Validation Split

Status: Complete
Date: 2026-06-13
Codename: `skill-validation-split`

## Purpose

Reduce `tools/validate_shared.py` ownership by moving skill catalog, presentation, prerequisite,
and skill golden cross-checks into a focused validator module while preserving the existing
`make validate-shared` behavior.

## Non-goals

- No rule/schema/content changes.
- No validation label changes unless needed for clarity.
- No broad rewrite of the shared validator.
- No gameplay tuning changes.

## Acceptance Criteria

1. Skill-specific validation lives in a dedicated Python module.
2. `tools/validate_shared.py` delegates skill checks to that module.
3. Existing skill validation failures remain covered, including class ownership, Magic Bolt tuning,
   presentation coverage, prerequisites, and skill golden parity.
4. Focused validator tests and `make validate-shared` pass.

## Likely Files

- `tools/validate_shared.py`
- `tools/validate_skills.py`
- `tools/test_validate_shared.py`
- `docs/as-built/v126_skill-validation-split.md`
- `PROGRESS.md`

## Test Proof

```bash
.venv/bin/pytest tools/test_validate_shared.py -q
make validate-shared
make ci
```
