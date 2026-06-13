# v134 As-Built: Unique Inspection UI

Date: 2026-06-13
Status: Complete

## What Changed

- The Godot shared item rule loader now caches `shared/rules/unique_effects.v0.json` and exposes
  effect lookup by id.
- Added `UniqueEffectTooltip`, a small formatter that resolves item `effect_ids` into readable
  unique-effect title and summary lines.
- Inventory plain tooltips append unique-effect text after existing stats, requirements, and
  comparisons so the description lands at the bottom.
- Inventory, stash/unique chest, and market rendered item tooltips include the same readable effect
  block through the shared tooltip panel.

## Proof

- `godot --headless --path client --script res://tests/test_shop_panel.gd`
- `godot --headless --path client --script res://tests/test_stash_panel.gd`
- `make client-unit`
- `make maintainability`
- `make ci`

## Notes

- Server-authored summary payloads are unchanged. If a payload still includes raw `Effect: <id>`
  metadata, the client now adds the readable bottom description without changing that contract.
