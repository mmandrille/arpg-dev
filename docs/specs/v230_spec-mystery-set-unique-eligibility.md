# v230 Spec: Mystery Set and Unique Eligibility

Status: Complete
Date: 2026-06-16
Codename: mystery-set-unique-eligibility

## Purpose

Let the mystery seller's concealed offers become eligible for catalog-backed set and unique items,
so blind purchases can satisfy the ADR promise that all non-common rarities may appear.

## Baseline

Builds on v51 mystery seller core, the enabled named unique catalog, the enabled set packages, and
the v204 set collection presentation. Current mystery rows are server-owned and concealed, but their
rules cap `max_rarity` at `rare` and normal template rolling never produces set packages.

ADR alignment:
- ADR-0001: stock generation, purchase, reveal, inventory mutation, and persistence remain
  authoritative in Go.
- ADR-0013: mystery seller should eventually allow magic, rare, set, and unique outcomes without
  leaking identity before purchase.
- ADR-0014 D2/D3/D5/D12: set/unique eligibility preserves loot hope, gold sinks, behavior-changing
  uniques, and trade/endgame value.

Asset/plugin decision: adopt existing shared rules, named unique/set catalogs, concealed shop rows,
and bot assertions; reject external assets/plugins.

## Non-goals

- No new unique items, set packages, unique effects, art, silhouettes, stash overflow, market
  binding rules, or economy rebalance beyond allowing the existing rarity cap to reach set/unique.
- No item identity leak in concealed offers.
- No client-side stock generation or reveal authority.
- No guarantee that every shop open contains a set/unique row; this slice adds eligibility and
  focused deterministic proof.

## Acceptance Criteria

- `town_mystery_seller.mystery_offers.max_rarity` can be configured to allow `unique`/`set`.
- Mystery rolling can choose enabled named unique items and enabled set pieces whose base template
  slot matches the concealed offer slot and whose minimum level is within the source depth.
- Concealed offer rows still expose only safe metadata before purchase.
- Buying a special mystery row reveals a normal owned inventory item with the correct `rarity`,
  display name, stats, requirements, and effect IDs/set payload.
- Shared validation rejects unsupported mystery rarity caps and requires pricing multipliers for
  every rarity that can be priced.
- Focused Go tests prove set/unique candidate eligibility, cap exclusion at `rare`, and purchase
  reveal behavior for at least one special mystery row.
- Existing client mystery seller scenario still opens, renders concealed rows, and buys one offer.

## Scope and Likely Files

- `shared/rules/shops.v0.json`
- `shared/rules/shops.v0.schema.json`
- `server/internal/game/shop.go`
- `server/internal/game/mystery_shop.go`
- `server/internal/game/rules.go`
- `server/internal/game/mystery_shop_test.go`
- `tools/bot/scenarios/client/24_mystery_seller_core.json`
- Lifecycle docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/as-built/v230_mystery-set-unique-eligibility.md`

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'Mystery|ShopRules' -count=1`
- `make bot-client scenario=24_mystery_seller_core.json HEADLESS=1`
- `make maintainability`

Manual visual proof, if desired:

- `make bot-visual scenario=24_mystery_seller_core.json`

## Open Questions and Risks

- No blocking questions. This slice keeps row concealment unchanged and uses catalog eligibility,
  not a new resource, new item family, or new client reveal path.
