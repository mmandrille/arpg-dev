# v472 Spec: Room-Corridor Anchor Fallback

Status: Complete
Date: 2026-09-28
Codename: `room-corridor-anchor-fallback`
Baseline: v471 `dungeon-kit-walls-floors` (commit `3c8bbdfd`)

## Problem

`GenerateDungeonLevel` fails for some seed/level pairs with
`could not place room-corridor layout after 96 attempts`. `Sim.ensureDungeonLevel` propagates the
error, so a session whose seed hits one of these floors can never enter it: a deterministic
progression blocker. The report-only wall-grid audit found 4/200 failing pairs; a wider sweep of
the `audit-NN` seeds found ~3% of floors in levels -1..-10 failing.

## Root cause

Stairs, the teleporter, and pre-layout chests ("anchors") are placed on the empty floor **before**
rooms exist, and stay fixed across all `room_corridor_pcg.max_attempts`. Each attempt only re-rolls
rooms around them. Two anchor geometries have no valid room solution, so all 96 attempts fail
identically — raising `max_attempts`, `hub_size_multiplier`, or `room_spacing` does not help:

1. **Anchor-room placement is infeasible** (every failure in the original audit set).
   Anchors are grouped by a 13-unit distance threshold, and each group must fit its own
   room of at most `room_size_max`, with `room_spacing + wall_thickness` padding around every
   room (10 units between room interiors). Anchor pairs 13–20 units apart sit in a dead zone: too
   far apart to share a room, too close for two padded rooms. The same happens when a chained
   group is larger than `room_size_max` (`roomContainingPoints` clamps to the max) or when the
   fixed spawn room blocks a nearby anchor.
2. **Anchor on the perimeter margin line is unreachable.** `stair_placement.margin_from_wall` and
   `teleporter_placement.margin_from_wall` (2.0) equal `room_corridor_pcg.margin_from_perimeter`
   (2.0). An anchor at x = width-2 forces its room's inner edge onto the anchor, putting the stair
   on the room wall. The egress door opens onto a 1-unit gap between the room wall and the
   perimeter wall, which the nav grid cannot traverse.

## Decision

Add a second **anchor fallback pass** that runs only after the normal
`max_attempts` pass fails, using its own RNG stream (`|room_corridor_anchor_fallback|`):

- Anchor clusters that cannot be placed are merged with their nearest cluster (by minimum
  point-to-point distance, lowest index on ties) and placement restarts. Each merge strictly
  reduces the cluster count, so the loop always terminates.
- Fallback anchor rooms may exceed `room_size_max` up to the cluster's padded bounding box, and may
  sit flush with the perimeter wall (margin = `wall_thickness`), giving edge anchors interior
  clearance (`playerRadius + 0.1`, degrading to the existing 0 inset when that is geometrically
  impossible).
- Reachability validation, normal room packing, hub room, corridors, and loops are unchanged.

### Why a fallback pass instead of changing the primary path or shared rules

- **Bit-identical existing floors.** Every floor that generates today keeps the same layout, so
  there is no golden churn, no bot scenario drift, and no reshuffle of existing characters'
  seeded dungeons. Verified by fingerprinting 240 floors before/after: only the previously failing
  floors changed.
- **Not a tuning problem.** The failure is a feasibility gap between two placement stages, not a
  balance value. The data-level alternative — raising stair/teleporter `margin_from_wall` above
  `margin_from_perimeter + wall_thickness + player clearance` — only addresses cause 2, moves
  every stair on every floor, and still leaves cause 1. Fallback room clearance/margin are derived
  from existing values (`playerRadius`, `wall_thickness`), so no new hardcoded tuning value is
  introduced.
- **Architectural debt, stated plainly.** The real fix is to generate rooms first and then place
  anchors inside them. That changes every layout and is deferred (see Non-goals).

## Acceptance

- The 8 pinned seed/level pairs (4 original audit failures + 4 perimeter-edge failures found by
  the wider sweep) generate, and their stairs/teleporters are inside a room.
- A seed sweep (`audit-00..23` x levels -1..-10, parallel) generates every floor; the sweep width
  can be raised with `ARPG_DUNGEON_SWEEP_SEEDS`.
- Fallback generation is deterministic (repeat generation gives identical rooms and walls).
- `shared/golden/dungeon_obstacles.json` unchanged; `make lint-determinism`, maintainability, and
  CODEMAP checks green.

## Non-goals

- Rooms-first generation (place anchors inside generated rooms).
- Moving the elite-objective chest into rooms (it is placed after the layout and may sit in a
  corridor; reachability still validates it).
- Changing `shared/rules/dungeon_generation.v0.json` values.
