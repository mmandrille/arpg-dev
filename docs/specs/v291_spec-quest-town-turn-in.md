# v291 Spec: Quest Town Turn-In

Status: Implemented
Date: 2026-06-19
Codename: `quest-town-turn-in`
Baseline: v290 `blacksmith-armor-recipe`

## Purpose

Close the first small quest loop by letting players return a quest item to a town quest giver for a
server-authoritative reward. The slice should reuse the existing `quest_leaf` item and town
interactable/service model while staying short of a full quest-log, dialog, or durable quest system.

## Non-goals

- Do not add durable quest state, quest offers, branching dialog, NPC relationship state, objective
  persistence, or multi-step quest chains.
- Do not change random quest reward floor generation, reward chest logic, elite objectives, or the
  quest journal's current-floor objective derivation.
- Do not add a new client panel, production NPC art, portraits, VO/audio, external assets, or plugins.
- Do not add XP rewards, item rewards, repeat limits, account persistence, or anti-farming rules yet.

## Acceptance Criteria

- Shared rules define a ready `town_quest_giver` interactable with a quest-turn-in service.
- The default town and vendor-lab worlds include the quest giver in reachable positions.
- Main gameplay config owns the required turn-in item and gold reward.
- Clicking the town quest giver consumes one configured quest item from inventory when present.
- A successful turn-in emits a `quest_turn_in_completed` event, removes the consumed inventory item,
  increases character gold, and updates character progression.
- Clicking the quest giver without the required item rejects with a stable `missing_quest_item`
  reason.
- Focused server tests prove success, missing-item rejection, wrong-target rejection, and config-owned
  reward behavior.
- Protocol and client bot scenarios prove a player can pick up `Quest Leaf`, return to the town
  quest giver, and receive the gold reward.

## Scope And Likely Files

- Shared: `shared/rules/interactables.v0.json`, `shared/rules/worlds.v0.json`,
  `shared/rules/main_config.v0.json`, `shared/rules/main_config.v0.schema.json`.
- Server: `server/internal/game/quest_turn_in.go`, `server/internal/game/quest_turn_in_test.go`,
  `server/internal/game/rules.go`, `server/internal/game/main_config_validation.go`.
- Client presentation: `client/scripts/town_node_factory.gd`,
  `client/tests/test_quest_giver_visual.gd`.
- Bot: `tools/bot/scenarios/97_quest_town_turn_in.json`,
  `tools/bot/scenarios/client/75_quest_town_turn_in.json`.
- Docs: v291 plan/as-built/lifecycle updates.

## Test And Bot Proof

Focused checks:

```bash
make validate-shared
(cd server && go test ./internal/game -run 'QuestTurnIn|MainConfig|Interactable' -count=1)
godot --headless --path client --script res://tests/test_item_visuals.gd
godot --headless --path client --script res://tests/test_quest_giver_visual.gd
make bot scenario=97_quest_town_turn_in
make bot-client scenario=75_quest_town_turn_in HEADLESS=1
make maintainability
```

Manual visual verification command:

```bash
make bot-visual scenario=75_quest_town_turn_in
```

## Asset And Plugin Decision

- Adopt: existing static town service placement, interactable click handling, quest item inventory,
  gold/progression changes, and bot action infrastructure.
- Borrow: vendor/bishop/mercenary service tests and bot step patterns.
- Reject: external assets/plugins, production NPC art, dialog systems, new UI panels, new audio, or
  new asset pipelines.

## Open Questions And Risks

- No blocking questions. The first turn-in is intentionally repeatable while the player has matching
  quest items; durable quest completion and anti-farming rules are future quest-system work.

## Outcome

- Implemented in v291. The quest giver was appended after existing scenario-critical town entities
  to preserve old hardcoded unique-chest entity IDs while still making the new service reachable.
