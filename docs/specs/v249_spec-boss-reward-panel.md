# v249 Spec - Boss Reward Panel

Status: Complete
Date: 2026-06-17
Codename: boss-reward-panel

## Purpose

Make boss victories feel less abrupt by replacing the disappearing boss bar with a compact reward
panel after a `boss_killed` event. The panel should reuse the existing boss kill status and tell the
player that the exit and boss reward chest are now available.

## Non-goals

- No server/protocol changes, loot table changes, XP tuning, boss balance, chest animation, audio,
  multi-boss reward summary, item preview, or external art/assets.
- No new boss reward economy. This is a client presentation layer for existing boss kill and reward
  chest behavior.

## Client Asset / Plugin Decision

- **Adopt:** Existing code-drawn `BossHealthBar` HUD surface and `boss_killed` client status.
- **Borrow:** Existing Cave Warden portrait identity for post-kill reward panel context.
- **Reject:** External assets/plugins, imported victory art, and new reward item icons.

## Acceptance Criteria

- When the client observes `boss_killed`, the boss HUD shows a compact reward panel with boss title,
  defeated status, and a short reward hint.
- The reward panel hides when a new live boss bar is shown or when the level/session clears.
- The existing boss health bar debug state exposes reward panel visibility, title, status, hint, and
  boss template id.
- Existing boss health, phase, and portrait behavior remains unchanged while a boss is alive.
- A focused Godot unit test proves reward panel show/hide state.
- A client bot scenario kills Cave Warden and asserts the visible reward panel state.

## Scope and Likely Files

- Client: `client/scripts/boss_health_bar.gd`, `client/scripts/boss_visuals_controller.gd`,
  `client/scripts/main.gd`, `client/scripts/bot_scenario_runner.gd`,
  `client/scripts/bot_step_catalog.gd`.
- Unit tests: `client/tests/test_boss_health_bar.gd`, `client/tests/test_client_bot.gd`.
- Bot/scenario: `tools/bot/scenarios/client/65_boss_reward_panel.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_boss_health_bar.gd`
- `godot --headless --path client --script res://tests/test_client_bot.gd`
- `make bot-client scenario=65_boss_reward_panel.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions.
- Risk: the client bot must kill the boss quickly and deterministically. Use the existing boss-floor
  lab and debug progression already proven by the boss special-drops protocol scenario if needed.
