# v159 As-built — Kill-gated elite objective

Date: 2026-06-14

## Shipped

- Elite-objective generated chests now retain runtime objective identity on their `LevelState`.
- Objective chest activation rejects with `elite_objective_incomplete` until at least one generated pack leader on that level has been killed.
- Once the objective is complete, the same chest reuses the existing treasure chest activation and loot-drop path.
- Moved generic interactable activation from `sim.go` into `server/internal/game/interactables.go`, lowering the `sim.go` ratchet baseline.
- Trimmed unrelated pre-existing `inventory_panel.gd` whitespace drift so the file-size ratchet is green without changing behavior.
- Updated protocol bot scenario `68_dungeon_elite_side_objective` to prove reject-then-kill-then-open through the normal WebSocket path.

## Verification

- `make maintainability`
- `cd server && go test ./internal/game -run 'TestDungeonEliteObjectiveChestRequiresEliteLeader|TestEliteObjectiveChestRequiresLeaderKill|TestTreasureChestOpensOnceAndDropsLoot' -count=1`
- `make bot scenario=68_dungeon_elite_side_objective.json`
- `make ci`

Visual/client verification command:

```bash
make bot-visual scenario=68_dungeon_elite_side_objective.json
```

## Notes

- v159 intentionally keeps the objective as a small “kill one elite leader, claim side reward” proof.
- Full quest log text, clearing every elite leader, named objective state, objective markers, and special chest presentation remain deferred.
