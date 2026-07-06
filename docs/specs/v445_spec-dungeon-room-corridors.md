# v445 — dungeon-room-corridors

**Status:** Complete  
**Date:** 2026-07-06  
**Codename:** dungeon-room-corridors

---

## Purpose

Replace the v330 divider walls plus v40 interior obstacle scatter with a deterministic
**room-and-hallway** floor plan for normal generated dungeon levels. Each floor places a
configurable set of non-overlapping rectangular rooms, connects them with open L-shaped
corridors (minimum spanning tree plus optional loop edges), and emits perimeter walls with
door archways at room–hallway joins. Interior line/L/T/block wall scatter is disabled when
structured PCG is active.

Boss floors remain on the fixed v35 arena layout.

---

## Non-goals

- Non-rectangular rooms, rotated walls, secret/destructible doors
- Enclosed corridor tunnel walls (hallways are open floor between room walls)
- Interactive doors at room connections (open archways only)
- Room typing (treasure/shrine/boss rooms), room-aware encounter scripting beyond spawn hints
- Client renderer changes
- Durable map persistence across sessions
- Final biome/difficulty balance

---

## Acceptance criteria

1. When `room_corridor_pcg.enabled` is true, normal non-boss floors generate at least the
   configured minimum room count with `room_wall` perimeter segments and connecting corridor zones.
2. v330 `room_layout` dividers are not emitted when structured PCG is enabled.
3. Interior obstacle scatter (`line`/`L`/`T`/`block` groups) is disabled or capped to zero when
   `disable_obstacle_scatter` is true.
4. Water and hole floor features may still generate when enabled.
5. `validateGeneratedDungeonReachability` passes for stairs, teleporters, chests, doors, monsters,
   and loot on sample seeds/levels.
6. Same `session_seed + level` produces identical layout (`make lint-determinism`, replay).
7. `reachable_dungeon_obstacles` bot scenario completes; unit tests cover room count, reachability,
   and boss-floor exemption.
8. `shared/golden/dungeon_obstacles.json` updated for the new layout contract.

---

## Scope and files likely touched

| File | Change |
|------|--------|
| `server/internal/game/dungeon_room_corridors.go` | New — room packing, MST corridors, wall emission |
| `server/internal/game/dungeon_room_corridors_test.go` | New — unit tests |
| `server/internal/game/dungeon_room_layout.go` | Pipeline gate: PCG vs legacy dividers |
| `server/internal/game/dungeon_gen.go` | Skip obstacle scatter when PCG disables it |
| `server/internal/game/dungeon_profiles.go` | `RoomCorridorPCGRules` + validation |
| `server/internal/game/dungeon_generation_rules.go` | Wire new rules field |
| `server/internal/game/dungeon_generated_types.go` | `rooms` on generated level |
| `server/internal/game/rules.go` | Validation wiring |
| `shared/rules/dungeon_generation.v0.json` | `room_corridor_pcg`; disable `room_layout` |
| `shared/rules/dungeon_generation.v0.schema.json` | Schema block |
| `shared/golden/dungeon_obstacles.json` | Updated fixture |
| `shared/golden/dungeon_obstacles.v0.schema.json` | Add `room_wall`, `corridor_wall` sources |
| `server/internal/game/dungeon_room_layout_test.go` | Update for PCG sources |
| `docs/plans/v445_2026-07-06-dungeon-room-corridors.md` | Plan |
| `docs/as-built/v445_dungeon-room-corridors.md` | As-built on finish |

---

## Test and bot proof

- Unit: room count, `room_wall` presence, reachability across seeds/levels, boss floor unaffected,
  legacy `room_layout` when PCG disabled.
- Golden: `dungeon_obstacles_golden_test.go`.
- Bot: extend `reachable_dungeon_obstacles` (pack member).

---

## Open questions and risks

| # | Question | Default |
|---|----------|---------|
| Q-1 | Replace dividers when PCG on? | Yes — mutually exclusive |
| Q-2 | Door type at connections? | Open archways |
| Q-3 | Interior scatter? | Disabled via rules flag |
| Q-4 | Hub room? | One larger room when enabled |

**Risks:** generation retry budget on tight floors; golden drift; stair/chest pre-placement must
fall inside rooms or reachable corridors.
