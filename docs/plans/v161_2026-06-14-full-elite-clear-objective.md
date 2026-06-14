# v161 Plan — Full elite clear objective

Status: Complete
Goal: Elite objective chests unlock only after all generated elite pack leaders on the floor are dead.
Architecture: The server remains authoritative over objective completion. Runtime `LevelState`
already tracks objective chest IDs and generated monsters already carry `monsterPackLeader`; the
lock check changes from "any dead leader" to "no live generated leader remains." No protocol/schema
or client code change is required.
Tech stack: Go sim, Python protocol bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v160 `dungeon-population-extraction` and v159 `kill-gated-elite-objective`. Existing
plugins/assets: reject; this is backend-owned objective completion logic. Client visual work is
deferred to the selected v162 objective chest presentation slice.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/interactables.go` | Clear-all elite objective lock check |
| Modify | `server/internal/game/dungeon_elite_objective_test.go` | Multi-leader completion coverage |
| Modify | `tools/bot/run.py` | Narrow leader selector support for `kill_monsters` |
| Modify | `tools/bot/scenarios/68_dungeon_elite_side_objective.json` | Protocol proof wording/alignment |
| Create | `docs/as-built/v161_full-elite-clear-objective.md` | As-built summary |
| Modify | `docs/specs/v161_spec-full-elite-clear-objective.md` | Mark complete at closeout |
| Modify | `PROGRESS.md` | Lifecycle closeout and next slice |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/bot/run.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `tools/bot/run.py`
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] No extraction needed; `tools/bot/run.py` stayed within its ratchet allowance and focused Go files are under 600 lines.

Verification:

```bash
make maintainability
```

## Task 1 — Clear-all objective gate

Files:
- Modify: `server/internal/game/interactables.go`

- [x] Change `eliteObjectiveChestLocked` so any live generated pack leader keeps the objective chest locked.
- [x] Keep the `elite_objective_incomplete` rejection reason and existing chest-open path.
- [x] Keep non-objective chests unaffected.

```bash
cd server && go test ./internal/game -run TestEliteObjectiveChestRequiresLeaderKill -count=1
```

## Task 2 — Multi-leader server proof

Files:
- Modify: `server/internal/game/dungeon_elite_objective_test.go`

- [x] Add or update focused coverage with at least two live elite pack leaders on one level.
- [x] Prove killing only one leader still rejects.
- [x] Prove killing all leaders opens the objective chest and drops loot.

```bash
cd server && go test ./internal/game -run 'TestEliteObjectiveChestRequiresAllLeaderKills|TestEliteObjectiveChestRequiresLeaderKill' -count=1
```

## Task 3 — Bot scenario proof

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/scenarios/68_dungeon_elite_side_objective.json`

- [x] Update scenario wording to say the chest opens after the generated elite objective is cleared.
- [x] Add a narrow optional `monster_pack_leader` selector to `kill_monsters`.
- [x] Keep the live protocol proof green for reject-then-complete-then-open.

```bash
make bot scenario=68_dungeon_elite_side_objective.json
```

## Task 4 — Lifecycle docs and CI

Files:
- Create: `docs/as-built/v161_full-elite-clear-objective.md`
- Modify: `docs/plans/v161_2026-06-14-full-elite-clear-objective.md`
- Modify: `docs/specs/v161_spec-full-elite-clear-objective.md`
- Modify: `PROGRESS.md`

- [x] Mark completed plan tasks.
- [x] Update spec status/as-built/progress.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `cd server && go test ./internal/game -run 'TestEliteObjectiveChestRequiresAllLeaderKills|TestEliteObjectiveChestRequiresLeaderKill' -count=1`
- [x] `make bot scenario=68_dungeon_elite_side_objective.json`
- [x] `make ci`
