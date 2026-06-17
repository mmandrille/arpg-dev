# v236 As-Built - Mystery Seller Silhouettes

Date: 2026-06-17

## What shipped

- Added code-drawn mystery offer silhouettes derived only from visible slot/category metadata.
- Added `Silhouette: ...` to mystery offer detail/tooltip lines so concealed rows are easier to scan
  without revealing item identity.
- Exposed `mystery_silhouette` in shop debug offer rows.
- Kept mystery rows identity-safe: item definition/template, rarity, stats, requirements,
  comparison, and equip preview remain hidden.
- Added `53_mystery_seller_silhouettes.json`, which opens the mystery seller and verifies concealed
  rows include the silhouette clue.

## Proof

```bash
godot --headless --path client --script res://tests/test_shop_panel.gd
make bot-client scenario=53_mystery_seller_silhouettes.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The enclosing batch-level `make ci` is
deferred until the selected v233-v240 feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=53_mystery_seller_silhouettes.json
```

## Scope limits

- No server/protocol changes shipped.
- No external assets, plugins, or new art pipeline shipped.
- No mystery offer generation, pricing, purchase, reroll, reveal, comparison, or requirement-preview
  behavior changed.
