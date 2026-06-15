# v133 Plan — Unique Validation Split

Status: Complete
Goal: Extract named-unique catalog validation into a focused helper with Python and Go parity tests.
Architecture: Keep rule ownership in shared JSON and authoritative validation in both Python
shared-contract checks and Go `LoadRules`. The Python helper receives already-loaded rule catalogs and
the existing report object, preserving CLI output while shrinking `tools/validate_shared.py`.
Tech stack: Python validator/tests, Go rule tests, docs.

## Baseline And Shortcut Decision

does not touch client UI, presentation, art, camera, or inventory controls.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `tools/validate_unique_items.py` | Focused named-unique validation helper |
| Modify | `tools/validate_shared.py` | Delegate named-unique validation to helper |
| Modify/Create | `tools/test_validate_shared.py` or `tools/test_validate_unique_items.py` | Python validator regression tests |
| Modify | `server/internal/game/unique_chest_test.go` | Go invalid named-unique rule tests |
| Create | `docs/as-built/v133_unique-validation-split.md` | As-built summary |
| Modify | `docs/specs/v133_spec-unique-validation-split.md` | Status closeout |
| Modify | `docs/plans/v133_2026-06-13-unique-validation-split.md` | Checkbox closeout |
| Modify | `PROGRESS.md` | Lifecycle and next-slice update |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/validate_shared.py`
- [x] `server/internal/game/unique_chest_test.go` not touched; Go parity tests live in a new focused file

Decision:
- [x] Extract a focused helper from `tools/validate_shared.py`.
- [x] Keep Go tests in `unique_chest_test.go` unless the file crosses 600 lines; if it does, create
  a focused `unique_items_validation_test.go`.

Verification:
```bash
make maintainability
```

## Task 1 — Python Validator Split

Files:
- Create: `tools/validate_unique_items.py`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Move named unique validation logic into a helper function that accepts unique items,
  item templates, unique effects, and the existing report object.
- [x] Step 1.2: Preserve all existing success/failure labels and details where practical.
- [x] Step 1.3: Keep `tools/validate_shared.py` responsible for loading files and high-level flow.

```bash
make validate-shared
```

## Task 2 — Python Validation Tests

Files:
- Create/Modify: `tools/test_validate_unique_items.py`

- [x] Step 2.1: Unit-test valid named unique catalog input.
- [x] Step 2.2: Unit-test mismatched id, unknown base template, enabled/status mismatch, disabled
  status mismatch, weak behavior hook, missing fixed stats, duplicate fixed effect, unknown fixed
  effect, inactive fixed effect, and incompatible fixed effect.

```bash
.venv/bin/python -m pytest tools/test_validate_unique_items.py -q
```

## Task 3 — Go Rule Parity Tests

Files:
- Modify: `server/internal/game/unique_chest_test.go`, or create `server/internal/game/unique_items_validation_test.go`

- [x] Step 3.1: Add focused tests that mutate loaded rules or temp rule files to prove Go rejects
  duplicate, unknown, inactive, and incompatible fixed effects.
- [x] Step 3.2: Cover enabled/status mismatch and unknown base template.

```bash
cd server && go test ./internal/game -run 'TestNamedUnique|TestLoadRules|TestUniqueItemValidation'
```

## Task 4 — Lifecycle Docs And CI

Files:
- Create: `docs/as-built/v133_unique-validation-split.md`
- Modify: `docs/specs/v133_spec-unique-validation-split.md`
- Modify: `docs/plans/v133_2026-06-13-unique-validation-split.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the spec and plan complete.
- [x] Step 4.2: Record v133 completion, next slice, and deferred scope in `PROGRESS.md`.
- [x] Step 4.3: Add the v133 as-built summary.

```bash
make maintainability
make ci
```

## Final Verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `.venv/bin/python -m pytest tools/test_validate_unique_items.py -q`
- [x] `cd server && go test ./internal/game -run 'TestNamedUnique|TestLoadRules|TestUniqueItemValidation'`
- [x] `make ci`
