# v158 As-built — Dungeon elite side objective

Date: 2026-06-14

## Shipped

- Added data-backed `elite_objective` dungeon generation rules and schema validation.
- The generator now uses a dedicated elite-objective RNG stream per depth.
- Generated non-boss dungeon floors place one extra reachable treasure chest only when the floor has an elite pack leader.
- The objective chest uses the configured interactable definition and loot table, independent of regular guarded chest placement.
- Protocol bot entity selectors can assert `monster_pack_leader` so scenarios can prove elite-floor runtime state.
- Added `68_dungeon_elite_side_objective` as a deterministic side-objective reward proof.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestDungeonEliteObjectiveChestRequiresEliteLeader|TestDungeonMonsterGenerationCanForceElitePackLeaders' -count=1`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=68_dungeon_elite_side_objective.json`

Visual/client verification command:

```bash
make bot-visual scenario=68_dungeon_elite_side_objective.json
```

## Notes

- This slice intentionally keeps the objective as a small generated-floor reward hook rather than a full quest system.
- Kill-gated completion, quest journal text, objective markers, and custom chest visuals remain deferred.
