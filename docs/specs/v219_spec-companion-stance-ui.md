# v219 Spec - Companion Stance UI

Codename: `companion-stance-ui`
Status: Complete
Baseline: v218 `rogue-duelist-mark`

## Purpose

Expose the existing server-authoritative companion stance command in the Godot UI so players can
switch owned companions and hired mercenaries between `assist`, `defend`, and `passive` without
using protocol-bot-only commands.

This advances the companion and mercenary loop by making v208 stance behavior player-facing while
keeping the server as the only authority for stance state and AI behavior.

## Non-goals

- No new protocol schema or server stance behavior.
- No per-companion stance targeting; the existing command applies to all living owned companions.
- No hold-position, retreat, durable stance persistence, or companion death/loss changes.
- No new external art, plugins, or production companion UI assets.

## Adopt / Borrow / Reject

- **Adopt:** Existing `companion_command_intent`, `companion_stance_changed` events, and companion
  `companion_stance` entity field from v208.
- **Borrow:** Existing mercenary roster panel and client-bot mercenary panel assertions from v207.
- **Reject:** New protocol, new art/plugin dependencies, and a separate companion-management window.

## Acceptance Criteria

1. The Mercenaries panel shows `assist`, `defend`, and `passive` stance controls when opened.
2. Clicking a stance sends `companion_command_intent` with that stance.
3. The panel reflects the authoritative companion stance from entity updates.
4. The focused mercenary panel unit test covers stance button state and debug payloads.
5. The existing mercenary roster client scenario clicks a stance control and proves the panel and
   companion state update.
6. Clicking a top-left companion HUD block opens the companion management panel/interface.

## Scope and Files Likely Touched

- `client/scripts/mercenary_panel.gd`
- `client/scripts/companion_bar.gd`
- `client/scripts/mercenary_panel_bridge.gd`
- `client/scripts/main.gd`
- `client/scripts/bot_mercenary_panel_assertions.gd`
- `client/scripts/bot_action_handlers.gd`
- `client/scripts/bot_step_catalog.gd`
- `client/scripts/bot_scenario_runner.gd`
- `client/tests/test_mercenary_panel.gd`
- `tools/bot/scenarios/client/47_mercenary_roster_ui.json`
- `docs/as-built/v219_companion-stance-ui.md`

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- `make client-unit`
- `make bot-client scenario=mercenary_roster_ui`

Manual visual check, if desired:

```bash
make bot-visual scenario=mercenary_roster_ui
```

## Open Questions and Risks

- None. The existing all-owned-companions command semantics are accepted for this slice.
