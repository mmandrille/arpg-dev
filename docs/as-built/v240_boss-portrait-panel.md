# v240 As-Built - Boss Portrait Panel

Date: 2026-06-17

## What shipped

- Added a code-drawn boss portrait tile to the existing top-center `BossHealthBar`.
- Keyed the portrait from `boss_template_id`, with a distinct Cave Warden portrait and a safe generic
  boss fallback.
- Exposed `portrait_visible`, `portrait_kind`, and `portrait_label` in boss health bar debug state.
- Extended boss health bar bot assertions and step validation with portrait fields.
- Preserved existing boss HP, phase countdown, phase ratio, and visibility behavior.
- Added `57_boss_portrait_panel.json`, which starts on the compact boss floor and verifies the Cave
  Warden portrait tile.

## Proof

```bash
godot --headless --path client --script res://tests/test_boss_health_bar.gd
godot --headless --path client --script res://tests/test_client_bot.gd
make bot-client scenario=57_boss_portrait_panel.json HEADLESS=1
make maintainability
make ci
```

All focused checks passed on 2026-06-17 during `$autoloop`. Batch-level `make ci` also passed after
the selected v233-v240 feature queue completed.

Manual visual proof, if desired:

```bash
make bot-visual scenario=57_boss_portrait_panel.json
```

## Scope limits

- No server/protocol changes, imported portrait art, asset manifest changes, model rendering,
  animation, boss combat changes, loot changes, audio changes, multi-boss selector, boss codex, or
  reward panel shipped.
