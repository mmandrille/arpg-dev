# v244 Plan - Material Wallet Window

Status: Complete
Goal: Add a draggable material wallet window opened from the compact HUD wallet readout.
Architecture: The server-owned account wallet and protocol stay unchanged. The existing
`CharacterBar` owns the compact wallet readout and will compose a focused `MaterialWalletPanel`
using shared item-rule metadata; bot state reads the same debug dictionary.
Tech stack: Godot UI/client bot, docs.

## Baseline and Asset Decision

Builds on v221 resource wallet persistence, v228 HUD wallet readout, v229 auto-pickup, and v237
catalog-backed wallet details.

Asset/plugin decision:
- Adopt existing `DraggableWindow` and `ItemRulesLoader`.
- Borrow existing `CharacterBar` wallet row/detail formatting and `assert_resource_wallet_panel`.
- Reject external assets/plugins and production material icons for this slice.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/material_wallet_panel.gd` | Draggable wallet detail window |
| Modify | `client/scripts/character_bar.gd` | Open/sync the material wallet window from the HUD wallet readout |
| Modify | `client/tests/test_character_bar.gd` | Unit proof for window open/refresh/empty behavior |
| Modify | `client/scripts/bot_controller.gd` | Tiny bot action dispatch for opening the HUD wallet window |
| Modify | `client/scripts/bot_step_catalog.gd` | Register/validate bot action and assertion keys |
| Modify | `client/scripts/bot_action_step_validator.gd` | Accept wallet-window open action |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Assert wallet-window visibility/row text |
| Add | `tools/bot/scenarios/client/61_material_wallet_window.json` | Client scenario proof |
| Add | `docs/as-built/v244_material-wallet-window.md` | As-built proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/bot_controller.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected

Decision:
- [x] Add the new UI in `material_wallet_panel.gd`, not in `main.gd` or a large existing panel.
- [x] Keep the `bot_controller.gd` change to the smallest action dispatch needed for bot proof.

Verification:
```bash
make maintainability
```

## Task 1 - Wallet window UI

Files:
- Add: `client/scripts/material_wallet_panel.gd`
- Modify: `client/scripts/character_bar.gd`

- [x] Build a small draggable `Material Wallet` window that renders nonzero wallet rows.
- [x] Let the HUD wallet readout open the window on click while empty wallets remain hidden/closed.
- [x] Refresh an already-open window when `set_resource_wallet` receives new balances.
- [x] Expose window visibility and rows through `CharacterBar.get_debug_state()`.

```bash
godot --headless --path client --script res://tests/test_character_bar.gd
```

## Task 2 - Tests and bot proof

Files:
- Modify: `client/tests/test_character_bar.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_action_step_validator.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Add: `tools/bot/scenarios/client/61_material_wallet_window.json`

- [x] Extend the character-bar unit test for wallet-window open, refresh, and empty-close behavior.
- [x] Add `open_resource_wallet_window` client-bot action.
- [x] Extend `assert_resource_wallet_panel` with wallet-window visibility/text expectations.
- [x] Add a scenario that auto-picks an upgrade shard, opens the window, and asserts row text.

```bash
godot --headless --path client --script res://tests/test_client_bot.gd
make bot-client scenario=61_material_wallet_window.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v244_material-wallet-window.md`

- [x] Record focused verification and deferred scope.

## Final Verification

- [x] `godot --headless --path client --script res://tests/test_character_bar.gd`
- [x] `godot --headless --path client --script res://tests/test_client_bot.gd`
- [x] `make bot-client scenario=61_material_wallet_window.json HEADLESS=1`
- [x] `make maintainability`
