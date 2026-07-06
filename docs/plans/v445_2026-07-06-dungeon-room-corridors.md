# v445 Plan — Structured room-and-hallway PCG

Status: Complete
Goal: Replace divider+scatter dungeon layout with room-graph PCG.
Architecture: Pack rectangular rooms, connect via MST+L corridors, emit `room_wall` perimeters with door gaps; disable interior scatter when PCG is on. Reuse reachability validator and corridor zones.

## File map
| Action | Path |
|--------|------|
| Create | `server/internal/game/dungeon_room_corridors.go` |
| Create | `server/internal/game/dungeon_room_corridors_test.go` |
| Modify | `dungeon_room_layout.go`, `dungeon_gen.go`, `dungeon_profiles.go`, `dungeon_generation_rules.go`, `rules.go` |
| Modify | `shared/rules/dungeon_generation.v0.json`, schema |
| Modify | `shared/golden/dungeon_obstacles.json`, schema |
| Modify | `tools/validate_dungeon_goldens.py` |

## Task 1 — Rules and types
- [x] `room_corridor_pcg` block in shared rules + schema
- [x] `RoomCorridorPCGRules` struct + validation
- [x] `rooms` on `generatedDungeonLevel`

## Task 2 — Generator
- [x] Room packing, MST, L-corridors, wall emission
- [x] Pipeline integration + disable scatter

## Task 3 — Tests and goldens
- [x] Unit tests for room walls, reachability, boss exemption
- [x] Regenerate `dungeon_obstacles` golden
- [x] Update `validate_dungeon_goldens.py`

## Final verification
- [x] `make validate-shared`
- [x] `go test ./internal/game/... -run 'RoomCorridor|DungeonObstacles'`
