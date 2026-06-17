# v252 Plan - Expanded Dungeon Profile

Status: Complete
Goal: Make ordinary generated dungeon floors larger and more populated through depth-banded shared rules.
Architecture: Add shared floor-profile data that overlays ordinary dungeon generation before
placement begins. Keep the authoritative generator deterministic and server-owned; do not change
wire protocol or client authority. Use a new focused helper/test file so touched grandfathered
generator files stay within the maintainability ratchet.
Tech stack: shared JSON/schema, Go simulation, Python protocol bot, docs.

## Baseline and Shortcut Decision

Builds on v251 direct auto-navigation and the current multi-level dungeon generator. Existing
client generated wall rendering and ground materials are reused.

Asset/plugin decision: reject external assets/plugins for this slice. Borrow the existing in-repo
ground/wall generated material pipeline and bot wall-layout scenarios; defer biome colors and river
visuals to later dedicated slices.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add depth-banded ordinary-floor profiles for size, monster count, pack count, and obstacle group count |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate profile structure |
| Modify | `server/internal/game/rules.go` | Load the profile list in `DungeonGenerationRules` |
| Add | `server/internal/game/dungeon_profiles.go` | Apply matching profile overlays without growing the generator coordinator |
| Modify | `server/internal/game/dungeon_gen.go` | Hook profile application into generation/navigation |
| Add | `server/internal/game/dungeon_profiles_test.go` | Focused profile selection, deterministic generation, reachability, and boss-floor exclusion tests |
| Modify | `docs/specs/v252_spec-expanded-dungeon-profile.md` | Mark status complete when shipped |
| Modify | `docs/progress/slice-lifecycle.md` | Add v252 lifecycle row |
| Add | `docs/as-built/v252_expanded-dungeon-profile.md` | Record shipped behavior and deferrals |
| Modify | `PROGRESS.md` | Update latest slice and deferred dungeon-generation scope |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/dungeon_gen.go`
- [x] `server/internal/game/rules.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none planned
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Extract focused helper/test file as part of this slice: `dungeon_profiles.go` and
  `dungeon_profiles_test.go`.

Verification:
```bash
make maintainability
```

## Task 1 - Shared Profile Data

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`

- [x] Add `floor_profiles` to the schema with depth bounds and overlayable generation values.
- [x] Add an ordinary-floor profile for depth 4+ non-boss floors while keeping entry depths 1-3
  stable for existing exact goldens and staying inside the current client presentation footprint.
- [x] Validate schema and shared cross-checks.

```bash
make validate-shared
```

## Task 2 - Server Profile Application

Files:
- Modify: `server/internal/game/rules.go`
- Add: `server/internal/game/dungeon_profiles.go`
- Modify: `server/internal/game/dungeon_gen.go`

- [x] Load `floor_profiles` into `DungeonGenerationRules`.
- [x] Apply the matching profile before ordinary dungeon generation starts.
- [x] Use profiled floor size when constructing navigation bounds.
- [x] Keep boss-floor rules unchanged.

```bash
cd server && go test ./internal/game -run 'TestDungeonFloorProfiles|TestDungeonMonsterGeneration|TestDungeonObstacleGeneration|TestBossFloorGeneration'
```

## Task 3 - Focused Proof

Files:
- Add: `server/internal/game/dungeon_profiles_test.go`
- Existing bot scenarios: `tools/bot/scenarios/12_dungeon_levels.json`,
  `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`

- [x] Add tests that compare entry and deeper profile outputs using rule-derived expectations.
- [x] Add tests that prove same seed/level stays deterministic and reachable.
- [x] Run dungeon protocol bot proofs.

```bash
make bot scenario=12_dungeon_levels
make bot scenario=28_reachable_dungeon_obstacles
```

## Task 4 - Lifecycle Docs

Files:
- Modify: `docs/specs/v252_spec-expanded-dungeon-profile.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v252_expanded-dungeon-profile.md`
- Modify: `PROGRESS.md`

- [x] Mark the spec complete.
- [x] Add v252 lifecycle and as-built notes.
- [x] Update `PROGRESS.md` current status and keep rivers/biome colors/final balance deferred.

```bash
make maintainability
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestDungeonFloorProfiles|TestDungeonMonsterGeneration|TestDungeonObstacleGeneration|TestBossFloorGeneration'`
- [x] `make bot scenario=12_dungeon_levels`
- [x] `make bot scenario=28_reachable_dungeon_obstacles`
- [x] `make maintainability`

Autoloop batch note: final `make ci` was attempted after the one selected slice. The Go-test
failure from profile-unaware reachability checks was fixed; the remaining red step is the protocol
bot catalog, with failures in `teleporter_lab`, `dungeon_monsters`, `boss_floor_gate`,
`account_stash_storage`, `pack_aggro_and_dungeon_packs`, and
`ranger_piercing_and_pinning_shots`.
