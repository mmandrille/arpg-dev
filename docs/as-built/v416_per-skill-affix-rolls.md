# v416 As-Built — Per-Skill Affix Rolls

Date: 2026-07-03

## What shipped

- Added `random_skill_level` (~5%, rare+) to `helm` and `amulet`, and `random_class_skill_level`
  (~10%) to all `class_required` item templates. Bonuses roll `1..item_level` for a deterministic
  skill pick from the full catalog (jewelry) or class-filtered pool (specialist gear).
- Persisted `skill_level_bonuses` on `ItemRollPayload`; wired rolls, rerolls, world loot presets,
  and shop pricing.
- Reworked `effectiveSkillRank` to stack `baseRank + all_skills + perSkillBonus` without
  `MaxRank` or stat-requirement clamps on use; spend and cast paths keep spend-only requirements.
- Server-authored `skill_bonus_status` on item/stash/loot/shop views (green active, red wrong
  class, gray unlearned).
- Client tooltip lines via `skill_bonus_tooltip.gd` in inventory, stash, shop, and market panels.
- Equip/unequip now emits `skill_progression_update` so protocol rank reflects gear bonuses.
- Extended bot scenario `skill_level_affix_lab` with pinned `+3 magic_bolt` amulet proof.

## Proof

```bash
make validate-shared
make maintainability
cd server && go test ./internal/game/... -run 'SkillLevelBonus|EquipEmitsSkillProgression|AllSkillsBonusIgnores|CastSkillIgnores' -count=1
godot --headless --path client --script res://tests/test_skill_bonus_tooltip.gd
make bot scenario=skill_level_affix_lab
```

## Deferred

- Skill formula scaling audit for gear-pushed ranks above `MaxRank` (v417 follow-up)
- Procedural affix names, crafting, production art
