# Spec: `mystery-seller-paid-reroll`

Status: Accepted
Date: 2026-06-11
Branch: `main`
Codename: `mystery-seller-paid-reroll`
Slice: v64 - mystery seller paid reroll
Baseline: v63 `runtime-sim-error-construction`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0013-mystery-seller-and-unidentified-item-offers.md`](../adr/0013-mystery-seller-and-unidentified-item-offers.md)
- [`../adr/0014-core-progression-and-endgame-design-rules.md`](../adr/0014-core-progression-and-endgame-design-rules.md)
- [`v51_spec-mystery-seller-core.md`](v51_spec-mystery-seller-core.md)

## 1. Purpose

The mystery seller is a blind-buy gold sink, but stock currently changes only when waypoint progress
changes. This slice adds a small paid reroll: while in town, a player can spend gold to replace the
current character-scoped mystery seller stock with a fresh deterministic hidden stock set.

This advances ADR-0013's paid refresh direction and ADR-0014's rule that gold should stay valuable,
without adding timers, daily clocks, new resources, or account-wide stock.

## 2. Non-goals

- No timer, daily, clock-based, town-visit, or account-wide mystery refresh.
- No new currency or resource.
- No visible item identity before purchase.
- No final economy tuning beyond a conservative fixed first-slice reroll cost.
- No reroll for normal town vendor stock or buyback.
- No stash overflow, refunds, binding, unique/set eligibility, or market rules.
- No separate mystery seller UI panel; reuse the existing shop panel.

## 3. Acceptance Criteria

1. Shared shop rules define a positive `reroll_cost` for mystery offers.
2. Protocol messages allow `shop_reroll_intent` with `shop_entity_id`.
3. The server registers `shop_reroll_intent` in the handler registry.
4. Reroll validates that the target is the actor's reachable/openable mystery seller, not a normal
   vendor.
5. Reroll rejects without mutation when the shop is not a mystery seller or the character lacks
   enough gold.
6. Successful reroll subtracts character gold, updates durable character gold/progression, replaces
   all available mystery stock rows with a new deterministic set, and emits actor-private
   `gold_update`, `character_progression_update`, `shop_stock_replace`, and `shop_reroll` event
   payloads.
7. The replacement stock uses a deterministic refresh key that includes the current waypoint refresh
   key plus a monotonic paid reroll index, so repeated paid rerolls produce different hidden rows
   and replay remains deterministic.
8. Purchased/consumed rows from the old stock do not carry into the new rerolled stock.
9. Reconnect, fresh session, `/state`/shop reopen, and replay see the rerolled stock, not the
   pre-reroll stock.
10. The Godot shop panel shows a reroll control only for mystery seller rows, disables it when gold
    is insufficient, sends `shop_reroll_intent`, and refreshes visible concealed rows after the
    server event.
11. Protocol and client bot scenarios can open the mystery seller, reroll once, assert gold spend,
    assert stock replacement, and then buy from the new concealed stock.
12. `make ci` passes.

## 4. Scope And Likely Files

```text
shared/rules/shops.v0.json
shared/rules/shops.v0.schema.json
shared/protocol/messages.v8.schema.json
shared/protocol/examples/mystery_shop_reroll_intent.json
server/internal/game/types.go
server/internal/game/rules.go
server/internal/game/shop.go
server/internal/game/handlers.go
server/internal/game/shop_test.go
client/scripts/shop_panel.gd
client/scripts/bot_controller.gd
client/scripts/bot_scenario_runner.gd
client/scripts/main.gd
client/tests/test_shop_panel.gd
client/tests/test_client_bot.gd
tools/bot/run.py
tools/bot/test_protocol.py
tools/bot/scenarios/39_mystery_seller_paid_reroll.json
tools/bot/scenarios/client/29_mystery_seller_paid_reroll.json
PROGRESS.md
docs/as-built/v64_mystery-seller-paid-reroll.md
```

## 5. Test And Bot Proof

- Shared validation covers the shop rule and new protocol intent example.
- Go tests cover successful reroll, insufficient gold rejection, non-mystery shop rejection,
  deterministic replacement stock, and consumed old-stock reset.
- Protocol bot proves open -> reroll -> gold spend -> stock replacement -> buy.
- Client unit/bot tests prove the mystery reroll control and emitted intent.
- Full `make ci` is the landing gate.

## 6. Open Questions And Risks

- Reroll cost is conservative for this slice: use a fixed rule value (`50`) and defer final pricing
  against visible vendor prices.
