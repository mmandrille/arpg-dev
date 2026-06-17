# v249 Plan - Boss Reward Panel

Status: Complete
Goal: Show a compact client HUD reward panel after a boss kill.
Architecture: Extend the existing boss HUD component with reward-panel state, invoke it from the
existing `boss_killed` client event path, and extend bot assertions to verify that state.
Tech stack: Godot UI/client bot, docs.

## Baseline and Asset Decision

Builds on v67 boss kill status and v240 boss portrait panel. No server changes are required because
the existing `boss_killed` event already includes the boss template id.

Asset/plugin decision:
- Adopt the existing code-drawn boss HUD component.
- Borrow Cave Warden portrait/title identity from `BossHealthBar`.
- Reject external assets/plugins and item-preview art.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/boss_health_bar.gd` | Add reward panel UI, show/clear methods, and debug state |
| Modify | `client/scripts/boss_visuals_controller.gd` | Forward reward panel display/clear calls |
| Modify | `client/scripts/main.gd` | Show reward panel from `boss_killed` without growing the file |
| Modify | `client/scripts/bot_debug_progression_setup.gd` | Seed full debug progression for client bot proofs |
| Modify | `client/scripts/bot_scenario_runner.gd` | Match reward panel debug assertions |
| Modify | `client/scripts/bot_step_catalog.gd` | Validate reward panel expectations |
| Modify | `client/tests/test_boss_health_bar.gd` | Prove reward panel state |
| Modify | `scripts/bot_client.sh` | Pass scenario debug progression JSON to Godot |
| Add | `tools/bot/scenarios/client/65_boss_reward_panel.json` | Client bot proof |
| Add | `docs/as-built/v249_boss-reward-panel.md` | As-built proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines, except grandfathered baselines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`

Decision:
- [x] Keep `main.gd` line-neutral.
- [x] Keep reward UI in `boss_health_bar.gd`, which is below the limit.

Verification:
```bash
make maintainability
```

## Task 1 - Boss HUD reward panel

Files:
- Modify: `client/scripts/boss_health_bar.gd`
- Modify: `client/scripts/boss_visuals_controller.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_boss_health_bar.gd`

- [x] Add reward panel UI and debug fields.
- [x] Hide reward panel when live boss state returns or level/session clears.
- [x] Invoke reward panel display from the existing `boss_killed` event branch.
- [x] Prove show/hide state in a focused Godot unit test.

```bash
godot --headless --path client --script res://tests/test_boss_health_bar.gd
```

## Task 2 - Client bot proof

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_debug_progression_setup.gd`
- Modify: `scripts/bot_client.sh`
- Add: `tools/bot/scenarios/client/65_boss_reward_panel.json`

- [x] Extend boss health bar assertions with reward panel fields.
- [x] Let client bot scenarios seed full debug progression for deterministic boss proofs.
- [x] Add a client bot scenario that kills Cave Warden and asserts the reward panel.
- [x] Prove bot validation and scenario execution.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
make bot-client scenario=65_boss_reward_panel.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v249_boss-reward-panel.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_boss_health_bar.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `make bot-client scenario=65_boss_reward_panel.json HEADLESS=1`
- [x] `make maintainability`
