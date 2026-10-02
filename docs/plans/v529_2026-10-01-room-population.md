# v529 Plan — Room Population

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
- **Slice/base:** v529 `room-population`, `876d02c872db4b465e92f460266f83ea6245c3c8`
- **Prerequisites:** v528 room roles; v522–v524 layout/role contracts. The coordinator transferred
  v522–v525, v527, and v528 dependency changes into this detached worktree. v530 overlap is coordinated
  below.
- **Goal:** place the existing ordinary-floor population in room areas according to usable area and
  the v528 room-role contract, while preserving totals, pack constraints, determinism, clearances,
  and reachability.

## Spec review

- Scope is limited to the placement of ordinary generated monsters. The existing area formula,
  chest bonus, boss floors, pack composition, protocol, and client outcomes remain unchanged.
- Acceptance criteria map to server unit tests, shared validation, determinism lint, and existing
  dungeon bot scenarios. No new client visual scenario is required because this slice makes no
  presentation claim or change.
- Add only `monster_placement.room_role_weights` for v529. The exact v528 room-role weight remains
  distinct under `room_corridor_pcg.room_roles`; v530 owns the sibling
  `monster_placement.encounter_composition` field.
- The integrated handoff carries canonical `room index`, `role`, and role `weight` candidates. v530
  resolves geometry from `generatedDungeonLevel.rooms[index]`, assigns complete packs based on those
  weights and usable area, and writes final per-room counts after placement succeeds.
- Server-generated positions remain authoritative; seeded PCG substreams and stable ordering
  preserve replay reconstruction. No protocol/schema/golden change is anticipated.
- The client asset decision is recorded in the spec: adopt the vendored KayKit Dungeon kit only if
  visual inspection is needed, borrow current room/wall/entity and rectangle contracts, reject new
  external assets/plugins/pipelines.
- No product decision is required before planning. The missing dependency is a code-integration
  gate, not an open product question.

## File map and ownership

| Surface | Expected paths | Ownership / conflict |
|---|---|---|
| Shared rules | `shared/rules/dungeon_generation.v0.json`, `shared/rules/dungeon_generation.v0.schema.json` | Add only v529's `monster_placement.room_role_weights`, keyed to v528 roles. v530 owns sibling `monster_placement.encounter_composition`; merge both additive fields. |
| Server generator | New `server/internal/game/dungeon_room_population.go`; coordinated narrow call-site change in `server/internal/game/dungeon_gen.go` | Helper emits canonical room-index/role/weight candidates. Preserve existing population selection, collision, anchor, corridor, and reachability behavior. v530 assigns whole packs to candidates and owns final per-room counts. |
| Server tests | New focused `server/internal/game/dungeon_room_population_test.go`; extend existing generation tests only if a test must assert an established contract | Own semantic/rule-derived placement behavior; avoid growing grandfathered coordinators. Tests should remain independent of exact current tuning values except determinism equality. |
| Shared validation | `tools/validate_shared.py` and focused validator tests only if needed | Prefer schema constraints for shape; add semantic validation only for role-key parity or role-weight validity not expressible in JSON Schema. |
| Registry/docs | `docs/CODEMAP.md`, later `docs/as-built/v529_room-population.md` | Add new generator/test module to the dungeon-generation CODEMAP row. As-built captures actual proof and limits after execution. Coordinator owns `PROGRESS.md` and lifecycle closeout after integration. |

## Integrated contract amendment

The v530 implementation moved room assignment from aggregate body-center preallocation into complete-pack placement. The candidate API retains canonical room indices, roles, and configured weights, but its returned counts start at zero. v530 writes final counts only after all packs are placed, preserving the original pack-size multiset and total. This avoids treating raw body-center capacity as proof that a formation fits. This integrated contract supersedes earlier plan bullets describing fixed per-room budgets.

## Ordered tasks

### 0. Dependency integration gate — before code

- [x] Confirm the checkout remains detached at assigned base `876d02c872db4b465e92f460266f83ea6245c3c8`;
  coordinator-supplied dependency files are present alongside v529 docs.
- [x] Confirm the coordinator transferred v522–v525, v527, and v528 changes without replacing the
  v529 spec/plan. No dependency commit SHAs were supplied; inventory exact paths in the as-built.
- [x] Inspect the exact v528 role IDs (`entry`, `transition`, `combat`, `reward_objective`,
  `boss_arena`), room shape/bounds APIs, validation, and generator call order.
- [x] Confirm the stable candidate information: canonical room index, role, and configured role weight;
  v530 assigns complete packs and writes final room counts after successful placement.
- [x] If integrated dependencies cannot provide stable room roles and bounds, stop and report the
  exact contract gap; do not invent compatibility behavior.
- [x] Coordinate the v530 shared-rule boundary: v529 owns only
  `monster_placement.room_role_weights`; v530 owns `monster_placement.encounter_composition`.

### 1. Shared role distribution rules

- [x] Add only `monster_placement.room_role_weights`, keyed to each v528 room role; keep values
  non-negative with at least one positive value and validate the
  allowed range in `dungeon_generation.v0.schema.json`.
- [x] Ensure invalid keys, missing required roles, and invalid values are rejected by shared-rule
  validation. Update a focused `tools/test_validate_shared.py` test only if schema validation alone
  cannot enforce role parity.
