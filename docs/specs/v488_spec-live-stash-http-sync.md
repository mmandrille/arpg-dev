# v488 — Live stash HTTP sync: account-stash changes made over HTTP reach the live session

- **Status:** Implemented (v488)
- **Date:** 2026-09-30
- **Codename:** `live-stash-http-sync`
- **Baseline:** v487 `kit-ground-loot` (`e94eebfe`)
- **Related:** v234 buyer offer cancel UI; replay system-input pattern v476–v481; v486 review rec #3
  (`buyer_offer_cancel_ui`, the last undiagnosed ci-full failure) and theme 1 (replay depends only on
  recorded data)

## Purpose

Market and account-stash HTTP routes write `account_stash_items` / `account_stash_gold` in Postgres.
A live session loads the account stash into the sim once, at session start
(`sessionsetup` → `LoadAccountStashForPlayer`), and never hears about later HTTP writes. So:

- Cancel an offer, buy a listing, cancel a listing: the item lands in the stash in the DB, but the
  open session's stash does not show it until the player reconnects. `buyer_offer_cancel_ui`
  times out on exactly this.
- v234 papered over it on the client (`main.gd` upserts the HTTP response items into local stash
  state), but `stash_opened` renders the sim's stash, so the item still vanishes.
- The sim's stale copy is a correctness risk: it can still offer an item HTTP already escrowed into
  a market offer or listing.

After this slice, a successful stash-mutating HTTP call marks the account's stash **dirty** on every
live session loop that holds a player of that account. At the start of its next tick the loop reads
the account's stash rows and gold from the store and records a stored-only, server-authored
**`system_account_stash_sync`** input for that tick. The sim replaces that player's stash from the
payload and emits the existing `stash_item_remove` / `stash_item_add` / `stash_gold_update`
changes. Replay applies the same recorded payload at the same tick.

### Design decisions

- **Dirty flag, loop-side read.** HTTP does not snapshot the stash itself. The loop reads the stash
  at tick start, holding its lock, after `flushDeferredPersist`. The previous tick's
  `persistTick` has then finished, so every sim-originated stash write is already in the DB. A
  snapshot taken by HTTP could miss a deposit the sim had applied but not yet persisted, and the
  sync would then wrongly remove it.
- **Full replace, server-side diff** (brief Q-2). The payload carries the whole stash
  (capacity ≤ 50). Diffing current vs new by `stash_item_id` emits removals (sorted IDs), then
  additions or in-place updates (sorted IDs, detected by comparing the marshalled
  `StashItemView`), then a gold update if gold changed.
- **Applied at tick start, outside `applyInput`.** Like `system_load_shed`, the sync inputs are split
  out of the tick's inputs (`splitAccountStashSyncInputs`) and applied before player inputs. They
  apply even if the player is dead (the stash is account state), and never produce acks or rejects.
  `applyInput` is not edited.
- **Stored-only.** Absent from `IsClientIntent` / `Decode`. Clients can never submit it; only the
  persisted input stream carries it (`inputdecode/system.go`). Only actor-less rows count.
- **No protocol change.** Existing stash change ops, fanned out to the owning player by
  `filterChangesForClient`. Stash change ops are not persisted from the change list (stash writes
  persist via events), so the sync cannot double-write DB rows.
- **ID safety.** Market refunds/deliveries reuse the item's original numeric `stash_item_id`, so
  `LoadAccountStash`'s parse keeps them. It bumps `nextID` past loaded IDs, which is deterministic
  because the payload is recorded. Withdrawals allocate a new instance ID, and stash IDs are unique
  per account (DB primary key), so a delivered ID cannot collide inside the stash.
- **Client** (brief Q-4): drop the v234 local `_upsert_stash_item` of HTTP response items for
  offer accept/cancel. The server op is the single authority, one tick later.

### Routes in scope (brief Q-1, Q-3)

