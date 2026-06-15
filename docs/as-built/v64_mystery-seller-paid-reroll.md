# v64 As-Built - Mystery Seller Paid Reroll

Status: Complete
Date: 2026-06-11
Branch: `main`
Spec: [`../specs/v64_spec-mystery-seller-paid-reroll.md`](../specs/v64_spec-mystery-seller-paid-reroll.md)
Plan: [`../plans/v64_2026-06-11-mystery-seller-paid-reroll.md`](../plans/v64_2026-06-11-mystery-seller-paid-reroll.md)

## What Shipped

- Added `reroll_cost: 50` to mystery seller rules and schema validation.
- Added protocol v8 `shop_reroll_intent` plus a validated example.
- Implemented server-authoritative mystery-only paid reroll:
  - validates reachable mystery seller target and sufficient gold,
  - spends character gold and persists progression updates,
  - replaces current mystery stock through existing shop stock persistence,
  - uses deterministic refresh keys with `|reroll:N` suffixes,
  - emits `shop_reroll` with refreshed concealed offers and total gold.
- Added input decoding and Go tests for success, insufficient-gold rejection, and normal-vendor rejection.
- Added a mystery-seller-only Godot shop reroll button, debug state, client status refresh, and bot click path.
- Added protocol and Godot client bot scenarios proving reroll -> refreshed stock -> mystery purchase.

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game ./internal/inputdecode -run 'TestMystery.*Reroll|TestDecodeShopReroll|TestShopRules|TestNewSim' -count=1`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make client-unit`
- `make bot scenario=mystery_seller_paid_reroll`
- `make bot-client scenario=29_mystery_seller_paid_reroll`
- `make test-go`
- `make ci`

Full `make ci` passed with 9 phases.

## Notes

  the existing server-backed `ShopPanel` already owned the relevant UI and interaction contract.
- Scenario funding sells three vendor-lab rolled items before rerolling so the bot can still afford
  a post-reroll mystery purchase.
