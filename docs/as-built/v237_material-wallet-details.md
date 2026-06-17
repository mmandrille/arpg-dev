# v237 As-Built - Material Wallet Details

Date: 2026-06-17

## What shipped

- Added catalog-backed material wallet tooltip details to the existing `CharacterBar` HUD readout.
- Kept the compact wallet text unchanged, so `upgrade_shard` still displays as `Shard N` in the HUD.
- Added detail lines with full shared item-rule name, count, category, and account-wide storage
  context.
- Exposed `wallet_tooltip` and `wallet_details` in character-bar debug state.
- Extended `assert_resource_wallet_panel` with `tooltip_contains`.
- Added `54_material_wallet_details.json`, which auto-picks an upgrade shard and verifies the HUD
  tooltip detail.

## Proof

```bash
godot --headless --path client --script res://tests/test_character_bar.gd
make bot-client scenario=54_material_wallet_details.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v233-v240 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=54_material_wallet_details.json
```

## Scope limits

- No server/protocol changes, wallet persistence changes, new resource types, resource trading,
  stash material tab, blacksmith recipe changes, icons, or external assets shipped.
- No standalone wallet window shipped.
