# v155 As-Built - Random Quest Reward Floors

Spec: [`docs/specs/v155_spec-random-quest-reward-floors.md`](../specs/v155_spec-random-quest-reward-floors.md)
Plan: [`docs/plans/v155_2026-06-14-random-quest-reward-floors.md`](../plans/v155_2026-06-14-random-quest-reward-floors.md)

## What shipped

- Added deterministic random quest reward floors to generated non-boss dungeon levels.
- Rolled quest reward floors on a separate server-owned RNG stream at roughly 10% of eligible
  floors.
- Placed one extra reachable treasure chest on rolled quest floors using the existing chest
  interactable and loot-table path.
- Excluded boss floors and town/static levels from quest reward rolls.
- Added focused Go coverage for deterministic distribution, boss exclusion, reachability, and chest
  activation.
- Added protocol bot proof:
  - `make bot scenario=65_random_quest_reward_floor.json`

## Key decisions

- Kept this as a first generated-quest foundation rather than a full quest log, NPC, or persistence
  system.
- Reused the existing treasure chest activation and loot-drop flow so the client needed no protocol
  or UI schema changes.
- Tagged quest reward chests only in generation internals so tests can distinguish them without
  expanding the snapshot contract.

## Deferred

- Town NPC quest offers, journal UI, and durable quest state.
- Objective types beyond the initial reward-floor chest.
- Quest-specific chest presentation or map markers.

## Verification

- `cd server && go test ./internal/game -run 'TestRandomQuest|TestGuardedChestGeneration|TestGeneratedDungeonSourcesUseDepthLootTables|TestDungeonStairsGolden|TestBossFloorGenerationGolden' -count=1`
- `make bot scenario=65_random_quest_reward_floor.json`
- `make maintainability`
- `make ci`
