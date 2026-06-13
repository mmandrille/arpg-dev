# v133 Spec: Unique Validation Split

Status: Complete
Date: 2026-06-13
Codename: `unique-validation-split`

## Purpose

Move named-unique catalog validation out of the large shared validator body into a focused helper
module with unit coverage. The behavior must stay the same, but future named unique work should have
a smaller validation surface and clearer parity with Go rule loading.

## Non-goals

- No new unique items, unique effects, stats, drop odds, or chest contents.
- No protocol, replay, persistence, or client UI changes.
- No broad rewrite of `tools/validate_shared.py`.
- No split of every unique-effect validator; this slice only handles named unique item catalog
  validation.

## Acceptance Criteria

- `make validate-shared` still validates `shared/rules/unique_items.v0.json` and emits the existing
  named-unique success/failure semantics.
- Named unique validation lives in a focused Python helper rather than inline in
  `tools/validate_shared.py`.
- Python unit tests cover bad named unique cases for mismatched ids, unknown base templates, invalid
  enabled/status combinations, duplicate/unknown/inactive/incompatible fixed effects, missing fixed
  stats, and weak behavior hooks.
- Go rule tests cover the same representative invalid named unique cases through `LoadRules`.
- The split does not change gameplay output, current valid rules, or purple chest behavior.
- `make maintainability` and `make ci` pass.

## Likely Files

- `tools/validate_shared.py`
- `tools/validate_unique_items.py` or an equivalent focused helper under `tools/`
- `tools/test_validate_shared.py` or a focused validator test file under `tools/`
- `server/internal/game/unique_chest_test.go`
- `PROGRESS.md`
- `docs/as-built/v133_unique-validation-split.md`

## Test And Bot Proof

- `make validate-shared`
- `.venv/bin/python -m pytest tools -q`
- `cd server && go test ./internal/game -run 'TestLoadRules|TestNamedUnique'`
- `make maintainability`
- `make ci`

No new bot scenario is required because this is a validation and maintainability slice with no
runtime gameplay behavior change. Existing CI bot coverage remains the regression proof for unchanged
purple chest and unique-effect behavior.

## Open Questions And Risks

- None blocking. Keep the helper interface intentionally small: accept already-loaded rule catalogs
  and a report object, then record failures without owning file I/O.
