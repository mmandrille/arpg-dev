# v526 Spec - Room Doors

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01
- **Codename:** room-doors

## Purpose

Place a rule-limited, deterministic set of closed doors at eligible room-to-room thresholds in
generated dungeons. Reuse the existing authoritative wooden_door interactable, barrier, interaction,
and pathfinding behavior. Generated openings and their closed barriers must remain traversable after
interaction without making generated objectives or rooms unreachable.

## Non-goals

- No new door entity type, intent, protocol field, client-authored state, or persistence format.
- No locked/keyed, destructible, secret, one-way, or boss-floor doors.
- No new door model, asset, plugin, art pipeline, or rotated door presentation. Eligibility must
  respect the barrier orientation supported by the existing wooden_door.
- No changes to room graph topology, corridor geometry, encounter authority, or obstacle scatter.
- No claim of performance improvement or broad visual redesign.

## Acceptance Criteria

1. Schema-validated shared dungeon-generation rules own whether room-threshold doors are enabled
   and a maximum placement count. Invalid values fail rule loading.
2. Candidates are generated room-to-room thresholds represented by the accepted room layout.
   Perimeter-anchor egress openings and unrelated obstacle gaps are excluded.
   v525 retains normalized topology edges and routed endpoints with the selected configured width;
   selection uses only north/south thresholds supported by the horizontal door barrier, at most one
   threshold per route.
3. Eligible thresholds are selected deterministically through the existing seeded PCG path with
   stable candidate ordering. Identical seed, rules, and level produce identical room doors and
   wall layout.
4. Each selected threshold produces a normal authoritative wooden_door entity in the existing
   closed state and leaves its opening clear of wall collision.
5. Closed doors block movement at their threshold; the existing validated action flow opens them,
   publishes authoritative entity state, and permits traversal. Client input cannot set door state
   directly. Existing close behavior remains valid if interaction rules permit closing.
6. Generation reachability validation treats generated doors consistently with existing doors.
   Seeded multi-level generation and a bot path prove room connections and required anchors do not
   become softlocked. The proof shows a reachable approach side and passage after opening.
7. Client rendering and collision presentation follow existing authoritative wall/entity state. No
   client-side generation or speculative door state is introduced.
8. Existing non-room obstacle doors, room topology, replay determinism, and current door
   interaction behavior remain covered by focused checks.

## Scope and Likely Files

- Shared rules/schema: shared/rules/dungeon_generation.v0.json and
  shared/rules/dungeon_generation.v0.schema.json.
- Server generation and existing door state:
  - server/internal/game/dungeon_room_corridors.go
  - server/internal/game/dungeon_room_corridor_routing.go
  - server/internal/game/dungeon_room_perimeter_walls.go
  - server/internal/game/dungeon_generated_types.go
  - server/internal/game/dungeon_generation_rules.go
  - server/internal/game/dungeon_profiles.go
  - server/internal/game/rules.go
  - server/internal/game/dungeon_door_rules.go
  - server/internal/game/dungeon_doors.go
  - server/internal/game/dungeon_population.go
  - server/internal/game/interactables.go only if focused coverage finds missing behavior; prefer
    no changes to this shared interaction path.
- Tests/bot/evidence: focused tests under server/internal/game; shared/golden/dungeon_obstacles.json
  only if its deterministic contract is extended; a focused tools/bot/scenarios scenario (or an
  existing reachable-dungeon scenario if it can express the required behavior); client tests only
  if authoritative entity/wall-state rendering lacks coverage.
- Handoff docs: docs/as-built/v526_room-doors.md, docs/progress/slice-lifecycle.md, PROGRESS.md, and
  docs/CODEMAP.md as warranted.

### Asset decision

- **Adopt:** the already vendored KayKit Dungeon kit where visuals need it.
- **Borrow:** existing wall/entity presentation, wooden_door definition, and collision/picking
  contracts.
- **Reject:** new external assets, plugins, or rendering pipelines.

## Focused Verification

Final commands depend on the integrated v522/v525 file map. Expected focused coverage:

- make validate-shared
- cd server && go test ./internal/game -run 'Door|RoomCorridor|DungeonObstacles'
- make bot scenario=<room-door scenario id>
- make bot-visual scenario=wall_floor_dungeon_rollout after dependencies are integrated
- make maintainability if touched files or new tests affect the ratchet

The existing visual scenario renders a generated depth-1 wall/floor layout and checks basic wall
presence; it does not prove door collision, authoritative interaction, room connectivity, or
softlock freedom. Those require server tests and the bot path above.

The coordinator runs combined make ci only after all accepted batch slices are integrated.

## Dependencies and Integration Risks

- **Base:** 876d02c872db4b465e92f460266f83ea6245c3c8.
- **Prerequisites:** v522–v525 and v527 are present as coordinator-supplied, uncommitted worktree
  changes on the recorded base. The coordinator approved this spec and plan before implementation.
- v522–v525 retain roomCorridorRoute records with normalized topology edge, both room thresholds,
  selected corridor width, route points, and zones. Extend those records in place; do not rebuild
  topology or use a parallel edge model.
- v525 configures corridor widths as [1.5, 2.0], validates them against player diameter and the
  narrowest room-shape cell, and emits room openings at the selected route width. Keep door
  collision aligned to those openings and preserve the width contract.
- v527 tests doorway clearance for every room mask, outside wall-obstacle joins, and representative
  generated layouts for determinism and reachability. Preserve those fixtures and add generated-door
  assertions alongside them.
- Preserve existing room and obstacle deterministic goldens unless a deliberate generated layout
  change is required and explained. Do not change protocol contracts unless implementation proves
  them necessary and the coordinator coordinates that change.
- Doors are gameplay state: the server must validate target existence, active-level membership,
  reachability/proximity, and legal state transitions through the established action flow. Never
  accept a client-provided open/closed value.

## Open Questions

None block specification or planning. The planned rule shape is
room_corridor_pcg.doors.enabled/max_count with a derived seeded PCG selection stream.
