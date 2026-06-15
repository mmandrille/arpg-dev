# v185 As-Built - Paladin Defense Fixes

Date: 2026-06-15

## What Shipped

- Effective shield block now flows through `derived_stats.block_percent`, fixing the character derived stat panel for Paladins using shields.
- Skill buff start/end now emits character progression updates when the active player's effective stats change, so Holy Shield's armor/block bonus appears and expires without waiting for unrelated stat changes.
- Sanctuary is now a data-driven `area_immunity_buff`: radius 5, 60 ticks active, 600 tick cooldown, effect id `sanctuary`.
- Incoming player damage checks Sanctuary immunity before applying damage from monster melee, monster projectiles, retaliation, or boss active hit phases.
- The client renders Sanctuary as a translucent yellow dome and displays `IMMUNE` for zero-damage immune combat events.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRulesLoad|TestSanctuary|TestHolyShield|TestCombatStatBreakdownsIncludeEquipmentAndCap' -count=1`
- `make client-unit`
- `ARPG_ADDR=:8097 scenario=70_paladin_sanctuary.json ./scripts/bot_local.sh`
- `make COMPOSE=true ci`

## Deferred

- Production-quality Sanctuary VFX/audio.
