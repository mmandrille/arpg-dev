# v173 Spec — Quest Floor Map Marker

Status: Approved for planning
Date: 2026-06-14
Codename: quest-floor-map-marker

## Purpose

Random quest reward floors already place an extra reachable treasure chest, but the chest is not visible as a quest reward in protocol or client presentation. This slice exposes server-authored quest reward metadata on the chest entity and renders a lightweight client marker so players and bots can identify the reward chest after descending onto a quest floor.

## Non-goals

- No full quest journal, minimap, compass, or floor-map UI.
- No changes to the reward floor spawn chance, chest loot table, or floor generation balance.
- No quest-specific opening rules; reward chests remain normal treasure chests mechanically.

## Acceptance Criteria

- Generated random quest reward chests are marked with optional `quest_reward: true` in v8 `session_snapshot` and `state_delta` entity views.
- Non-quest chests omit `quest_reward`.
- The Godot client stores the quest reward flag and renders a distinct display-only marker on quest reward treasure chests.
- The marker remains visible after the chest opens, matching the existing objective chest behavior.
- Bot/debug entity output can filter on `quest_reward` and `has_quest_marker`.
- A pinned client bot scenario reaches the existing deterministic random quest reward floor and asserts the quest reward marker is present.

## Scope and Files Likely Touched

- Contracts: `shared/protocol/session_snapshot.v8.schema.json`, `shared/protocol/state_delta.v8.schema.json`.
- Server: `server/internal/game/types.go`, `server/internal/game/level.go`, `server/internal/game/dungeon_population.go`, `server/internal/game/sim.go`, random quest tests.
- Client: `client/scripts/chest_presentation.gd`, `client/scripts/main.gd`, `client/tests/test_item_visuals.gd`.
- Bot tooling: `client/scripts/bot_scenario_runner.gd`, new client bot scenario under `tools/bot/scenarios/client/`.
- Docs: this spec, matching plan, as-built notes, and `PROGRESS.md`.

## Test and Bot Proof

- `make validate-shared` validates the new optional protocol field.
- Focused Go tests prove quest reward generated chests become runtime entity views with `QuestReward`.
- `make client-unit` covers marker creation and opened-state persistence.
- Existing protocol bot `make bot scenario=65_random_quest_reward_floor.json` remains green.
- New client bot scenario `make bot-client scenario=42_quest_reward_chest_presentation.json` asserts `quest_reward` and `has_quest_marker` on the pinned floor.
- Final `make ci` passes before commit.

## Open Questions and Risks

- No blocking questions.
- Risk: `client/scripts/main.gd` is grandfathered over the line limit, so marker-specific behavior should stay in `chest_presentation.gd` and avoid growing `main.gd` beyond the maintenance ratchet allowance.
