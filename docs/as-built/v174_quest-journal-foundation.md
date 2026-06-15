# v174 As-built — Quest Journal Foundation

Date: 2026-06-14

## What shipped

- Added a `J`-toggle quest journal panel using the existing draggable-window UI pattern.
- The journal derives current-floor objectives from server-authored `quest_reward` chest entity metadata.
- Reward-floor objectives show as active while the marked reward chest is closed and complete once it opens.
- Bot debug state now exposes `quest_journal_panel`, and client bot steps can wait/assert journal state.
- Added client bot scenario `43_quest_journal_foundation.json` for the pinned `v155_bot_quest_0015` reward floor.
- Lowered the `client/scripts/main.gd` file-size baseline after removing redundant blank spacing.

## Proof

- `make maintainability`
- `make client-unit`
- `make bot-client scenario=43_quest_journal_foundation.json`

## Scope limits

- No server quest model, durable quest persistence, NPC offers, turn-ins, minimap, or floor-map UI shipped.
- The journal is currently a current-floor display over existing entity metadata.
