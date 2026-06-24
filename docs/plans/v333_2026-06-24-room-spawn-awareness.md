# v333 Plan — room-spawn-awareness

Date: 2026-06-24  
Spec: [`v333_spec-room-spawn-awareness.md`](../specs/v333_spec-room-spawn-awareness.md)

## Tasks

- [x] Task 1: Record corridor-gap exclusion zones during `placeRoomLayout`.
- [x] Task 2: Block monster pack centers/members overlapping corridor zones.
- [x] Task 3: Bias elite objective chest toward pack leader, rejecting corridors.
- [x] Task 4: Add `room_cluster_radius` to shared rules + validation.
- [x] Task 5: Focused Go tests + `make validate-shared`.

## Verification

```bash
cd server && go test ./internal/game/... -run 'RoomSpawn|RoomLayout|EliteObjective|Pack'
make validate-shared
```
