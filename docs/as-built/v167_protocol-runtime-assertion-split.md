# v167 As-built — Protocol Runtime Assertion Split

Date: 2026-06-14

## What shipped

- Added `tools/bot/runtime_economy_assertions.py` for runtime shop/stash economy assertions.
- Kept `run_runtime_assertions` as the public entrypoint and delegated handled economy assertions to
  the helper.
- Preserved existing protocol bot assertion syntax and helper behavior.

## Proof

- `.venv/bin/pytest tools/bot/test_protocol.py` passed with 63 tests.
- `tools/bot/runtime_assertions.py` dropped from 520 to 480 lines; the new helper is 45 lines.

## Deferred

- Movement, event/combat, inventory/equipment, world/teleporter, and progression runtime assertion
  domains remain candidates for later focused splits.
