# v228 As-Built - Resource Wallet Panel

Date: 2026-06-16

## What shipped

- Added a compact account-resource wallet readout to the existing `CharacterBar` HUD.
- Synced the client-owned `resource_wallet` state into the HUD after inventory/UI refreshes, covering
  snapshot and `resource_wallet_update` paths without server or protocol changes.
- Hid the wallet readout for empty/zero balances and exposed `wallet_visible`, `wallet_text`, and
  `wallet_rows` in character-bar debug state.
- Added a reusable `assert_resource_wallet_panel` client-bot assertion and extended the blacksmith
  upgrade scenario to prove shard pickup shows `Shard 1` and spending it hides the readout.

## Proof

```bash
make client-unit
make bot-client scenario=39_blacksmith_upgrade_ui.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=39_blacksmith_upgrade_ui.json
```

## Scope limits

- No server, protocol, shared-rule, resource ownership, stash, market, or inventory footprint changes
  shipped.
- No production resource icons, art assets, external UI plugins, or material auto-pickup behavior
  shipped; v229 owns material auto-pickup.