- [x] Run `make validate-shared` (2269 checks passed; CODEMAP validation passed).

### 2. Deterministic room population helper

- [x] Add a focused helper file, keeping it below the 600-line maintainability target. Derive each
  room's eligible area from the integrated shape-cell mask after monster clearance and corridor,
  anchor, and blocker exclusions; reject zero-capacity rooms.
- [x] Allocate the unchanged ordinary target across eligible rooms using usable area ×
  `monster_placement.room_role_weights`. Return one record per eligible room in canonical order with
  room index, role, and assigned ordinary-monster count; counts sum exactly to the current target.
  Use a named seeded PCG stream and stable room ordering.
- [x] Reuse existing generation predicates for floor bounds, blocked walls/obstacles, player spawn
  distance, stairs/teleporters/chests, corridor zones, and reachability. Reassign a pack when a room
  lacks a valid candidate; fail with seed-independent contextual error text if the full target cannot
  be placed. Do not silently lower counts or fall back to corridor placement.
- [x] Keep boss-floor and room-layout-disabled paths unchanged. Preserve existing monster definition,
  rarity, and pack selection semantics, including champion-minion expansion; do not alter unrelated
  generation RNG streams.
- [x] Agree directly with v530 on the exact result type/signature before changing the generator call
  site. Ensure v530 can resolve canonical room geometry by index and writes final counts only after complete-pack placement succeeds.

### 3. Focused proof

- [x] Add focused tests for repeated seed/level equality; varied seed behavior with the same
  rule-derived target; target and chest bonus accounting; pack count/size bounds; usable-area and
  role-weight influence using a temp-rule fixture; corridor/inset/anchor exclusions; unreachable or
  blocked candidates; layout-disabled behavior; and boss-floor behavior.
- [x] Prefer semantic/rule-derived assertions; exact assertions only for identical-seed equality or
  formula/contract ownership. Keep tests in a focused file, not `game_test.go` unless a small seam is
  essential.
- [x] Run focused Go tests and `cd server && go run ./cmd/determinism-lint ./internal/game/...`.
  Focused Go tests pass; determinism lint exits 1 on inherited package baseline map-range findings.
- [x] Run `make validate-shared`.
- [x] Run `make bot scenario=14_dungeon_monsters` and
  `make bot scenario=28_reachable_dungeon_obstacles`; also run
  `make bot scenario=12_dungeon_levels` because population eligibility affects generation validation.
- [ ] Do not run `make ci`, `make ci-full`, `/finish`, commit, push, or modify coordinator `main`.

### 4. Handoff evidence

- [x] Update `docs/CODEMAP.md` for any new generator/test/validator files.
- [x] Add `docs/as-built/v529_room-population.md` with integrated dependency revisions, behavior,
  exact focused commands/results, seeded sweep/sample sizes if run, failed/limited evidence, and
  unverified acceptance criteria.
- [x] Record full changed/deleted/untracked paths, ignored evidence, likely shared-file conflicts
  (especially v528/v530), and all acceptance criteria not met. Leave `PROGRESS.md` and lifecycle
  closeout to the coordinator after integration.

## Focused verification commands

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestDungeonRoomPopulation' -count=1
cd server && go run ./cmd/determinism-lint ./internal/game/...
make bot scenario=14_dungeon_monsters
make bot scenario=28_reachable_dungeon_obstacles
make bot scenario=12_dungeon_levels
```

Run the `12_dungeon_levels` scenario only if the change touches shared generation ordering or
reachability validation; document skipped commands and reasons. No visual capture or performance
comparison is in scope because no visual or smoothness change is claimed. The coordinator runs
combined `make ci` only after all accepted slices are integrated.

## Integration and handoff risks

- v528's exact room role enum/rectangle semantics are not present at the base commit. Implementation
  is blocked until those dependency changes are integrated and inspected here.
- v530 encounter-composition may change `dungeon_gen.go` and
  `shared/rules/dungeon_generation.v0.json`; preserve both intended changes during merge and verify
  the final role-weight configuration against the combined schema.
- A configured target may exceed the capacity of a pathological room layout after obstacle and
  anchor clearances. v529 measures room weight using body and per-target clearances (while preserving
  full pack-spread clearance in the existing actual-placement predicate); v530 enforces full
  pack spacing while assigning complete packs to the weighted room candidates. The 24-seed × four-floor generation sweep passed 96/96
  cases after this separation and a small positive default transition-role weight.
- Before the body-clearance and transition-weight adjustment, `TestGeneratedDungeonTargetsReachable`
  exposed a no-usable-room error for `v40_reachability` floor -2. It is now covered by
  `TestDungeonRoomPopulationUsesTransitionCapacityWhenCombatRoomsAreBlocked`, and the original
  four-floor reachability test passes.
- `make bot scenario=14_dungeon_monsters` completed gameplay, reconnect, and replay assertions on
  two isolated runs but exceeded its 15-second scenario budget by 0.22 and 0.36 seconds. A concurrent
  first attempt also collided during PostgreSQL schema setup and is not counted as gameplay evidence.
- `make bot scenario=28_reachable_dungeon_obstacles` and `make bot scenario=12_dungeon_levels` passed.
- Monster placement changes are not proof of combat balance or player-visible quality. Existing bot
  scenario success proves only the exercised server/client protocol flow and its assertions.
