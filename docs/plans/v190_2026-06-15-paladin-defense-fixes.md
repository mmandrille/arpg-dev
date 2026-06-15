# v190 Plan - Paladin Defense Fixes

Status: Complete
Goal: Make Paladin block, Holy Shield, and Sanctuary behavior match the in-game expectations.
Architecture: Keep combat outcomes server-owned. Shared skill data owns Sanctuary duration/radius/cooldown, Go sim owns immunity and derived stats, Godot renders only effect ids and combat text.
Tech stack: shared JSON schemas, Go sim, Python protocol bot scenario, Godot client unit tests.

## Baseline and Shortcut Decision

Builds on v81 Holy Shield and v171 Sanctuary. Godot plugin adoption checklist: reject external plugins/assets because the requested yellow dome is a small code-native marker using existing status-effect presentation helpers.

## Tasks

- [x] Expose effective `block_percent` in `DerivedStatsView` and add a regression assertion for shield block.
- [x] Refresh character progression changes when skill buffs start/end so Holy Shield updates derived stats immediately.
- [x] Convert Sanctuary from an area stat buff to an `area_immunity_buff` with radius 5, 60 tick duration, and 600 tick cooldown.
- [x] Add server damage immunity checks for monster attacks, projectiles, retaliation, and boss active phases.
- [x] Add `immune` combat outcome schema support and update the Sanctuary bot scenario timing.
- [x] Render Sanctuary as a yellow dome from `effect_ids` and add client unit proof.
- [x] Update lifecycle docs and run final verification.

## Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestRulesLoad|TestSanctuary|TestHolyShield|TestCombatStatBreakdownsIncludeEquipmentAndCap' -count=1`
- [x] `make client-unit`
- [x] `ARPG_ADDR=:8097 scenario=70_paladin_sanctuary.json ./scripts/bot_local.sh`
- [x] `make COMPOSE=true ci`
