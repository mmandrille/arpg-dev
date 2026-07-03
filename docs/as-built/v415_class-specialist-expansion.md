# v415 As-Built — Class Specialist Expansion

Date: 2026-07-03

## What shipped

- Added ten `class_specialist` templates (two per class): war bracers / iron girdle (barbarian),
  arcane robes / focus band (sorcerer), consecrated aegis / crusader plate (paladin), assassin
  gloves / silent treads (rogue), hawk amulet / pathfinder boots (ranger). Catalog is now 15
  specialists (three slots/classes minimum).
- Wired all new templates into `dungeon_mob_tc_depth_1` and extended `class_specialist_gear_lab`
  loot spawns.
- Server exposes `stat: "class"` rows with `class_id` in `requirement_status`; `requirements_met`
  and equip preview honor class locks.
- Client tooltips format class requirement lines via `ItemRequirementViews.format_requirement_status`.
- Fixed `items.v0.schema.json` to include `rogue` in `class_required` enum.

## Proof

```bash
make validate-shared
make validate-assets
cd server && go test ./internal/game/... -run 'ClassSpecialist|ClassRequirement' -count=1
godot --headless --path client --script res://tests/test_item_requirement_views.gd
make bot scenario=class_specialist_gear_lab
```

## Deferred

- Vendor/mystery specialist pools, unique/set specialist packages, production specialist art
- Affix grammar / multi-suffix naming for specialists
