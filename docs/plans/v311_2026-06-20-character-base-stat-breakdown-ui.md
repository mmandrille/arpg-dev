# v311 Plan - Character Base Stat Breakdown UI

Status: Complete
Goal: Show allocated/effective base stats in the character panel and give base stats source-rich formula tooltips.
Architecture: Server remains authoritative for effective stat values and source rows. The client only formats protocol data and reuses the existing custom tooltip implementation.
Tech stack: Go protocol view helpers, JSON schemas, Godot 4 GDScript panel/test.

## Baseline and asset/plugin decision

Builds on v310. Asset/plugin decision: reject external UI plugins and reuse the existing Godot label custom-tooltip path.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/types.go` | Add `effective_base_stats` to the progression view |
| Modify | `server/internal/game/sim.go` | Populate effective stats and include base breakdown rows |
| Create | `server/internal/game/base_stat_breakdowns.go` | Build base stat formula source rows |
| Modify | `server/internal/game/character_stats_test.go` | Focused server proof |
| Modify | `shared/protocol/session_snapshot.v8.schema.json` | Protocol shape |
| Modify | `shared/protocol/state_delta.v8.schema.json` | Delta protocol shape |
| Modify | `client/scripts/character_stats_panel.gd` | Base/effective rows and stat tooltip formatting |
| Create | `client/scripts/stat_tooltip_label.gd` | Shared opaque stat tooltip label |
| Create | `client/scripts/character_panel_styles.gd` | Shared character panel style builder |
| Modify | `client/tests/test_character_stats_panel.gd` | Focused panel proof |
| Create | `docs/as-built/v311_character-base-stat-breakdown-ui.md` | Completion proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines. Keep the new server helper focused rather than growing `sim.go` with breakdown assembly.

## Task 1 - Protocol and server data

- [x] Add `effective_base_stats` to `CharacterProgressionView`.
- [x] Add base-stat breakdown rows for `str`, `dex`, `vit`, and `magic`.
- [x] Update v8 protocol schemas for the new field, keys, and source kinds.
- [x] Add a focused Go test covering effective stats and source rows.

## Task 2 - Character panel rendering

- [x] Render base-stat rows as a `NAME` / `BASE` / `EFFECTIVE` table.
- [x] Render derived stat rows as a `NAME` / `VALUE` table.
- [x] Attach the same custom tooltip behavior to base-stat effective value cells.
- [x] Emphasize effective values and color boosted/reduced stats green/red.
- [x] Include effective stat values in derived stat character-formula source text.
- [x] Extend focused Godot coverage for stat rows and formula text.

## Final verification

- [x] `cd server && go test ./internal/game -run TestCharacterProgressionViewEffectiveBaseStatsAndBreakdowns -count=1`
- [x] `godot --headless --path client --script res://tests/test_character_stats_panel.gd`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -count=1`
- [x] `make client-unit`
- [x] `make maintainability`
