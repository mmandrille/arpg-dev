# v171 As-Built - Sorcerer Paladin Third Skills

Date: 2026-06-14
Status: Complete

## What shipped

- Added Sorcerer `arcane_barrage`, a tier-3 force projectile skill requiring `ligthing`.
- Added Paladin `sanctuary`, a tier-3 area defense buff requiring `holy_shield`.
- Added shared skill presentation metadata and English/Spanish text for both skills.
- Extended shared skill validation so the new class ownership and prerequisite links are guarded.
- Extended Go, Python, and Godot skill loader/panel tests for the expanded class catalogs.
- Added protocol bot proofs:
  - `make bot scenario=69_sorcerer_arcane_barrage.json`
  - `make bot scenario=70_paladin_sanctuary.json`

## Key decisions

- Reused existing `projectile_attack` and `area_stat_buff` capability types to keep the slice
  focused on player-visible class expansion without a new mechanics framework.
- Kept the shipped misspelled `ligthing` skill id as the Sorcerer prerequisite to avoid a broad
  compatibility rename.
- Rejected external Godot skill-tree/VFX plugins; existing shared-rule loaders and code-native
  presentation cover this scope.

## Deferred

- New skill capability types and more distinctive Sorcerer/Paladin mechanics.
- Passive skill trees, mana regeneration, and final skill balance.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestLoadRules|TestThirdClassSkillsRequirePrerequisites|TestClassSkillGates' -count=1`
- `.venv/bin/pytest tools/bot/test_skill_demo.py -q`
- `make client-unit`
- `make bot scenario=69_sorcerer_arcane_barrage.json`
- `make bot scenario=70_paladin_sanctuary.json`
- `make maintainability`
- `make ci`

Visual verification commands:

```bash
make bot-visual scenario=69_sorcerer_arcane_barrage.json
make bot-visual scenario=70_paladin_sanctuary.json
```
