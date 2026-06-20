# v311 Spec - Character Base Stat Breakdown UI

Status: Complete
Date: 2026-06-20
Codename: character-base-stat-breakdown-ui

## Purpose

Make the character panel base-stat section show both allocated and effective stat values, and explain effective stats with the same source-by-source formula tooltip used by derived stats.

## Non-goals

- No formula or balance tuning changes.
- No new external assets, plugins, or UI dependencies; adopt/reject decision: reject external UI plugins and reuse the existing character-panel tooltip path.
- No redesign of stat allocation or the wider inventory/equipment UI.

## Scope And Likely Files

- `server/internal/game/types.go` - expose effective base stats in `character_progression`.
- `server/internal/game/base_stat_breakdowns.go` - build base-stat breakdown rows from base stats, equipment, set bonuses, passive skills, and active stat effects.
- `server/internal/game/sim.go` - include the new field and prepend base-stat breakdown rows.
- `shared/protocol/session_snapshot.v8.schema.json` and `shared/protocol/state_delta.v8.schema.json` - describe effective base stats and base-stat breakdown keys/source kinds.
- `client/scripts/character_stats_panel.gd` - render base/effective stat pairs and tooltip formulas for base stats; include effective stat values in derived stat formula source text.
- `client/tests/test_character_stats_panel.gd` and focused Go tests - prove UI formatting and server-authored breakdowns.

## Acceptance Criteria

- Base stats render as a compact table with `NAME`, `BASE`, and `EFFECTIVE` columns.
- Derived stats render as a compact table with `NAME` and `VALUE` columns.
- Hovering a base stat's `EFFECTIVE` value shows an opaque formula tooltip with one contribution per line and a final value.
- Effective values are visually emphasized: green when higher than base and red when lower than base.
- Equipment source text inside base-stat tooltips uses item names without item IDs or item source categories.
- Derived stat formula rows sourced from a base stat include the effective stat value in the parenthetical source text, e.g. `42 Strength, Character formula`.
- Protocol validation accepts base-stat breakdown rows and the new effective base stat object.

## Testing Plan

- Run the focused server test:
  ```bash
  cd server && go test ./internal/game -run TestCharacterProgressionViewEffectiveBaseStatsAndBreakdowns -count=1
  ```
- Run the focused Godot panel test:
  ```bash
  godot --headless --path client --script res://tests/test_character_stats_panel.gd
  ```
- Run shared protocol validation:
  ```bash
  make validate-shared
  ```
