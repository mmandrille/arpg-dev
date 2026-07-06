# v445 As-Built — dungeon-room-corridors

**Date:** 2026-07-06  
**Status:** Complete

## What was proved

Normal generated dungeon floors now use structured **room-and-hallway PCG** instead of v330 cross-floor dividers plus v40 interior wall scatter. The generator packs 5–7 rectangular rooms (optional hub), connects them with an MST plus 1–2 loop edges, carves L-shaped open corridors, and emits `room_wall` perimeter segments with door archways. Interior line/L/T/block scatter is disabled when PCG is active; water and holes still generate. Reachability validation and corridor zones for spawn avoidance are unchanged.

## Key decisions

- **Algorithm:** Rejection-sampled room packing → Prim MST → L-corridor zones → perimeter walls with sorted door gaps. Mutually exclusive with legacy `room_layout` dividers.
- **No corridor tunnel walls:** Hallways are open floor between room perimeters (matches spec non-goals).
- **Golden contract:** `validate_dungeon_goldens.py` branches on `room_corridor_pcg.enabled` (room_wall count vs scatter shape families).

## Deferred

- Room typing, enclosed corridor walls, interactive doors at connections
- Room-aware loot/encounter scripting beyond corridor-zone spawn hints
- Anchor-forced room placement around pre-placed stairs (reachability retry suffices)
