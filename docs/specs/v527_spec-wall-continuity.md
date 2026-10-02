# v527 Spec — Wall Continuity

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
**Date:** 2026-10-01
**Codename:** wall-continuity
**Baseline:** `876d02c872db4b465e92f460266f83ea6245c3c8` (v521)
**Dependencies:** v522 and v523; implementation waits until both are integrated into this worktree.

## Purpose

Strengthen correctness of generated dungeon walls where rectangular room perimeters, corridor
openings, floor edges, and solid obstacles meet. Tests must detect unintended walkable leaks and
intended openings blocked by rectangle overlap or clearance errors, using the same player-sized
collision semantics as authoritative server movement and reachability.

The server continues to own collision, door/opening semantics, and generated-target reachability.
The emitted wall rectangles remain the authoritative contract consumed by the existing renderer.

## Non-goals

- No polygonal or non-rectangular wall geometry, new wall protocol shape, or client-side collision
  authority.
- No changes to room/corridor layout policy, encounter placement, door gameplay, or obstacle tuning
  except a correction required to satisfy the stated continuity contract.
- No nondeterministic generation, new RNG source, or changes to seed/stream ownership.
- No new shared tuning values unless the implementation demonstrates a data-owned policy is needed.
- No external art, plugin, dependency, or asset pipeline.
- No broad renderer or visual redesign.

## Acceptance criteria

1. Focused deterministic geometry tests cover sealed room corners and perimeter joins, corridor
   openings, and room/corridor/solid-obstacle intersections using the existing rectangular AABB
   wall representation and player collision radius.
2. The tests fail for both classes of defect: an unintended traversable route through a wall join,
   and an intended opening or connecting corridor made impassable by adjacent/overlapping rectangles.
3. Generated normal dungeon floors remain reachable from the established start to all targets
   accepted by the generator's reachability contract across a deterministic representative seed /
   level set, including the v522/v523 behavior after integration.
4. Repeated generation with the same seed and level yields identical wall rectangles and continuity
   results. Generation remains on the seeded PCG path with stable ordering.
5. Server simulation remains the only authority for wall blocking and generated target reachability;
   client wall data remains rectangle-based and retains the renderer's current fields and meanings.
   No protocol/schema change is expected.
6. The `wall_floor_dungeon_rollout` visual route runs once dependencies are available. It confirms
   rendered wall layout remains present; it is not treated as proof of collision continuity.
7. Focused Go checks pass. No full CI is run in the slice worktree.

## Likely surfaces

- Server generation and geometry: `server/internal/game/dungeon_room_perimeter_walls.go`,
  `dungeon_room_corridors.go`, and only the specific join/clearance helper implicated by the
  post-dependency inspection.
- Server tests: prefer a focused new test file such as
  `server/internal/game/dungeon_wall_continuity_test.go`; reuse existing generated-floor
  reachability and determinism helpers rather than changing unrelated tests.
- Existing bot / visual proof: `tools/bot/scenarios/client/79_wall_floor_dungeon_rollout.json`
  and the existing renderer path; add no new scenario unless the existing proof cannot exercise
  the integrated generated layout.
- Documentation: this spec, the implementation plan, then as-built handoff evidence and registries
  only when implementation is complete and evidence is available.

## Asset decision

- **Adopt:** the already vendored KayKit Dungeon kit only if a focused generated-wall visual is
  needed.
- **Borrow:** existing wall/entity presentation and the authoritative rectangle input contract.
- **Reject:** new external assets, plugins, dependencies, and asset pipelines.

## Focused verification

Expected server checks after dependency integration (refine against the final touched files):

```bash
go test ./internal/game -run 'WallContinuity|RoomCorridorLayout|Reachability'
go test ./internal/game -run 'RoomCorridorLayout_AnchorFallbackDeterministic|WallContinuity'
make bot-visual scenario=wall_floor_dungeon_rollout
```

The visual scenario is a renderer/input regression check. Geometry and traversability claims must
come from deterministic server tests. The coordinator owns combined `make ci` after all batch
slices are integrated.

## Risks and review notes

- v522/v523 may change the exact generation seams under test; integrate both before selecting
  implementation sites or fixing any expected geometry.
- Grid reachability alone can miss sub-cell gaps or overstate doorway clearance. Tests must use the
  production player-radius/AABB predicates for geometric assertions, alongside the existing
  generated-target reachability validator.
- A focused fixture can pass while seed-specific joins still fail. Include both hand-authored
  geometric edge cases and deterministic generated seeds.
- This changes authoritative gameplay geometry. Preserve server ownership and avoid trusting any
  client-derived walkability result.
