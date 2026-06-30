# v387 As-Built — class weapon affinities

## What shipped

- Template-owned `class_affinities` roll ranges on five exemplar weapon families (rogue dagger,
  barbarian war hammer, heraldic shield non-paladin penalty, ranger compound bow reach, sorcerer staff
  max mana).
- Rolled affinities persist in `ItemRollPayload.class_affinities`; server exposes
  `class_affinity_status` on inventory, loot, shop, and stash views.
- Active affinities apply to authoritative derived stats (`damage_percent`, `attack_speed_percent`,
  `reach_percent`, `max_mana_percent`); inactive rows are display-only.
- Client tooltips show green/red class affinity lines (inventory, shop, stash, market fallback).
- Extended bot `109_class_weapon_affinities_lab` proves rogue active/inactive affinities and shield
  penalty stacking.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run ClassAffinity`
- `make client-unit`
- `make bot scenario=109_class_weapon_affinities_lab`

## Deferred

- Production art for affinity families (borrowed placeholder GLBs/presentations).
- Market REST listings still rely on client fallback when `class_affinity_status` is absent on wire.
