# v133 As-Built — Unique Validation Split

Date: 2026-06-13

## What shipped

- Extracted named unique item catalog validation from `tools/validate_shared.py` into
  `tools/validate_unique_items.py`.
- Preserved the existing `make validate-shared` labels and success/failure semantics for valid
  named unique rules.
- Added focused Python tests for valid catalogs and representative bad named unique definitions:
  mismatched ids, unknown templates, invalid statuses, weak behavior hooks, missing fixed stats, and
  duplicate/unknown/inactive/incompatible fixed effects.
- Added Go rule-loading parity tests that mutate temp rule catalogs and prove `LoadRules` rejects
  invalid named unique packages.

## Proof

- `.venv/bin/python -m pytest tools/test_validate_unique_items.py -q`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestNamedUnique|TestLoadRules|TestUniqueItemValidation'`
- `.venv/bin/python -m pytest tools -q`
- `make maintainability`
- `make ci`

## Deferred

- Broader extraction of the remaining `tools/validate_shared.py` cross-check sections.
- Unique-effect tooltip inspection remains the next player-facing slice.
