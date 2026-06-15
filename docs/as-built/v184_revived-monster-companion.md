# v184 As-Built: Revived Monster Companion

Date: 2026-06-15

## Shipped

- Added Sorcerer `revive` as a `revive_companion` skill with data-driven rank power:
  - rank 1: 50% HP/damage
  - each additional rank: +10%
  - one active revived companion for this slice
- Revive targets dead monster entities (`hp: 0`) and rejects:
  - missing/non-monster targets
  - living monsters
  - boss entities
- Revived monsters become server-owned `companion` entities:
  - owned by the caster
  - source skill set to `revive`
  - original `monster_def_id` preserved for client rendering
  - original dead monster entity consumed
  - loot table forced to `no_drop`
- Existing companion AI drives follow and deterministic melee attacks.
- Bot runtime can now resolve monster targets with `alive: false` for corpse/dead-entity scenarios.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'Revive|Companion' -count=1`
- `.venv/bin/python -m pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=75_sorcerer_revive_companion.json`
- `make ci`

## Visual Check

```bash
make bot-visual scenario=75_sorcerer_revive_companion.json
```
