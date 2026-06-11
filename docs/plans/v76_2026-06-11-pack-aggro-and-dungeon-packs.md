# v76 Plan — Pack Aggro and Dungeon Packs

Status: Complete
Goal: Make generated dungeon fights spawn and activate as deterministic monster packs.
Architecture: Pack placement is shared-rule-driven and resolved inside deterministic dungeon generation. The server keeps pack identity internal and uses monster `assist_radius` for combat joins. Existing protocol events (`monster_aggro`) prove the behavior without schema changes.
Tech stack: Shared JSON rules/schemas, Go sim and dungeon generator, Python protocol bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v37 aggro-on-hit, v40 reachable dungeon obstacles, v52 ranged monster AI, and v72 monster visual catalog. No Godot plugin or asset shortcut applies; this slice does not add client UI/art and uses existing monster rendering.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.json` / schema | Add data-driven `assist_radius` for chase monsters. |
| Modify | `shared/rules/dungeon_generation.v0.json` / schema | Add pack count/size/spread rules. |
| Modify | `server/internal/game/rules.go` | Load and validate new rule fields. |
| Modify | `server/internal/game/dungeon_gen.go` | Generate 5-7 packs with 2-5 members and reachable close placement. |
| Modify | `server/internal/game/sim.go` | Use assist radius for on-hit joins. |
| Modify | `server/internal/game/game_test.go` | Unit/golden-style tests for packs and assist activation. |
| Add | `tools/bot/scenarios/42_pack_aggro_and_dungeon_packs.json` | Runtime proof that one pull aggros multiple monsters. |
| Modify | `tools/bot/run.py` | Add only minimal scenario assertion support if needed. |
| Modify | `PROGRESS.md`, `docs/as-built/v76_pack-aggro-and-dungeon-packs.md` | Finish lifecycle docs. |

## Task 1 — Shared Pack Rules

Files:
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `server/internal/game/rules.go`

- [x] Step 1.1: Add `assist_radius` to chase monster rules and validation.
- [x] Step 1.2: Add `pack_count`, `pack_size`, and `pack_member_radius` under `monster_placement`.
- [x] Step 1.3: Validate pack count/size ranges and member radius.

```bash
make validate-shared
```

## Task 2 — Deterministic Pack Generation

Files:
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Replace regular generated single placement with pack placement for non-boss floors.
- [x] Step 2.2: Preserve minimum monster guarantees by assigning required definitions into early pack slots.
- [x] Step 2.3: Keep chest monster bonus as extra members distributed into packs without violating 2-5 base size when possible.
- [x] Step 2.4: Add tests for 5-7 packs, 2-5 members, close spacing, reachability, and deterministic output.

```bash
cd server && go test ./internal/game/... -run 'TestGeneratedDungeonPack|TestAggroOnHit'
```

## Task 3 — Assist Radius Aggro

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Use `assist_radius` for combat joins between monsters.
- [x] Step 3.2: Keep passive `aggro_radius` unchanged for normal proximity acquisition.
- [x] Step 3.3: Cover near/far assist behavior in server tests.

```bash
cd server && go test ./internal/game/... -run 'TestAggroOnHit'
```

## Task 4 — Bot Scenario

Files:
- Add: `tools/bot/scenarios/42_pack_aggro_and_dungeon_packs.json`
- Modify: `tools/bot/run.py` if needed

- [x] Step 4.1: Add a generated dungeon scenario that pulls one pack member.
- [x] Step 4.2: Assert multiple `monster_aggro` events after the pull damages a generated monster.

```bash
make bot
```

## Task 5 — Lifecycle Docs and CI

Files:
- Modify: `docs/specs/v76_spec-pack-aggro-and-dungeon-packs.md`
- Modify: `docs/plans/v76_2026-06-11-pack-aggro-and-dungeon-packs.md`
- Add: `docs/as-built/v76_pack-aggro-and-dungeon-packs.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark the spec and plan complete.
- [x] Step 5.2: Add the v76 pack-aggro as-built and progress note without moving the existing v78 latest-completed status backward.
- [x] Step 5.3: Record deferred scope: elite leaders, role composition, and client aggro readability.

```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestDungeonMonsterGeneration|TestDungeonMonsterGenerationCreatesDeterministicPacks|TestAggroOnHit|TestChampionMonstersSpawnWithCommonMinions|TestGuardedChestGenerationGolden'`
- [x] `ARPG_ADDR=:8888 SCENARIO=pack_aggro_and_dungeon_packs ./scripts/bot_local.sh`
- [x] `make ci`
