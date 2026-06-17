# v249 As-Built - Boss Reward Panel

Date: 2026-06-17

## What shipped

- Added a compact reward panel state to the existing boss HUD component.
- Displayed the reward panel from the existing `boss_killed` client event path with boss title,
  defeated status, and the hint `Exit unlocked - claim the boss chest`.
- Kept live boss health, phase, and portrait behavior separate from post-kill reward panel state.
- Exposed reward panel visibility, title, status, hint, and boss template id through boss HUD debug
  state.
- Extended boss health bar bot assertions with reward panel fields.
- Let client bot scenarios pass full `debug_progression` JSON into Godot, normalizing parsed numeric
  maps back to integers before calling the server debug progression endpoint.
- Added `65_boss_reward_panel.json`, which seeds a deterministic boss-proof character, kills Cave
  Warden, and asserts the reward panel.

## Proof

```bash
godot --headless --path client --script res://tests/test_boss_health_bar.gd
godot --headless --path client --script res://tests/test_client_bot.gd
bash -n scripts/bot_client.sh
make bot-client scenario=65_boss_reward_panel.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The selected v241-v250 batch-level
`make ci` also passed on 2026-06-17 after v250.

Manual visual proof, if desired:

```bash
make bot-visual scenario=65_boss_reward_panel.json
```

## Scope limits

- No server/protocol changes, loot table changes, XP tuning, boss balance, chest animation, audio,
  multi-boss reward summary, item preview, external art, or external plugins shipped.
