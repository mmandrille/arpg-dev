# v330 — dungeon-rooms

**Status:** Draft  
**Date:** 2026-06-23  
**Codename:** dungeon-rooms

---

## Purpose

Replace the current uniform random-scatter obstacle layout with a room-and-corridor floor plan.
Each dungeon floor now contains 2–4 cross-floor divider walls that partition the open space into
visually distinct rooms/areas, each divider having 1–2 open corridor gaps (2 tiles wide, no door
object). Interior obstacle scatter (existing line/L/T/block groups) continues to run inside the
partitioned areas, increasing effective wall density. The combined result: structured rooms, clear
walkable corridors, and more walls overall — playable immediately with the existing client renderer.

---

## Non-goals

- BSP-formal room placement with tracked room objects (deferred to future slice if needed)
- Non-rectangular room shapes
- Corridor walls enclosing hallways (gaps through dividers are corridors; no extra flanking walls)
- Interactive doors or locked corridors (open passages only this slice)
- Boss floor changes (boss floors use a fixed layout and are unaffected)
- Client changes (the `wallObstacle` wire format already renders correctly)
- Water and hole feature changes

---

## Acceptance criteria

1. **Room dividers present**: every non-boss dungeon floor generates 2–4 cross-floor divider walls
   (mix of horizontal and vertical), each with at least one 2-tile-wide open corridor gap.
2. **Reachability**: stairs, teleporters, chests, doors, and monsters all remain reachable from
   player spawn. `validateGeneratedDungeonReachability` passes before the level is committed.
3. **Interior scatter**: the existing obstacle scatter still fires after room layout completes,
   adding internal walls/rocks/columns within partitioned areas.
4. **Increased density**: effective obstacle group count target raised from max 15 to max 25
   (formula driven; data-only change in `dungeon_generation.v0.json`).
5. **Determinism**: same seed+level always produces identical layout (seeded RNG only, no
   `time.Now()`). `make lint-determinism` and `make replay` pass.
6. **Existing CI green**: `make ci` passes unmodified; no new golden deltas required (floor layout
   does not affect combat, loot, or golden formula fixtures).
7. **Bot scenario**: bot scenario `01_dungeon_basic` (or equivalent) completes without timing out;
   the bot reaches down stairs through at least one corridor gap.

---

## Algorithm

### Room layout pass (new: `dungeon_room_layout.go`)

1. Choose divider counts from rules: `N_h` horizontal and `N_v` vertical dividers (each 0–2,
   at least one non-zero).
2. For each horizontal divider:
   - Pick a random Y position in the interior (margin from top/bottom perimeter).
   - Pick a random X span covering 65–85% of the floor width, centered with margin from left/right.
   - Carve 1–2 corridor gaps (2 tiles wide) at random X positions within the wall span;
     gaps must be at least `min_gap_separation` tiles apart.
   - Produce two or more wall segments (identical to `splitWallForGeneratedDoor` but no door
     interactable; source = `"room_divider"`).
3. Repeat for vertical dividers (swap X/Y roles).
4. After placing all dividers, run `validateGeneratedDungeonReachability`. On failure, retry with
   a new RNG attempt (up to `max_attempts` in rules).
5. Add all produced wall segments to `out.walls` with `source = "room_divider"`.

### Obstacle scatter pass (existing, unchanged)

- `placeDungeonObstacles` runs after room layout; groups land within the now-partitioned space.
- Interior walls, rocks, columns, rubble clusters fill the rooms.

---

## Scope and files likely touched

| File | Change |
|------|--------|
| `server/internal/game/dungeon_room_layout.go` | **New** — `placeRoomLayout`, divider wall generation, corridor gap carving |
| `server/internal/game/dungeon_gen.go` | Call `placeRoomLayout` before `placeDungeonObstacles` in the main generation flow |
| `server/internal/game/dungeon_profiles.go` | Add `RoomLayout RoomLayoutRules` field to `DungeonGenerationRules`; add validation |
| `shared/rules/dungeon_generation.v0.json` | Add `room_layout` block; raise `target_group_count_formula.max` from 15 to 25 |
| `shared/rules/dungeon_generation.v0.schema.json` | Add `room_layout` schema block |
| `server/internal/game/dungeon_room_layout_test.go` | **New** — unit tests for divider placement, gap carving, reachability post-layout |
| `docs/plans/v330_2026-06-23-dungeon-rooms.md` | Plan |
| `docs/as-built/v330_dungeon-rooms.md` | As-built (on finish) |
| `PROGRESS.md` | Lifecycle update (on finish) |

No client, shared/protocol, golden formula, or bot scenario file changes expected.

---

## Rules schema additions (`room_layout`)

```json
"room_layout": {
  "enabled": true,
  "max_attempts": 32,
  "horizontal_dividers": { "min": 1, "max": 2 },
  "vertical_dividers":   { "min": 0, "max": 1 },
  "min_dividers_total":  2,
  "wall_span_ratio":     { "min": 0.65, "max": 0.85 },
  "corridor_width":      2.0,
  "min_gap_separation":  6.0,
  "corridors_per_wall":  { "min": 1, "max": 2 },
  "margin_from_perimeter": 5.0
}
```

`wall_span_ratio` is relative to the relevant floor dimension. All values are data-driven; no
tuning requires code changes.

---

## Test and bot proof

- **Unit** (`dungeon_room_layout_test.go`): dividers present, gap widths correct, wall source is
  `"room_divider"`, reachability passes for fixed seeds; coverage of 0-horizontal + 1-vertical
  and 2-horizontal + 0-vertical boundary cases.
- **Existing dungeon tests** continue to pass (reachability validator unchanged).
- **`make validate-shared`** passes with updated schema.
- **Bot** (`make bot scenario=01_dungeon_basic` or equivalent): bot reaches down stairs through
  corridor gap within scenario time budget.
- **CI** (`make ci`): all nine gates green.

---

## Open questions and risks

- **Monster placement under increased density**: with 25 obstacle groups + room dividers, monster
  placement `max_attempts` (512) may need a small raise if spawn positions are frequently blocked.
  Verify during implementation; if needed, raise to 768 in rules JSON (no code change).
- **`dungeon_gen.go` line count**: file is already large. `placeRoomLayout` goes in a new
  dedicated file; call site in `dungeon_gen.go` is one line per flow branch. Monitor baseline.
