# v159 Plan — Kill-gated elite objective

Status: Complete
Goal: Elite objective chests require killing a generated pack leader before opening.
Architecture: The server remains authoritative over objective completion. Generated dungeon metadata marks the objective chest, runtime activation rejects until at least one generated pack leader on the active level has been killed, and the existing chest loot path is reused after completion. No protocol/schema change is required because the client already receives accepted/rejected envelopes and chest state deltas.
Tech stack: Go sim, Python protocol bot, SDD docs.

## Baseline and shortcut decision

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `server/internal/game/interactables.go` | Focused interactable activation and elite-objective gate |
| Modify | `server/internal/game/sim.go` | Remove moved activation body; preserve objective metadata when spawning chests |
| Modify | `server/internal/game/level.go` | Track runtime elite-objective chest IDs per level |
| Modify | `server/internal/game/dungeon_elite_objective_test.go` | Focused reject-then-kill-then-open coverage |
| Modify | `tools/bot/scenarios/68_dungeon_elite_side_objective.json` | Protocol proof of gated objective |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `sim.go` baseline after extraction |
| Create | `docs/as-built/v159_kill-gated-elite-objective.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle closeout and deferred queue |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] `server/internal/game/handlers.go`
- [x] `server/internal/game/game_test.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:

```bash
make maintainability
```

## Task 1 — Interactable activation extraction and gate

Files:
- Create: `server/internal/game/interactables.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/level.go`

- [x] Move `activateInteractable` out of `sim.go` into `interactables.go`.
- [x] Track generated elite-objective chest IDs on the owning `LevelState`.
- [x] Reject objective chest activation until any `monsterPackLeader` on that level has been killed.

```bash
cd server && go test ./internal/game -run TestTreasureChestOpensOnceAndDropsLoot -count=1
```

## Task 2 — Focused server proof

Files:
- Modify: `server/internal/game/dungeon_elite_objective_test.go`

- [x] Add a generated-floor test that opening the objective chest rejects before an elite pack leader dies.
- [x] Mark one leader killed and prove the same chest opens and drops loot.

```bash
cd server && go test ./internal/game -run 'TestDungeonEliteObjectiveChestRequiresEliteLeader|TestEliteObjectiveChestRequiresLeaderKill' -count=1
```

## Task 3 — Bot scenario proof

Files:
- Modify: `tools/bot/scenarios/68_dungeon_elite_side_objective.json`

- [x] Update the scenario to attempt the objective chest first and expect rejection.
- [x] Kill a generated elite pack leader, then open the chest and assert loot/drop events.

```bash
make bot scenario=68_dungeon_elite_side_objective.json
```

## Task 4 — Lifecycle docs and CI

Files:
- Create: `docs/as-built/v159_kill-gated-elite-objective.md`
- Modify: `docs/plans/v159_2026-06-14-kill-gated-elite-objective.md`
- Modify: `docs/specs/v159_spec-kill-gated-elite-objective.md`
- Modify: `PROGRESS.md`
- Modify: `.maintainability/file-size-baseline.tsv`

- [x] Mark completed plan tasks.
- [x] Update spec status/as-built/progress.
- [x] Lower maintainability baseline for `sim.go`.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `cd server && go test ./internal/game -run 'TestDungeonEliteObjectiveChestRequiresEliteLeader|TestEliteObjectiveChestRequiresLeaderKill' -count=1`
- [x] `make bot scenario=68_dungeon_elite_side_objective.json`
- [x] `make ci`
