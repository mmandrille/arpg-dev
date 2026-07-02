# v405 As-Built — Class Specialist Gear

Date: 2026-07-02

## What Shipped

- Added `equipment_category: class_specialist` and template `class_required` for five rolled archetypes:
  Skull Face (barbarian head, `damage_max`), Magic Book (sorcerer off-hand, mana sustain),
  Scepter (paladin 1H melee, `skill_damage_percent`), Shadow Mask (rogue head, evade),
  Hunting Quiver (ranger belt, attack speed).
- Specialist templates forbid `armor` in `base_stats` and `rollable_stats`; validation enforced in Go and `validate_shared.py`.
- Rolled items now honor template `class_required` at equip time (`class_requirement_not_met`).
- Fixed empty `base_stats` roll panic (`cloneIntMap` nil assignment) in item roll/reroll paths.
- Lab world `class_specialist_gear_lab`, depth-1 treasure class entries, presentation/visual mappings.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'ClassSpecialist' -count=1`
- `make bot scenario=class_specialist_gear_lab`

## Deferred

- Class requirement in `requirement_status` UI rows
- Unique/set specialist packages and affix grammar
- Production book/scepter/skull art
