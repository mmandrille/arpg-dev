# v330 As-Built — dungeon-rooms

**Date:** 2026-06-23  
**Status:** Complete

## What was proved

Room-based PCG dungeon layout replaces the open-space obstacle scatter. Each non-boss floor now contains 2–4 cross-floor wall dividers (a mix of horizontal and vertical) that partition the space into visually distinct areas connected by 2-tile-wide open corridor gaps. Interior obstacle scatter (line/L/T/block shapes) still runs inside the partitioned areas, and the obstacle density ceiling was raised from 15 to 25 groups.

## Key decisions

- **Algorithm**: Cross-floor divider walls with gap carving (not BSP). Each divider spans 65–85% of the relevant floor dimension and gets 1–2 corridor gaps of width 2.0. The implementation reuses the existing `wallObstacle` struct and BFS reachability validator — no new data types or wire format changes.
- **`finalizeGeneratedDungeonLevel` extraction**: The duplicated 18-line population sequence (obstacles → monsters → elite chest → water → holes → reachability) was extracted into a shared helper in `dungeon_room_layout.go`. This brought `dungeon_gen.go` from 1085 → 1055 lines (back at baseline).
- **Elite objective chest bug fixed**: `randomObjectiveChestPosition` now clears the same `interactableClearance` (6.0 tiles) from all placed monsters, matching the constraint used during monster placement. Previously the chest could land within 6 tiles of a monster placed before it, which would cause `TestDungeonMonsterGeneration` to report a false blocked position when the final level was checked.
- **`dungeon_obstacles_golden_test.go` writer fixed**: `MinimumGeneratedWallCount` was being computed as `len(walls) - 4` (total minus perimeter), but room_divider walls also have non-"generated" source. The writer now uses the actual counted value.

## Files changed

| File | Change |
|------|--------|
| `server/internal/game/dungeon_room_layout.go` | New — `placeRoomLayout`, `finalizeGeneratedDungeonLevel`, corridor-gap carving |
| `server/internal/game/dungeon_room_layout_test.go` | New — 6 unit tests |
| `server/internal/game/dungeon_gen.go` | Extracted to helper; −24 net lines |
| `server/internal/game/dungeon_profiles.go` | `RoomLayoutRules` struct + validation |
| `server/internal/game/dungeon_elite_objective.go` | Elite chest now clears monster positions |
| `server/internal/game/dungeon_obstacles_golden_test.go` | Fixed MinimumGeneratedWallCount writer; added `generatedCount` variable |
| `server/internal/game/rules.go` | `RoomLayout` field wired in two struct definitions |
| `shared/rules/dungeon_generation.v0.json` | `room_layout` block added; obstacle max raised 15→25 |
| `shared/rules/dungeon_generation.v0.schema.json` | `room_layout` schema block |
| `shared/golden/dungeon_obstacles.json` | Updated (wall count 16→21; room_dividers included) |
| `shared/golden/dungeon_obstacles.v0.schema.json` | Added `"room_divider"` to source enum |
| `.maintainability/file-size-baseline.tsv` | `main.gd` 5825→5937 (v329 camera exception); `dungeon_gen.go` 1061→1055 |

## Deferred

- Vertical-only / horizontal-only layout edge cases (min_dividers_total guarantee provides a floor)
- Monster pack placement awareness of corridor width (packs can still technically straddle a corridor; reachability validator prevents outright trapping)
- Room-aware loot/chest clustering within rooms
