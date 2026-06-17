# v246 As-Built - Blacksmith Upgrade History

Date: 2026-06-17

## What shipped

- Added `BlacksmithUpgradeHistory`, a compact client-side history helper for recent upgrade attempts.
- Recorded upgrade attempts after authoritative `update_after_upgrade` results so each row includes
  recipe label, item name, success/failure wording, and gold spent.
- Kept history local to the current panel session, newest-first, hidden when empty, and capped at
  four entries.
- Exposed history visibility, row count, max entries, rows, and combined text in blacksmith debug
  state for focused tests.
- Added `63_blacksmith_upgrade_history.json` as the bot proof that an upgrade still completes with
  the history hook active.

## Proof

```bash
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=63_blacksmith_upgrade_history.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The selected v241-v250 batch-level
`make ci` also passed on 2026-06-17 after v250.

Manual visual proof, if desired:

```bash
make bot-visual scenario=63_blacksmith_upgrade_history.json
```

## Scope limits

- No server/protocol history, durable audit log, account-wide receipt, timestamps, filters,
  analytics, market integration, new recipes, balance changes, external assets, or external plugins
  shipped.
