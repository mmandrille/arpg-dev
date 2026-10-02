# v529 Spec — Room Population

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Date:** 2026-10-01
- **Codename:** room-population
- **Base:** `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Dependencies:** v528 room roles; v522–v524 room layout and role contracts. Coordinator transfer
  also includes v525 corridor routing and v527 wall continuity.

## Purpose

Distribute the existing area-derived ordinary-floor monster population among usable generated
rooms. Room usable area and the room role assigned by v528 determine the relative placement weight.
The resulting floor should use its rooms for encounters while preserving the configured monster
total, pack constraints, seeded determinism, and existing spawn-clearance and reachability rules.

## Non-goals

- No change to the area formula, its min/max bounds, the guarded-chest population bonus, or boss-floor
  population.
- No new monster definitions, pack-composition rules, rarity/loot/stat tuning, AI behavior, or
  difficulty rebalance.
- No protocol, persistence, replay-format, client-rendering, asset, plugin, or external pipeline
  changes.
- No monsters in corridors or room passages; no population of floors without the v528 generated-room
  contract.

## Behavior and invariants

- Compute the same ordinary-floor target as today: the rule-derived population plus the existing
  guarded-chest bonus. The planned pack sizes must sum to that target and retain the configured pack
  count and size bounds whenever those bounds admit a solution. Existing champion minions remain
  their current separate composition behavior.
- For each generated room, compute usable area from the v528 room mask after monster clearance and
  corridor/anchor/blocker exclusions. Exclude zero-capacity rooms. Allocate the unchanged ordinary
  monster target deterministically by usable area multiplied by its schema-backed
  `monster_placement.room_role_weights` entry for the exact v528 room role.
- Emit stable room candidate rows in canonical room order containing room index, role ID, configured
  role weight, and an initially zero monster count. The index refers to canonical
  `generatedDungeonLevel.rooms` ordering so v530 can resolve room geometry.
- Use the exact role vocabulary supplied by v528. Every supported role has a schema-validated
  population weight; a missing/unknown role must not silently acquire an arbitrary weight. Preserve
  `room_corridor_pcg.room_roles` and its distinct role-assignment weights.
- Keep room-role weighting separate from pack/member selection. v530 owns whole-pack assignment,
  role-keyed composition, elite/guard relationships, and in-room placement; it preserves the
  original pack-size multiset and floor total while writing final room counts after placement succeeds.
- If weighted allocation cannot use a room, redistribute that share across remaining eligible rooms
  in a stable order. If all configured monsters cannot be placed without violating the invariants,
  generation returns a contextual error instead of dropping monsters or placing them in corridors.
- Preserve seeded PCG determinism. Any new random decisions use a named dungeon-generation substream
  keyed by seed and level; no map iteration affects ordering. Do not use `math/rand` or wall-clock
  state.
- Retain existing server-side collision/blocker, wall margin, spawn-anchor distance, protected
  interactable clearance, and generated reachability validation. The client remains presentation
  only, and generated floors remain reproducible during replay.

## Acceptance criteria

1. For ordinary floors with room generation and role assignment enabled, the candidate provider
   returns eligible room rows in canonical room order with the configured role weight. v530 uses
   usable area × `monster_placement.room_role_weights[role]` to prefer rooms when assigning complete
   packs; zero-weight roles remain excluded.
2. The target remains derived from the unchanged area formula plus the existing guarded-chest bonus.
   Existing pack-size generation continues to use the same input RNG state and configured pack
   count/size bounds; no configured population is silently omitted.
3. Repeated generation with the same seed and level returns identical room candidates, pack
   assignments, and monster placements; a distinct seed can vary assignment without changing the
   target total.
4. Room usable area and `monster_placement.room_role_weights` affect pack-assignment preference as
   configured, using rule-derived assertions.
5. Usable-area measurement excludes room-mask clearance, corridors, protected anchors, blocking
   obstacles, and unreachable candidate centers. Monster positions still pass the existing corridor,
   obstacle, anchor, door, and reachability checks. v530 owns assigning complete packs to room
   candidates and writes each room's final count only after placement succeeds.
6. Boss floors and generation with room layout/roles disabled preserve their existing population
   path.
7. Shared rule data validates against its schema; no wire contract or golden needs to change.

## Security and authority

This is server-side procedural gameplay generation, not a security-sensitive application surface.
The server continues to own collision, reachability, encounter population, and generated outcomes;
the client cannot submit or override room assignments. No new input, endpoint, persistence, secret,
or dependency surface is introduced.

## Asset decision

- **Adopt:** already vendored KayKit Dungeon kit only if visual inspection is needed to interpret
  existing room presentation; this slice does not add or alter visuals.
- **Borrow:** existing wall/entity presentation and rectangle/room contracts from the repository.
- **Reject:** new external assets, plugins, or asset pipelines.

## Likely implementation surfaces

- Shared: `shared/rules/dungeon_generation.v0.json` and
  `shared/rules/dungeon_generation.v0.schema.json` for role weights, using the exact v528 role enum;
  `tools/validate_shared.py` only if shared-data invariants need an additional domain check.
- Server: `server/internal/game/dungeon_gen.go`, a focused room-population helper under
  `server/internal/game/`, `server/internal/game/dungeon_generated_types.go` only if room assignment
  needs transient generated data, and focused dungeon generation tests.
- Bot: existing `12_dungeon_levels`, `14_dungeon_monsters`, and `28_reachable_dungeon_obstacles`
  scenarios for end-to-end authority/reachability; add a scenario only if existing assertions cannot
  observe the placement behavior without incidental movement.
- Docs: plan and as-built evidence; update `docs/CODEMAP.md` if a new module/test is added.

## Focused verification

- `make validate-shared`
- Focused Go tests covering deterministic allocation, rule-derived totals and pack bounds, usable
  room/role weighting, corridor and anchor exclusions, room-layout-disabled behavior, and boss-floor
  behavior (exact test selector recorded in the plan/as-built after implementation).
- `cd server && go run ./cmd/determinism-lint ./internal/game/...`
- `make bot scenario=14_dungeon_monsters`
- `make bot scenario=28_reachable_dungeon_obstacles`
- `make bot scenario=12_dungeon_levels` if changes alter generation sequencing or target validation.
- Do not run `make ci` or `make ci-full` in this batch slice session.

## Review notes and integration risk

The assigned base commit does not contain the prerequisite work, but the coordinator has transferred
v522–v525, v527, and v528 changes into this detached worktree for execution. The integrated role
vocabulary is `entry`, `transition`, `combat`, `reward_objective`, and `boss_arena`. Roles live on
`dungeonRoom` beside shape-cell geometry (`innerMin`, `innerMax`, `shapeCells`); corridor zones and
selected routes live on `generatedDungeonLevel`. `finalizeGeneratedDungeonLevel` creates rooms and
role-assigned anchors before obstacles and monster placement. `DungeonRoomRoleRule.Weight` controls
role assignment; v529 population weights are the distinct `monster_placement.room_role_weights`.

The integrated contract was refined by v530: v529 supplies canonical room-index/role candidates and
role weights; v530 assigns the unchanged pack-size multiset to those candidates and writes final
per-room counts after complete packs are placed. This replaces v529's initial aggregate preallocation,
whose body-center estimate did not prove that complete formations fit. The original total and every
pack-size occurrence remain unchanged. The additive shared-rule fields are
`monster_placement.room_role_weights` (v529) and `monster_placement.encounter_composition` (v530);
preserve `room_corridor_pcg.room_roles`. v526 adds loop-route doors after monsters/water/holes in
root's current integrated ordering, so population/door clearance requires a combined integration
check against that state; this detached worktree predates v526 and does not import its changes.
