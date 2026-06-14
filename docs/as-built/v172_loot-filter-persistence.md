# v172 As-Built - Loot Filter Persistence

Date: 2026-06-14
Status: Complete

## What shipped

- `ClientSettings` now persists `loot_filter_mode` in `user://settings.json`.
- Missing, malformed, or unknown saved modes normalize to `All`.
- `LootLabelFilter` can restore a mode label directly without cycling.
- `main.gd` initializes the loot filter from loaded client settings and saves after the existing
  `L` cycle input.
- The behavior remains display-only client state; server loot ownership, protocol, replay, and
  pickup rules are unchanged.

## Proof

- `client/tests/test_loot_label_filter.gd` covers valid mode restore and invalid fallback.
- `client/tests/test_client_bot.gd` covers `loot_filter_mode` parsing, save shape, and reload.
- `make client-unit`
- `make maintainability`
- `make ci`

## Deferred

- Category filtering and additional filter modes.
- Settings-panel controls for loot filtering.
- Account-synced settings.
