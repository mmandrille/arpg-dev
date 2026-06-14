# v173 As-built — Quest Floor Map Marker

Date: 2026-06-14

## What shipped

- Random quest reward chests now carry server-authored `quest_reward` metadata through v8 entity views.
- `LevelState` tracks generated quest reward chest ids separately from elite objective chest ids.
- The Godot client stores `quest_reward`, renders a distinct `QuestRewardMarker`, and keeps it visible after the chest opens.
- Client bot/debug entity records now expose `quest_reward` and `has_quest_marker`.
- Added client bot scenario `42_quest_reward_chest_presentation.json` for the pinned `v155_bot_quest_0015` reward floor.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRandomQuest' -count=1`
- `make client-unit`
- `make bot scenario=65_random_quest_reward_floor.json`
- `make bot-client scenario=42_quest_reward_chest_presentation.json`
- `make maintainability`

## Scope limits

- This is presentation and protocol metadata only; quest reward chest spawn chance, loot, and opening rules are unchanged.
- No minimap, quest journal, or generalized floor-map UI shipped in this slice.
