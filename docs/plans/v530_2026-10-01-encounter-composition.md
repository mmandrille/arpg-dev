# v530 Plan — Encounter composition by room role

**Status:** Complete — integrated; combined `make ci` passed 2026-10-02 (21m23s).
**Date:** 2026-10-01
**Base:** `876d02c872db4b465e92f460266f83ea6245c3c8`
**Dependency:** v529 is present as an uncommitted overlay on that base.

## Goal and ownership

Consume v529's canonical room-index/role candidate rows to place the existing complete pack-size
multiset into normal dungeon rooms using shared, validated composition rules. Preserve the total
population, pack identities, elite/guard relationships, objective behavior, fixed boss-floor path,
and protocol contracts.

v529 owns the room candidate contract and role weights. Its
`allocateDungeonRoomPopulationBudgets(rng, rules, level)` returns eligible room rows in canonical
`level.rooms` order, including zero-count candidates. `RoomIndex` resolves against that room list;
`MonsterCount` starts at zero. v530 owns pack-size assignment, member roles, elite/guard composition,
and room-local placement; it writes final per-room counts after successful placement. There is no
separate v529 integration commit: both dependency and implementation overlays are based on
`876d02c872db4b465e92f460266f83ea6245c3c8`.

The unchanged seeded pack-size multiset includes the guarded-chest bonus when generated. v530 assigns
each occurrence once, never splits a pack, and keeps packs within one eligible room. Shared tuning
belongs under `monster_placement.encounter_composition`; v529's
`monster_placement.room_role_weights` and v528's `room_corridor_pcg.room_roles` remain intact.

## Implementation decisions

- Validate encounter templates, pack/member bounds, role constraints, and elite/guard rules from the
  shared schema and semantic rule validator. Keep server authority for room, count, pack, elite,
  guard, collision, and reachability data.
- Use named seeded random streams for composition and objective placement.
- Place each pack transactionally using room interiors, body clearance from corridors, obstacles,
  spread bounds, and reachability checks. Retry placement with bounded seeded attempts; return a
  useful generation error if no complete assignment fits.
- Reserve the optional elite-objective chest location before pack placement. If the reservation
  makes the assignment impossible, retry with the same pack-size multiset and without the optional
  reservation; that floor then has no objective chest.
- Keep boss floors on their fixed existing generation path. Do not add client inputs or protocol
  fields.
- Asset decision: adopt the already vendored KayKit Dungeon kit where visuals are needed, borrow
  existing wall/entity presentation and rectangle contracts, reject new external assets, plugins,
  and pipelines. No client presentation change was required here.

## Changed surfaces

v530 adds `server/internal/game/dungeon_encounter_composition.go`, its focused tests, and
`server/internal/game/dungeon_generation_reachability.go`; modifies generation, objective, rule
validation, shared rules/schema, generation tests, CODEMAP, and
`tools/bot/scenarios/14_dungeon_monsters.json`. It overlaps v529 in generated room data, placement,
shared rules, and tests. The coordinator must preserve the candidate API and integrate this slice
after v522–v529.

## Focused verification

- Focused Go suite covering population, encounter composition, room roles/shapes, objectives, and
  monster generation: PASS.
- Deep-floor reachability/determinism and elite objective focused suite: PASS.
- Shared validation via the existing venv: PASS, 2,271 checks. The `make validate-shared` wrapper
  could not provision its venv because the configured index lacked `setuptools>=68`.
- CODEMAP validation and `make maintainability`: PASS.
- Bot scenarios `14_dungeon_monsters`, `28_reachable_dungeon_obstacles`, and
  `77_elite_minion_pack_ai`: PASS. Scenario 14 uses a test-only debug progression fixture so the
  character survives the generated packs; this is protocol-flow evidence, not a balance claim.
- `git diff --check`: PASS. Bot checks used `scripts/bot_local.sh` with the already provisioned
  workspace environment because wrapper provisioning has the same package-index limitation.
- Combined `make ci`: not run in this slice worktree; coordinator-owned after integration.

Exact focused Go command:

```bash
cd server && go test ./internal/game -run 'TestDungeonRoomPopulation|TestDungeonEncounter|TestDungeonRoomRole|TestGeneratedDungeonAnchorsFollowRoomRoles|TestDungeonFloorProfilesGenerate|TestDungeonRoomThresholdDoors|TestDungeonRoomShapes|TestEliteObjective|TestDungeonEliteObjective|TestDungeonMonsterGeneration' -count=1 -timeout 8m
```

## Closeout gate

No worker commit or combined CI. The coordinator compares the v530 diff with the current v522–v529
overlay, resolves overlapping generation files, checks the as-built inventory and ignored evidence,
then runs the integrated batch gate and owns lifecycle updates.
