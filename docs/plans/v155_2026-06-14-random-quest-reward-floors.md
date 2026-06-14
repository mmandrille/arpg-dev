# v155 Plan - Random Quest Reward Floors

Status: Ready for implementation
Goal: Add deterministic 10% generated quest reward floors with openable reward chests.
Architecture: Keep the first quest slice server-owned and deterministic. Reuse generated dungeon
placement, reachability validation, chest opening, and loot drops instead of adding a new quest
protocol surface.
Tech stack: Go dungeon generation, protocol bot scenario, lifecycle docs.

## Baseline and shortcut decision

Builds on v154. Godot plugin decision: reject/none; this slice reuses existing chest and loot
presentation and does not add client UI.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/game/dungeon_gen.go` | Quest floor roll and reward chest placement |
| Add | `server/internal/game/random_quest_floors_test.go` | Determinism/distribution/opening tests |
| Add | `tools/bot/scenarios/65_random_quest_reward_floor.json` | Protocol proof |
| Add | `docs/as-built/v155_random-quest-reward-floors.md` | Shipped proof |
| Modify | `PROGRESS.md` | Lifecycle and deferred scope |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 - Dungeon generation

Files:
- Modify: `server/internal/game/dungeon_gen.go`
- Add: `server/internal/game/random_quest_floors_test.go`

- [ ] Add deterministic 10% quest reward floor roll for non-boss generated floors.
- [ ] Place one extra reachable reward chest on eligible floors.
- [ ] Prove determinism, distribution, boss exclusion, and chest open/loot behavior.

```bash
cd server && go test ./internal/game -run 'TestRandomQuest'
```

## Task 2 - Bot proof

Files:
- Add: `tools/bot/scenarios/65_random_quest_reward_floor.json`

- [ ] Add protocol bot proof for descending to a seeded quest floor and opening a chest.

```bash
make bot scenario=65_random_quest_reward_floor.json
```

## Task 3 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v155_random-quest-reward-floors.md`
- Modify: `PROGRESS.md`
- Modify: this plan

- [ ] Record shipped proof and update lifecycle docs.
- [ ] Run final verification.

```bash
make maintainability
make ci
```

## Final verification

- [ ] `cd server && go test ./internal/game -run 'TestRandomQuest'`
- [ ] `make bot scenario=65_random_quest_reward_floor.json`
- [ ] `make maintainability`
- [ ] `make ci`

## Deferred scope

- Town NPC quest offers.
- Quest log UI.
- Durable quest progress.
- Kill/collect/reach objective catalogs.

