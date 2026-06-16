# v207 Spec: Mercenary Roster UI

Status: Complete
Date: 2026-06-15
Codename: mercenary-roster-ui

## Purpose

Give the v206 mercenary hiring board a player-visible Godot roster panel. When the player interacts with the town mercenary board, the client should open a compact panel that shows the fixed guard hire offer, the player's current gold/affordability, and the currently hired mercenary after the authoritative hire event arrives.

This slice is display-first. The server still owns the hire action and replacement behavior; the panel reflects the existing `mercenary_board_opened` and `mercenary_hired` events plus current companion state.

## Baseline

Builds on v206 mercenary hiring board:

- The server emits `mercenary_board_opened` before attempting the hire and emits `mercenary_hired` after a successful spend/spawn.
- The client already renders owned companions through `CompanionBar`, including a mercenary icon kind for `mercenary_guard`.
- Existing town services use compact draggable panels and client bot debug-state assertions.

Asset/plugin decision: borrow existing in-repo draggable service-panel and `CompanionBar` presentation patterns. Reject external assets/plugins and new art pipelines; this is a Godot control/UI slice using code-native styling.

## Non-goals

- No server/protocol schema changes.
- No second confirmation button or new client-authored hire command beyond the existing board interaction.
- No multiple offer browsing, player-character mercenary listings, player-set pricing, durable roster records, death recovery, insurance, or snapshot refresh rules.
- No companion stance/command UI; that remains the next selected slice.
- No production mercenary portraits or imported UI art.

## Acceptance Criteria

- The Godot client opens a `Mercenaries` panel when it receives `mercenary_board_opened`.
- The panel shows the fixed `mercenary_guard` offer with offer id, service id, price, player gold, and affordability.
- When `mercenary_hired` arrives, the panel updates to a successful status and shows the hired companion's monster definition/entity id when available.
- The panel refreshes affordability when the player's gold changes.
- The existing companion bar remains synchronized and shows the hired mercenary as a companion.
- A focused Godot unit test covers panel rendering/debug state and hired-roster updates without a live server.
- A headless client bot scenario clicks the mercenary board and asserts the panel, hire event, companion bar, and hired roster state.

## Scope and Likely Files

- Client UI: add `client/scripts/mercenary_panel.gd`.
- Client wiring: update `client/scripts/main.gd` with a preload, panel instance, event handlers, gold refresh, and bot debug state; keep heavy glue in focused helper scripts.
- Client bot: update `client/scripts/bot_step_catalog.gd`, assertion/wait handlers, and a focused mercenary assertion helper for `wait_mercenary_panel` / `assert_mercenary_panel`.
- Tests: add `client/tests/test_mercenary_panel.gd`.
- Scenario: add `tools/bot/scenarios/client/47_mercenary_roster_ui.json`.
- Docs: update plan/as-built/progress/lifecycle/scenario catalog.

## Test and Bot Proof

- `make client-unit`
- `make bot-client scenario=mercenary_roster_ui`
- `make bot scenario=mercenary_hiring_board`
- `make maintainability`
- Final `make ci`

## Open Questions and Risks

- No blocking questions.
- Risk: v206's board action immediately hires, so the panel can show an offer and resulting roster but cannot provide a separate confirmation flow without a protocol/server change. This is intentional for v207; richer roster/hire UX remains deferred.
- Risk: client bot assertions add small hooks to already-large bot files. Keep matching logic compact and panel behavior in the new focused script.
