# v240 Plan - Boss Portrait Panel

Status: Complete
Goal: Add a boss-template keyed portrait tile to the existing boss health bar.
Architecture: Render the portrait client-side from existing boss bar state; keep server/protocol and
boss mechanics unchanged.
Tech stack: Godot UI/client bot, docs.

## Baseline and Asset Decision

Builds on v53 boss health bar UI and v57 boss phase readability. Asset/plugin decision: reject
external assets/plugins; the portrait is code-drawn in `BossHealthBar` and keyed by
`boss_template_id`.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/boss_health_bar.gd` | Render portrait tile and expose debug state |
| Modify | `client/scripts/bot_scenario_runner.gd` | Let boss bar assertions check portrait kind |
| Modify | `client/scripts/bot_step_catalog.gd` | Validate portrait assertion fields |
| Modify | `client/tests/test_boss_health_bar.gd` | Prove portrait live/hide behavior |
| Add | `tools/bot/scenarios/client/57_boss_portrait_panel.json` | Client proof |
| Add | `docs/as-built/v240_boss-portrait-panel.md` | Slice proof |

## Maintenance Ratchet

Target: touched source/test/tool files stay at or below their allowed baselines.

Hotspot / over-limit files touched:
- [x] `client/scripts/bot_scenario_runner.gd`

Decision:
- [x] Add only narrow portrait fields to existing boss bar assertions.
- [x] Keep drawing local to `BossHealthBar`; no imported art or asset manifest changes.

Verification:
```bash
make maintainability
```

## Task 1 - Portrait UI/debug state

Files:
- Modify: `client/scripts/boss_health_bar.gd`

- [x] Add a portrait tile to the health bar layout.
- [x] Draw a distinct Cave Warden portrait from `boss_template_id`.
- [x] Expose `portrait_visible`, `portrait_kind`, and `portrait_label` in debug state.
- [x] Preserve existing HP/phase behavior.

## Task 2 - Tests and bot proof

Files:
- Modify: `client/tests/test_boss_health_bar.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Add: `tools/bot/scenarios/client/57_boss_portrait_panel.json`

- [x] Assert live Cave Warden portrait fields in the focused unit.
- [x] Assert portrait clears when the boss hides/dies.
- [x] Extend boss health bar assertions with portrait fields.
- [x] Add a scenario that reaches the boss floor and asserts the Cave Warden portrait kind.

```bash
godot --headless --path client --script res://tests/test_boss_health_bar.gd
make bot-client scenario=57_boss_portrait_panel.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/scenario-catalog.md`
- Modify: `docs/progress/slice-codename-index.md`
- Add: `docs/as-built/v240_boss-portrait-panel.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_boss_health_bar.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `make bot-client scenario=57_boss_portrait_panel.json HEADLESS=1`
- [x] `make maintainability`
- [x] `make ci`
