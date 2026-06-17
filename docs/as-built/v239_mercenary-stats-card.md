# v239 As-Built - Mercenary Stats Card

Date: 2026-06-17

## What shipped

- Added a compact stats card to the mercenary panel for the first active hired companion.
- Card lines show only existing client companion state: display name, HP current/max, stance, active
  state, and entity id.
- The card hides when the roster is empty, including after `mercenary_lost` clears the active hire.
- Exposed `stats_card_visible`, `stats_card_text`, and `stats_card_lines` in mercenary panel debug
  state.
- Extended the focused mercenary bot assertion helper with `stats_card_contains`.
- Added `56_mercenary_stats_card.json`, which hires the fixed guard and verifies the stats card.

## Proof

```bash
godot --headless --path client --script res://tests/test_mercenary_panel.gd
make bot-client scenario=56_mercenary_stats_card.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v233-v240 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=56_mercenary_stats_card.json
```

## Scope limits

- No server/protocol changes, durable roster, mercenary gear, attack/armor exposure, recovery timers,
  per-mercenary commands, companion AI changes, external assets, or portrait art shipped.
