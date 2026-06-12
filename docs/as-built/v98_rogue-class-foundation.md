# v98 As-Built - Rogue class foundation

Date: 2026-06-12
Status: Complete

## What Shipped

- Added Rogue as the fourth playable class with dexterity-leaning shared progression stats.
- Added `starter_rogue_sword` and seeds new Rogues with common swords in both `main_hand` and
  `off_hand`, plus one `red_potion` and one `blue_potion`.
- Added Rogue-only off-hand one-handed melee weapon equip support while preserving non-Rogue
  off-hand weapon rejection and two-handed hand blocking.
- Added Rogue client presentation: fourth character-picker option, Rogue text keys, dagger class
  icon, and deterministic slimmer `character_rogue_v0` GLB model.
- Added protocol bot proof for creating a Rogue and observing the starter dual swords and potions
  through the normal snapshot path.

## Proof

- `make gen-assets`
- `make validate-shared`
- `cd server && go test ./internal/http -run TestCreatedCharactersReceiveClassStarterLoadouts`
- `cd server && go test ./internal/game -run 'TestLoadRules|TestRogueOffhandWeaponEquipRules|TestEquipmentWrongSlotRejects'`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make client-unit`
- `make bot scenario=47_rogue_class_foundation`
- `make maintainability`
- `make ci`

## Scope Limits

- Poison Stab DOT is not implemented yet.
- Dash movement/damage is not implemented yet.
- Off-hand independent attack timing and 1.5x off-hand attack speed are deferred; v98 establishes
  the class, starter kit, model, picker path, and equip contract first.

## Maintainability Note

This slice touched grandfathered server simulation and client test files in narrow, class/equipment
paths. The new behavior is data-backed where possible, with focused tests and a protocol bot
scenario covering the new class path.

`make maintainability` required a documented file-size baseline update for the focused Rogue edits
and for pre-existing drift surfaced by the checker (`main.gd`, `test_item_visuals.gd`, `rules.go`,
and the showme visual capture script). The ratchet now enforces growth from the v98 baseline.
