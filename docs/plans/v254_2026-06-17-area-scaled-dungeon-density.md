# v254 Plan - Area-Scaled Dungeon Density

Status: Complete
Goal: Derive ordinary dungeon obstacle and enemy density from each generated floor's area.
Architecture: Keep dungeon density server-authoritative and data-driven. Shared rules define
area-density formulas with min/max caps; the Go generator resolves those formulas after selecting
the effective floor size and before placing generated content. Boss floors continue to use their
compact boss-floor rules.
Tech stack: shared JSON/schema, Go simulation, Python shared validator, protocol bot, docs.

## Baseline and Shortcut Decision

Builds on v252 expanded dungeon profiles and v253 fog-of-war visibility. Existing generated wall
rendering and monster spawning stay in place; this slice changes only ordinary-floor generation
density inputs.

Asset/plugin decision: reject external assets/plugins. Borrow the existing in-repo generated wall
rendering and dungeon bot scenarios; no client visual code is in scope.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Replace fixed ordinary-floor density counts with area-density formulas and conservative increased caps |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate density formula structure |
| Modify | `tools/validate_shared.py` | Cross-check density formulas, monster minimums, and obstacle ranges |
| Modify | `server/internal/game/rules.go` | Load/validate formula fields and expose resolved base density |
| Modify | `server/internal/game/dungeon_profiles.go` | Resolve formulas from effective floor area and apply them to ordinary floors |
| Modify | `server/internal/game/dungeon_profiles_test.go` | Add rule-derived formula, determinism, reachability, and boss-floor exclusion proof |
| Modify | `docs/specs/v254_spec-area-scaled-dungeon-density.md` | Mark status complete when shipped |
| Modify | `docs/progress/slice-lifecycle.md` | Add v254 lifecycle row |
| Add | `docs/as-built/v254_area-scaled-dungeon-density.md` | Record shipped behavior and deferrals |
| Modify | `PROGRESS.md` | Update current status and remove the shipped population-count/final-density gap wording |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `game_test.go` baseline after moving the obstacle golden test |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/rules.go`
- [x] Other over-limit files from `.maintainability/file-size-baseline.tsv`: `tools/validate_shared.py` stayed within allowance after extracting `tools/dungeon_density.py`; `server/internal/game/game_test.go` shrank after moving the obstacle golden proof and its baseline was lowered.
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Extract focused helper/test work into `dungeon_profiles.go` and `dungeon_profiles_test.go`;
  keep `dungeon_gen.go` untouched unless a focused verification failure requires a hook fix.

Verification:
```bash
make maintainability
```

## Task 1 - Shared Density Rules

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Add area-density formula objects for monster population, pack-count range, and
  obstacle-group range.
- [x] Tune the first formulas so entry floors increase modestly and depth-4 profiled floors
  increase more because their area is larger.
- [x] Validate formula structure, positive area divisors, min/max caps, and monster minimums.

```bash
make validate-shared
```

## Task 2 - Server Formula Resolution

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/dungeon_profiles.go`

- [x] Add Go rule structs for area-density formulas.
- [x] Resolve formulas from `floor_size.width * floor_size.height` after applying the matching
  ordinary-floor profile.
- [x] Preserve derived base density on loaded rules for tests and helpers that inspect
  `rules.DungeonGeneration` directly.
- [x] Keep boss floors on `boss_floor.floor_size` and `boss_floor.monster_count`.

```bash
cd server && go test ./internal/game -run 'TestDungeonDensityFormulas|TestDungeonFloorProfiles|TestBossFloorGeneration'
```

## Task 3 - Focused Dungeon Proof

Files:
- Modify: `server/internal/game/dungeon_profiles_test.go`
- Existing bot scenarios: `tools/bot/scenarios/12_dungeon_levels.json`,
  `tools/bot/scenarios/14_dungeon_monsters.json`,
  `tools/bot/scenarios/28_reachable_dungeon_obstacles.json`

- [x] Add tests that compute expected population and ranges from shared rules rather than fixed
  copied tuning values.
- [x] Prove entry and depth-4 ordinary floors both use formula-derived densities.
- [x] Prove same seed/level stays deterministic and reachable after denser obstacles/monsters.
- [x] Run focused dungeon protocol bot scenarios.
- [x] Refresh the explicit obstacle golden and rerun density-sensitive CI bot proofs after the
  conservative formula retune.
- [x] Keep unrelated account-stash persistence proof out of generated dungeon traversal after the
  denser paths made `/state` replay exceed the bot HTTP timeout.

```bash
cd server && go test ./internal/game -run 'TestDungeonDensityFormulas|TestDungeonFloorProfiles|TestDungeonMonsterGeneration|TestDungeonObstacleGeneration|TestBossFloorGeneration'
make bot scenario=12_dungeon_levels
make bot scenario=14_dungeon_monsters
make bot scenario=28_reachable_dungeon_obstacles
make bot scenario=13_teleporter_lab
make bot scenario=17_treasure_classes_and_guarded_chests
make bot scenario=42_pack_aggro_and_dungeon_packs
make bot scenario=68_dungeon_elite_side_objective
make bot scenario=77_elite_minion_pack_ai
VERBOSE=1 make bot scenario=account_stash_storage
SCENARIO=44_elite_objective_hud HEADLESS=1 ./scripts/bot_client_local.sh
```

## Task 4 - Lifecycle Docs

Files:
- Modify: `docs/specs/v254_spec-area-scaled-dungeon-density.md`
- Modify: `docs/plans/v254_2026-06-17-area-scaled-dungeon-density.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v254_area-scaled-dungeon-density.md`
- Modify: `PROGRESS.md`

- [x] Mark the spec and plan complete.
- [x] Add v254 lifecycle and as-built notes.
- [x] Update `PROGRESS.md` current status, CI note, and deferred dungeon-generation wording.

```bash
make maintainability
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestDungeonDensityFormulas|TestDungeonFloorProfiles|TestDungeonMonsterGeneration|TestDungeonObstacleGeneration|TestBossFloorGeneration'`
- [x] `make bot scenario=12_dungeon_levels`
- [x] `make bot scenario=14_dungeon_monsters`
- [x] `make bot scenario=28_reachable_dungeon_obstacles`
- [x] `make bot scenario=13_teleporter_lab`
- [x] `make bot scenario=17_treasure_classes_and_guarded_chests`
- [x] `make bot scenario=42_pack_aggro_and_dungeon_packs`
- [x] `make bot scenario=68_dungeon_elite_side_objective`
- [x] `make bot scenario=77_elite_minion_pack_ai`
- [x] `VERBOSE=1 make bot scenario=account_stash_storage`
- [x] `SCENARIO=44_elite_objective_hud HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make maintainability`
- [x] `make ci`

Autoloop batch note: final `make ci` passed after the selected slice commit and CI-stabilization
follow-up.
