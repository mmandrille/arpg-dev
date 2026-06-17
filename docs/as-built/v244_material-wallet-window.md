# v244 As-Built - Material Wallet Window

Date: 2026-06-17

## What shipped

- Added a text-first draggable `Material Wallet` window backed by the existing account
  `resource_wallet` client state.
- Let the compact `CharacterBar` wallet readout open the window when nonzero balances are visible.
- Rendered each nonzero resource with shared-rule display name, amount, category, and
  account-wide storage context.
- Kept empty/zero wallets hidden and closed, and refreshed the open window when wallet balances
  changed.
- Exposed wallet-window visibility, row count, rows, and text through character-bar debug state.
- Added `open_resource_wallet_window` client-bot action and extended `assert_resource_wallet_panel`
  with wallet-window assertions.

## Proof

```bash
godot --headless --path client --script res://tests/test_character_bar.gd
godot --headless --path client --script res://tests/test_client_bot.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=61_material_wallet_window.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The selected v241-v250 batch-level
`make ci` also passed on 2026-06-17 after v250.

Manual visual proof, if desired:

```bash
make bot-visual scenario=61_material_wallet_window.json
```

## Scope limits

- No server/protocol changes, wallet persistence changes, new resource types, drop tuning,
  resource trading, stash material tab, blacksmith recipe changes, production icons, external
  assets, or external plugins shipped.
- The compact HUD wallet text remains unchanged as the quick-scan surface.
