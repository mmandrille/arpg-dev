# v311 As-Built - Character Base Stat Breakdown UI

Status: Complete
Date: 2026-06-20

## Summary

- Added `effective_base_stats` to the server-authored character progression view and v8 protocol schemas.
- Added base-stat breakdown rows for `str`, `dex`, `vit`, and `magic`, sourced from allocated base stats, equipment, set bonuses, passive skills, and active stat effects.
- Updated the character panel to display base stats as a compact `NAME` / `BASE` / `EFFECTIVE` table.
- Updated derived stats to display as a compact `NAME` / `VALUE` table.
- Reused the opaque formula tooltip path on base-stat `EFFECTIVE` cells, with item-provided stat contributions shown by item name only.
- Styled effective stat values as visually heavier text, green when higher than base and red when lower than base.
- Updated derived-stat formula source text so character-formula rows for Strength, Dexterity, Vitality, or Magic include the effective stat value in parentheses.

## Verification

```bash
cd server && go test ./internal/game -run TestCharacterProgressionViewEffectiveBaseStatsAndBreakdowns -count=1
godot --headless --path client --script res://tests/test_character_stats_panel.gd
make validate-shared
cd server && go test ./internal/game -count=1
make client-unit
make maintainability
```

All checks passed on 2026-06-20.
