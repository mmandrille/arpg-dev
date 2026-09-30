# v488 As-Built — Live stash HTTP sync

- **Date:** 2026-09-30
- **Spec:** [`v488_spec-live-stash-http-sync.md`](../specs/v488_spec-live-stash-http-sync.md) ·
  **Plan:** [`v488_2026-09-30-live-stash-http-sync.md`](../plans/v488_2026-09-30-live-stash-http-sync.md)

## What shipped

**HTTP stash writes now reach the live session on the next tick, and replay reproduces them.**
- Before: market and account-stash routes wrote `account_stash_items` / `account_stash_gold` in
  Postgres, but a live session only loaded the stash at session start. A refunded, purchased or
  cancelled item did not appear until reconnect (`buyer_offer_cancel_ui` failed on this), and the
  sim kept offering items HTTP had already escrowed.
- Now: every stash-writing POST (market listing create/purchase/cancel, offer create/accept/cancel,
  account-stash upgrade/renew/merge) calls `Hub.NotifyAccountStashChanged` on success
  (`http/stash_sync.go`, wrapped inside `requireAuth`). Purchase also notifies the seller (stash gold
  credit) and accept the bidder (listing delivery).
- Each live loop takes its dirty accounts at the start of the next tick, after
  `flushDeferredPersist`, reads the stash rows and gold (250 ms bounded context), and records one
  stored-only **`system_account_stash_sync`** row per player of that account at that tick
  (`realtime/stash_sync.go`). Reading inside the loop, after the previous tick's persistence, means
  a deposit the sim applied but had not yet persisted can never be wiped by the sync.
- The sim splits these rows out of the tick's inputs (like `system_load_shed`) and applies them
  before player inputs, even for a dead player, without acks (`game/account_stash_sync.go`). It
  replaces the player's stash and emits the existing `stash_item_remove`, then `stash_item_add`
  (new or changed, by ascending ID), then `stash_gold_update`. No protocol change.
- `inputdecode` encodes/decodes the row (`system_stash_sync.go`); clients cannot submit it.
- A failed read is logged and metered and the flag is dropped, rather than retried and stalling
  every tick of an outage.
- Client: the market handler no longer patches local stash state after HTTP calls (listing
  create/cancel, stash offer, purchase, offer accept/cancel). The server ops are the single
  authority. `main.gd` 6,656 → 6,632.

## Proof

| Check | Result |
|-------|--------|
| `go test ./internal/game/ -run 'AccountStashSync\|PlayerIDsForAccount'` | PASS: diff order (remove, upsert, gold), unchanged stash emits nothing, dead player, actor-bearing row ignored, **sync + deposit in the same tick keeps both** |
| `go test ./internal/inputdecode/` | PASS: round trip, not a client intent, bad player rejected |
| `go test ./internal/realtime/ -run 'StashSync\|HTTPStash\|UnrecordedStash'` (+ `-race`) | PASS: live stash updated next tick, guest untouched, one recorded row, `replay.Reconstruct` stash equals live, `replay.Verify` matches; **the unrecorded variant diverges**; absent account reads nothing; read failure metered once and not retried. Red when the tick hook is removed. |
| `go test ./internal/http/ -run 'StashSync\|NotifyStash'` | PASS: 2xx notifies the caller, 4xx does not, no typed-nil hub |
| `go test ./...` against the per-checkout DB, `go vet`, gofmt, `make lint-determinism` | PASS |
| `make bot-client SCENARIO=buyer_offer_cancel_ui HEADLESS=1` | **PASS** (failed before this slice) |
| 18 other market/stash/blacksmith client scenarios (`account_stash_panel`, `market_purchase_ui`, `market_cancel_listing_ui`, `my_market_offers_panel`, `blacksmith_*`, …) | PASS |
| `ARPG_BOT_SCHEMA_VALIDATION=strict make bot scenario=account_stash_storage` | PASS |
| `make client-unit`, `make validate-shared`, `make maintainability`, `make ci` | PASS |

## Scope limits and follow-ups

- **Character inventory changed over HTTP is still not synced.** `publish_inventory` and
  `offer_inventory` (listing/offering straight from the character's bag) remove character items over
  HTTP while the live sim still holds them; the client only removes them locally. This is the more
  serious sibling of this slice: the sim can still use or persist those items. Next candidate.
- Bidders refunded by a listing cancel, and refunds from lazy listing expiry
  (`ExpireMarketListings` runs inside read paths), do not notify (accounts unknown at the HTTP
  layer). They resync on the next session build.
- Replay verification compares events; a stash sync emits changes only. The Go parity test compares
  the reconstructed stash directly. A future replay-contract ADR should decide whether state
  snapshots join the verify.
- Single server process only (the hub is in-memory).
