# v174 Spec — Quest Journal Foundation

Status: Approved for planning
Date: 2026-06-14
Codename: quest-journal-foundation

## Purpose

Add the first player-facing quest journal surface. The journal should expose the current generated quest-floor reward objective using the `quest_reward` chest metadata from v173, so a player can press `J` and understand whether the current floor has an active reward chest and whether it has been opened.

## Non-goals

- No durable quest state, NPC quest offers, turn-ins, or account/character persistence.
- No new server quest model or protocol schema beyond existing `quest_reward` entity metadata.
- No minimap pins, compass, full floor map, or multi-quest tracking.
- No changes to random quest reward floor chance, chest loot, or opening mechanics.

## Acceptance Criteria

- The Godot client has a quest journal panel toggled by `J` during gameplay.
- On a floor with a closed `quest_reward` treasure chest, the journal shows one active objective to find/open the reward chest.
- After that chest opens, the journal marks the objective complete.
- On floors without a quest reward chest, the journal shows an empty/no-active-quest state.
- Bot/debug state exposes journal visibility and the current journal objective list.
- A pinned client bot scenario descends to `v155_bot_quest_0015`, opens the journal, and asserts the active reward objective is visible.

## Scope and Files Likely Touched

- Client UI: new `client/scripts/quest_journal_panel.gd`.
- Client wiring: `client/scripts/main.gd` for panel creation, `J` key handling, objective-state sync, and bot debug output.
- Client tests: new `client/tests/test_quest_journal_panel.gd`, plus `scripts/client_smoke.sh` gate.
- Bot tooling: `client/scripts/bot_scenario_runner.gd` for journal assertions and a new scenario under `tools/bot/scenarios/client/`.
- Docs: this spec, matching plan, as-built notes, and `PROGRESS.md`.

## Test and Bot Proof

- `make client-unit` covers the standalone journal panel states.
- `make bot-client scenario=43_quest_journal_foundation.json` proves the pinned reward floor journal objective is visible in the client.
- `make maintainability` proves touched grandfathered files stay within the ratchet.
- Final `make ci` passes before commit.

## Open Questions and Risks

- No blocking questions.
- Risk: `client/scripts/main.gd` and `client/scripts/bot_scenario_runner.gd` are grandfathered files; the implementation must keep journal logic in focused files and offset small wiring additions with line-count reductions.
