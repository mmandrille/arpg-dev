# v80 As-built - Combat Threat Readability

Status: Complete

## Shipped

- Existing authoritative `monster_aggro` events now produce display-only `AGGRO` floating text on
  the aggroing monster.
- `DamageNumber` has a `threat` variant so aggro text is bot-visible and visually distinct from
  normal damage, miss, block, crit, heal, mana, and skill-reject text.
- Floating threat text respects the existing local `floating_combat_text` setting.
- A focused client-bot scenario starts in a generated dungeon pack, waits for `monster_aggro`, and
  asserts the `threat` damage number text.
- The v80 engineering review set was written and review cadence moved to v90.

## Proof

- `make client-unit`
- `SCENARIO=31_combat_threat_readability HEADLESS=1 ./scripts/bot_client_local.sh`
- `ARPG_ADDR=:8888 SCENARIO=pack_aggro_and_dungeon_packs ./scripts/bot_local.sh`
- `make ci`

## Deferred

- Persistent target indicators, sound, minimap danger hints, elite labels, elite auras, and richer
  pack-leader readability remain future work.
