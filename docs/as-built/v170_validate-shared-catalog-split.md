# v170 As-built — Validate Shared Catalog Split

Date: 2026-06-14

## What shipped

- Added `tools/validate_main_config.py` for `main_config` gameplay validation.
- Updated `tools/validate_shared.py` to delegate main-config gameplay checks while keeping shared
  file loading and later orchestration local to the entrypoint.
- Added a focused Python regression for an invalid `base_drop_rate_percent` value.

## Proof

- `.venv/bin/pytest tools/test_validate_shared.py -q` passed.
- `make validate-shared` passed with 1128 checks.
- `validate_shared.py` dropped from 3169 to 3140 lines, and its maintainability baseline was
  lowered from 3149 to 3140.

## Deferred

- Additional `validate_shared.py` domains, such as combat, dungeon generation, shops, worlds,
  item visuals, and item presentations, remain candidates for later focused helper modules.
