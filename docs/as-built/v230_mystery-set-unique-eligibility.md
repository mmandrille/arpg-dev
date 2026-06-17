# v230 As-Built - Mystery Set and Unique Eligibility

Date: 2026-06-16

## What shipped

- Broadened the mystery seller rarity cap so concealed offers can be eligible for catalog-backed
  unique and set payloads while normal generated vendor stock remains capped at rare.
- Added set pricing support to shop rules/schema validation so special mystery rows can use the
  existing generated-item pricing path.
- Extracted mystery seller stock rolling from `shop.go` into `mystery_shop.go`, lowering the
  `shop.go` line-count baseline from 1221 to 1118.
- Added set/unique candidate selection by concealed slot and source depth using existing named
  unique and set item payload builders.
- Preserved pre-purchase concealment: shop rows still expose only slot/source/depth/price metadata,
  and purchase reveals the owned inventory item's rarity, name, stats, requirements, and effects.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'Mystery|ShopRules' -count=1
make bot-client scenario=24_mystery_seller_core.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. `make maintainability` printed a
transient YARA sync warning from the backend, then passed the local ratchets. The enclosing
batch-level `make ci` is deferred until the selected feature queue completes.

Manual visual proof, if desired:

```bash
make bot-visual scenario=24_mystery_seller_core.json
```

## Scope limits

- No new unique items, set packages, unique effects, art, silhouettes, stash overflow, market
  binding rules, or economy rebalance shipped.
- No guaranteed set/unique stock on every shop open shipped; this slice adds eligibility.
- No client-authoritative stock generation or pre-purchase identity leak shipped.
