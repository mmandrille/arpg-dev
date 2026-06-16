# v219 Plan - Companion Stance UI

Status: Ready for implementation
Goal: expose existing companion stance commands through the mercenary panel.
Architecture: keep stance authority on the server and make the client a command surface plus
renderer of authoritative companion state. Reuse the existing mercenary panel instead of adding a
new companion-management window.
Tech stack: Godot client, Godot client-bot scenario, lifecycle docs.

## Baseline and Shortcut Decision

Builds on v208 companion stance command and v207 mercenary roster UI. Asset/plugin decision:
borrow existing mercenary panel/button styling and client-bot assertions; reject external assets,
plugins, or new production UI art.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/mercenary_panel.gd` | Add stance controls, debug state, and signal emission. |
| Modify | `client/scripts/companion_bar.gd` | Emit a companion selection event from top-left HUD slots. |
| Modify | `client/scripts/main.gd` | Send `companion_command_intent` and sync companion stance into the panel. |
| Modify | `client/scripts/bot_mercenary_panel_assertions.gd` | Let client-bot assertions inspect selected stance. |
| Modify | `client/scripts/bot_action_handlers.gd` | Add click action for stance buttons. |
| Modify | `client/scripts/bot_step_catalog.gd` | Register the new client-bot step. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Log the new step succinctly. |
| Modify | `client/tests/test_mercenary_panel.gd` | Cover stance sync and signal emission. |
| Modify | `tools/bot/scenarios/client/47_mercenary_roster_ui.json` | Click passive stance and assert UI state. |
| Add | `docs/as-built/v219_companion-stance-ui.md` | Record proof and scope limits. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: check before finish.
- [ ] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Defer extraction with rationale: this slice adds a narrow signal handler to existing UI wiring;
  broader `main.gd` extraction is unrelated to the player-facing stance control.

Verification:

```bash
make maintainability
```

## Task 1 - Client Stance Controls

Files:
- Modify: `client/scripts/mercenary_panel.gd`
- Modify: `client/scripts/main.gd`

- [x] Add stance button row and signal.
- [x] Sync selected stance from companion entity state.
- [x] Send `companion_command_intent` from the main scene.
- [x] Open the companion panel from top-left companion HUD slot selection.

```bash
godot --headless --path client --script res://tests/test_mercenary_panel.gd
```

## Task 2 - Client Bot Proof

Files:
- Modify: `client/scripts/bot_mercenary_panel_assertions.gd`
- Modify: `client/scripts/bot_action_handlers.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `tools/bot/scenarios/client/47_mercenary_roster_ui.json`

- [x] Add a stance-click bot step.
- [x] Assert the selected panel stance after the authoritative entity update.

```bash
make bot-client scenario=mercenary_roster_ui
```

## Task 3 - Lifecycle Docs

Files:
- Add: `docs/as-built/v219_companion-stance-ui.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`

- [x] Record focused proof and scope limits.
- [x] Update current status and lifecycle rows.

```bash
make maintainability
```

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- [x] `make client-unit`
- [x] `make bot-client scenario=mercenary_roster_ui`
- [x] `make maintainability`

Final batch `make ci` is deferred to the enclosing `$autoloop` gate.
