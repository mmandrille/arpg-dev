# v102 As-Built: Class Bot-Visual Scenarios

Date: 2026-06-12
Codename: `class-bot-visual-scenarios`

## What shipped

- Added `paladin_class_foundation`, `barbarian_class_foundation`, and `sorcerer_class_foundation`
  protocol bot scenarios with visual replay metadata.
- Kept `rogue_class_foundation` as the fourth class-foundation scenario under the same coverage
  rule.
- Added rule-derived Python coverage that fails when a playable class lacks a
  `{class_id}_class_foundation` scenario or when a class skill is not referenced by that scenario.
- Updated bot character selection to prefer a named class bot character over the default local
  `Hero` character, so Barbarian class scenarios use the v97 starter axe loadout instead of a
  compatibility/default character.

## Proof

Targeted protocol proof:

```bash
make bot scenario=paladin_class_foundation
make bot scenario=barbarian_class_foundation
make bot scenario=sorcerer_class_foundation
make bot scenario=rogue_class_foundation
make bot scenario=paladin_class_foundation,barbarian_class_foundation,sorcerer_class_foundation,rogue_class_foundation
```

Visual proof:

```bash
make bot-visual scenario=paladin_class_foundation
make bot-visual scenario=barbarian_class_foundation
make bot-visual scenario=sorcerer_class_foundation
```

Focused Python proof:

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

## Notes

- No protocol schema or shared gameplay rule changes were needed.
- Existing Godot shutdown RID/texture leak warnings still print during visual replay, but all three
  `make bot-visual` commands completed successfully.
