# v230 Plan - Mystery Set and Unique Eligibility

Status: Complete
Goal: Allow mystery seller concealed stock to include catalog-backed set and unique outcomes.
Architecture: Extract mystery seller rolling from `shop.go`, broaden shared rarity caps, then use
existing named unique and set payload builders for revealed purchase items.
Tech stack: Shared JSON rules/schema, Go authoritative shop generation, Godot client bot proof, SDD
docs.

## Baseline and shortcut decision

Reuse the existing concealed `ShopOfferView` contract and existing `ItemRollPayload` shape. Special
items should enter inventory through the same purchase route as current mystery items. Asset/plugin
decision: adopt existing shared catalogs and code-native UI; reject external assets/plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/shops.v0.json` | Allow mystery seller `max_rarity` to reach set/unique and price set rows |
| Modify | `shared/rules/shops.v0.schema.json` | Validate set/unique mystery caps and pricing |
| Modify | `server/internal/game/rules.go` | Validate mystery caps and pricing across special rarities |
| Modify | `server/internal/game/shop.go` | Keep generic shop logic, move mystery-specific rolling out |
| Add | `server/internal/game/mystery_shop.go` | Mystery source depth, cap checks, and set/unique candidate rolling |
| Add | `server/internal/game/mystery_shop_test.go` | Focused mystery set/unique eligibility and reveal tests |
| Verify | `tools/bot/scenarios/client/24_mystery_seller_core.json` | Existing client proof remains concealed and buyable |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `shop.go` baseline if extraction shrinks it |
| Modify | `PROGRESS.md` | Current status after completion |
| Modify | `docs/progress/slice-lifecycle.md` | Lifecycle row |
| Add | `docs/as-built/v230_mystery-set-unique-eligibility.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/shop.go`
- [x] `server/internal/game/shop_test.go` was not touched; tests went into `mystery_shop_test.go`
- [x] `server/internal/game/rules.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none touched
- [x] Did every touched grandfathered file stay within the ratchet? `shop.go` dropped to 1118
  lines and `rules.go` grew by 3 lines, within the allowed +25 budget.

Decision:
- [x] Extract focused helper/module/test file as part of this slice: move mystery seller rolling
  helpers from `shop.go` into `mystery_shop.go`, then lower the `shop.go` baseline if applicable.
- [ ] Defer extraction with rationale: not needed because extraction shipped.

Verification:
```bash
make maintainability
```

## Task 1 - Shared rarity cap and pricing rules

Files:
- Modify: `shared/rules/shops.v0.json`
- Modify: `shared/rules/shops.v0.schema.json`
- Modify: `server/internal/game/rules.go`

- [x] Step 1.1: Allow mystery `max_rarity` to include set/unique while keeping generated vendor
  stock capped at rare.
- [x] Step 1.2: Add set pricing support and validation.
```bash
make validate-shared
```

## Task 2 - Mystery special candidate generation

Files:
- Modify: `server/internal/game/shop.go`
- Add: `server/internal/game/mystery_shop.go`
- Add: `server/internal/game/mystery_shop_test.go`

- [x] Step 2.1: Extract existing mystery rolling helpers from `shop.go`.
- [x] Step 2.2: Add deterministic set/unique candidate selection by eligible slot and source depth.
- [x] Step 2.3: Prove rare cap excludes special candidates and unique/set cap can reveal a special
  owned item without leaking concealed offer identity.
```bash
cd server && go test ./internal/game -run 'Mystery|ShopRules' -count=1
```

## Task 3 - Client and lifecycle proof

Files:
- Verify: `tools/bot/scenarios/client/24_mystery_seller_core.json`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v230_mystery-set-unique-eligibility.md`

- [x] Step 3.1: Keep the client mystery seller flow green with concealed rows and one purchase.
- [x] Step 3.2: Record v230 as complete with focused proof and note the final batch CI is pending.
```bash
make bot-client scenario=24_mystery_seller_core.json HEADLESS=1
make maintainability
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'Mystery|ShopRules' -count=1`
- [x] `make bot-client scenario=24_mystery_seller_core.json HEADLESS=1`
- [x] `make maintainability`
- [x] Batch-level `make ci` passed after the selected v226-v232 `$autoloop` queue.
