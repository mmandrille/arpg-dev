# v333 As-Built — room-spawn-awareness

**Date:** 2026-06-24  
**Status:** Complete

## What was proved

Room-layout dungeon floors now record corridor-gap exclusion zones and keep monster packs and elite
objective chests out of those chokepoints. Elite chests sample positions near the pack leader within
`room_cluster_radius` before falling back to corridor-safe random placement.

## Key decisions

- Corridor zones are AABBs derived from divider gap centers plus `pack_member_radius` padding.
- Elite chest clustering uses leader-offset sampling rather than whole-floor random first pass.
- Combined corridor-pack-placement and room-chest-clustering into one slice per user direction.

## Verification

```bash
cd server && go test ./internal/game/... -run 'RoomSpawn|RoomLayout|EliteObjective'
make validate-shared
```
