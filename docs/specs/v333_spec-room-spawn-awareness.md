# v333 Spec — Room Spawn Awareness

Status: Approved  
Date: 2026-06-24  
Codename: `room-spawn-awareness`

## Purpose

Make v330 room-and-corridor dungeon layout affect spawn placement. Monster packs and elite
objective chests must avoid corridor gap zones so fights happen inside partitioned rooms, and
elite chests cluster in the same room area as the elite pack leader when possible.

## Non-goals

- No BSP room objects, corridor wall enclosures, or interactive doors in corridors.
- No boss-floor layout changes.
- No client or protocol changes.
- No loot-table, monster-count, or combat tuning changes.

## Acceptance criteria

1. Generated non-boss floors with `room_layout.enabled` record deterministic corridor-gap exclusion
   zones when dividers are placed.
2. Monster pack centers and pack members are rejected when they overlap a corridor exclusion zone
   (rule-derived clearance from `monster_placement.pack_member_radius`).
3. Elite objective chest placement rejects corridor zones and prefers positions within
   `elite_objective.room_cluster_radius` of the elite pack leader when one exists.
4. Reachability validation still passes for all generated targets.
5. Focused Go tests prove corridor avoidance and elite chest clustering semantics.
6. Existing dungeon generation and population tests remain green.

## Scope and files

- `server/internal/game/dungeon_room_layout.go` — corridor zone recording
- `server/internal/game/dungeon_generated_types.go` — corridor zone storage
- `server/internal/game/dungeon_gen.go` — pack placement guards
- `server/internal/game/dungeon_elite_objective.go` — chest clustering
- `server/internal/game/dungeon_profiles.go` — optional `room_cluster_radius` rule field
- `shared/rules/dungeon_generation.v0.json` + schema
- `server/internal/game/dungeon_room_layout_test.go`

## Test and bot proof

```bash
cd server && go test ./internal/game/... -run 'RoomSpawn|RoomLayout|EliteObjective|Pack'
make validate-shared
```

## Open questions and risks

- Corridor inference must stay deterministic and derived from placed divider gaps only.
