# v136 Spec: Unique Chest Client Proof

Status: Complete
Date: 2026-06-13
Codename: `unique-chest-client-proof`

## Purpose

Add a Godot client bot proof for the purple town unique chest so the client-visible panel shows the
hand-authored named uniques and their readable unique-effect summaries. This closes the UI proof gap
after the named unique catalog and tooltip improvements.

## Non-goals

- No new unique items, unique effects, balance changes, or chest server behavior changes.
- No new art or tooltip layout redesign.
- No changes to natural unique drop generation.
- No broad client bot runner rewrite.

## Acceptance Criteria

- A client bot scenario opens `town_unique_chest` in the Godot client.
- The scenario asserts the stash panel is visible in unique chest mode with unique chest rows.
- The scenario proves `Embercall Blade` and `Stormstring Bow` each appear exactly once.
- The scenario proves their readable effect summaries are available through the client row/debug
  state used by bot assertions.
- `make client-unit` and `make ci` pass.

## Likely Files

- `tools/bot/scenarios/client/40_unique_chest_client_proof.json`
- `client/scripts/stash_panel.gd`
- `client/scripts/bot_scenario_runner.gd`
- `PROGRESS.md`
- `docs/as-built/v136_unique-chest-client-proof.md`

## Test And Bot Proof

- `make client-unit`
- `SCENARIO=unique_chest_client_proof HEADLESS=1 make bot-client`
- `make maintainability`
- `make ci`

Visual verification command: `make bot-visual scenario=61_purple_town_unique_chest.json`.

## Open Questions And Risks

- Bot assertions should remain generic enough for stash rows without hardcoding this unique chest in
  the runner.