Every successful (2xx) POST that writes an account stash, for the calling account:
`POST /v0/market/listings` (create, escrows a stash item),
`/listings/{id}/purchase`, `/listings/{id}/cancel`, `/listings/{id}/offers` (create, escrows),
`/offers/{id}/accept`, `/offers/{id}/cancel`, and `POST /v0/account-stash/*` (upgrade, renew,
upgrade-shard and leveled-consumable merge). Counterpart accounts whose stash the same call
writes, known from the response: the **seller** on purchase (gold credit) and the **bidder** on
offer accept (listing delivery). Every live loop holding a player of a notified account syncs.

## Non-goals

- Bidders refunded by a listing cancel, and refunds from lazy listing expiry
  (`ExpireMarketListings` runs inside read paths). Their accounts are not known at the HTTP layer;
  they resync on their next session build. Follow-up if it matters.
- Syncing character inventory, gold or the resource wallet/bag from HTTP (e.g. inventory item
  upgrade changes the character item; only its stash side effects sync here).
- Cross-process session ownership (single server process only).
- A protocol bot scenario: the bot has no in-session HTTP market step, and adding one means growing
  the frozen, grandfathered `tools/bot/run.py`. Proof is Go loop tests plus the real-client
  scenario 51.

## Acceptance criteria

- [ ] After a successful offer cancel over HTTP, the bidder's live session emits
  `stash_item_add` for each refunded item on the next tick, and a later `stash_opened` lists them.
- [ ] Purchase: the buyer's stash gains the item and its gold drops; a live seller's stash gold
  rises. Accept: the live bidder receives the listing item.
- [ ] A stash change the sim applied before the HTTP call (deposit) survives the sync (not removed).
- [ ] `replay.Verify` matches for a session with a mid-session sync. A live-vs-replay parity test is
  red when the sync is applied live but not recorded.
- [ ] `inputdecode` rejects `system_account_stash_sync` from clients (`IsClientIntent` false,
  `Decode` fails) and round-trips it from storage.
- [ ] A notified account with no live session is a no-op; a loop without that account is untouched.
- [ ] `make bot-client SCENARIO=buyer_offer_cancel_ui HEADLESS=1` passes (it fails today).
- [ ] `make lint-determinism` clean (sorted diff iteration); `make test-go`, `make client-unit` green.

## Scope and files likely touched

| Area | Files |
|------|-------|
| Sim | new `server/internal/game/account_stash_sync.go` (+ test); `input.go` (`StashSync` field); `tick_results.go` (split + apply at tick start) |
| Decode | `server/internal/inputdecode/system.go` (type, encode, decode) + test |
| Realtime | new `server/internal/realtime/stash_sync.go` (dirty set, tick-start read + record, `Hub.NotifyAccountStashChanged`) + test; `session_tick.go` (one call) |
| HTTP | new `server/internal/http/stash_sync.go` (2xx wrapper + counterpart notify); route registrations in `market.go`, `account_stash*.go` |
| Client | `client/scripts/main.gd` (drop the v234 local upsert, net-shrink) |
| Docs | `docs/as-built/v488_live-stash-http-sync.md`, lifecycle row, `PROGRESS.md`, `docs/CODEMAP.md` |

## Test and bot proof

- Go: sim diff unit test (remove/add/update/gold, sorted, dead player); decode round-trip and
  client rejection; realtime loop tests with the in-memory repo: HTTP-style notify → next tick emits
  stash ops to the owner only, deposit-then-sync keeps the deposit, `replay.Verify` matches, and a
  red variant that applies the sync without recording it.
- HTTP: a test that a 2xx stash route notifies the hub and a 4xx does not.
- Client bot: scenario 51 `buyer_offer_cancel_ui`.

## Open questions and risks

- **DB read on the tick path.** Only on ticks with a dirty account, bounded by a short context
  timeout (250 ms). On read failure the error is logged and metered, the dirty flag is dropped (a
  retry every tick would stall ticks for the whole outage), and the tick proceeds without a sync.
  The next HTTP stash change or session build resyncs.
- **Two syncs in one tick** for the same player (two accounts can't share a player, but two notifies
  can): the dirty set dedupes per account.
