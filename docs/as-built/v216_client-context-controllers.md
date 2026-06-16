# v216 As-Built — Client Context Controllers

Date: 2026-06-16

## Shipped

- Extracted town/interactable primitive construction into `TownNodeFactory`.
- Extracted boss health bar, phase countdown, active boss lookup, and telegraph marker presentation into `BossVisualsController`.
- Added `BossVisualsContext` as a narrow data carrier for live entity state and explicit tint/status callables.
- Added direct preload tests for `TownNodeFactory`, `BossVisualsContext`, and `BossVisualsController`.

## Proof

- `godot --headless --path client --script res://tests/test_factories.gd`
- `make client-unit`

## Notes

- `main.gd` no longer owns town construction or boss visual coordination; it wires context references and keeps model tint/status refresh implementation local.
- Wall rendering now lazily initializes its renderer from `walls_root`, preserving lightweight tests that instantiate `main.gd` without running `_ready`.
