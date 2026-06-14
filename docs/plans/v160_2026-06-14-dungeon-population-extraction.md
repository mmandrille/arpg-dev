# v160 Plan — Dungeon population extraction

Status: Complete
Goal: Move generated dungeon runtime population out of `sim.go` without changing gameplay behavior.
Architecture: Dungeon generation remains data-driven and deterministic. `Sim.ensureDungeonLevel`
continues to create `LevelState`, but a focused population module converts generated level output
into runtime entities in the existing allocation order. No protocol, client, or shared rule changes
are required.
Tech stack: Go sim, Go tests, SDD docs.

## Baseline and shortcut decision

Builds on v159 `kill-gated-elite-objective`. Existing plugins/assets: reject; this is backend-only
simulation structure with no Godot client work.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `server/internal/game/dungeon_population.go` | Runtime conversion of generated dungeon output into level entities |
| Create | `server/internal/game/dungeon_population_test.go` | Focused coverage for objective chest ID tracking and generated monster runtime state |
| Modify | `server/internal/game/sim.go` | Delegate population and remove moved body |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `sim.go` baseline after verified shrink |
| Create | `docs/as-built/v160_dungeon-population-extraction.md` | As-built summary |
| Modify | `docs/specs/v160_spec-dungeon-population-extraction.md` | Mark complete at closeout |
| Modify | `PROGRESS.md` | Lifecycle closeout and next slice |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:

```bash
make maintainability
```

## Task 1 — Extract runtime population

Files:
- Create: `server/internal/game/dungeon_population.go`
- Modify: `server/internal/game/sim.go`

- [x] Move `populateDungeonLevel` and generated monster runtime stat helpers into the focused file.
- [x] Preserve entity allocation order: stairs, teleporters, chests, loose loot, monsters, corpses.
- [x] Preserve boss visual model/tint/scale selection and generated rarity scaling behavior.
- [x] Keep `ensureDungeonLevel` as the only caller.

```bash
cd server && go test ./internal/game -run 'TestGeneratedDungeonSourcesUseDepthLootTables|TestGeneratedDungeonMonsterRarityGolden' -count=1
```

## Task 2 — Focused extraction proof

Files:
- Create: `server/internal/game/dungeon_population_test.go`

- [x] Add coverage proving elite-objective chests are still tracked on runtime `LevelState`.
- [x] Add coverage proving generated boss and rarity runtime fields survive population.

```bash
cd server && go test ./internal/game -run 'TestPopulateDungeonLevelTracksEliteObjectiveChestIDs|TestPopulateDungeonLevelPreservesBossAndRarityRuntimeState' -count=1
```

## Task 3 — Regression proof for objective and generated dungeons

Files:
- Existing Go tests only

- [x] Run existing generated dungeon and objective tests that cover runtime behavior.

```bash
cd server && go test ./internal/game -run 'TestGeneratedDungeonSourcesUseDepthLootTables|TestGeneratedDungeonMonsterRarityGolden|TestDungeonEliteObjectiveChestRequiresEliteLeader|TestEliteObjectiveChestRequiresLeaderKill' -count=1
```

## Task 4 — Lifecycle docs and CI

Files:
- Create: `docs/as-built/v160_dungeon-population-extraction.md`
- Modify: `docs/plans/v160_2026-06-14-dungeon-population-extraction.md`
- Modify: `docs/specs/v160_spec-dungeon-population-extraction.md`
- Modify: `PROGRESS.md`
- Modify: `.maintainability/file-size-baseline.tsv`

- [x] Mark completed plan tasks.
- [x] Update spec status/as-built/progress.
- [x] Lower `sim.go` maintainability baseline after the shrink is verified.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `cd server && go test ./internal/game -run 'TestPopulateDungeonLevelTracksEliteObjectiveChestIDs|TestPopulateDungeonLevelPreservesBossAndRarityRuntimeState' -count=1`
- [x] `cd server && go test ./internal/game -run 'TestGeneratedDungeonSourcesUseDepthLootTables|TestGeneratedDungeonMonsterRarityGolden|TestDungeonEliteObjectiveChestRequiresEliteLeader|TestEliteObjectiveChestRequiresLeaderKill' -count=1`
- [x] `make ci`
