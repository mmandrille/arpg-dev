# v488 Plan — Live stash HTTP sync

Status: Implemented
Spec: [`docs/specs/v488_spec-live-stash-http-sync.md`](../specs/v488_spec-live-stash-http-sync.md)
Goal: stash changes committed over HTTP appear in the account's live session on the next tick,
recorded so replay stays exact.
Architecture: HTTP marks the account dirty on every live loop (`Hub.NotifyAccountStashChanged`). At
tick start the loop reads stash rows + gold from the store (bounded ctx), records a stored-only
`system_account_stash_sync` row per affected player at the current tick, and passes it with the
tick's inputs. The sim splits these inputs out (like `system_load_shed`), replaces the player's
stash and emits the existing stash change ops.
Tech stack: Go sim + realtime + http, GDScript client, Godot client bot.

## Baseline and shortcut decision

- Reuses: the v476–v481 stored system-input pattern (`inputdecode/system.go`,
  `systemInputRowLocked`, `persistSystemInput`), `LoadAccountStash`, `stashItemView`,
  `filterChangesForClient` owner scoping, and the realtime in-memory repo test harness
  (`load_shed_replay_test.go`).
- No client asset or plugin work (the client change only deletes code).
- Spec review (2026-09-30): PASS. `playerState.AccountID` maps accounts to players; the stash
  conversion lives in `sessionsetup/convert.go` (to be exported); stash change ops are not persisted
  from the change list, and refund/delivery rows keep numeric stash IDs.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `server/internal/game/account_stash_sync.go` | type const, `AccountStashSync`, `splitAccountStashSyncInputs`, `applyAccountStashSyncs` (diff + replace), `PlayerIDsForAccount` |
| Create | `server/internal/game/account_stash_sync_test.go` | diff order, update detection, gold, dead player, actor-bearing input ignored |
| Modify | `server/internal/game/input.go`, `tick_results.go` | `StashSync` field; split + apply at tick start |
| Modify | `server/internal/inputdecode/system.go` (+ test) | `TypeSystemAccountStashSync`, encode/decode, client rejection |
| Modify | `server/internal/sessionsetup/convert.go`, `sessionsetup.go` | export `PersistedStashItems` |
| Create | `server/internal/realtime/stash_sync.go` | dirty set, `Hub.NotifyAccountStashChanged`, `stashSyncInputsLocked` (tick-start read + record) |
| Create | `server/internal/realtime/stash_sync_test.go` | owner-only ops, deposit survives, replay parity, red variant, no-op for absent account |
| Modify | `server/internal/realtime/session_tick.go` | one call before `TickResultsProfiled` |
| Create | `server/internal/http/stash_sync.go` (+ test) | 2xx wrapper, counterpart notify |
| Modify | `server/internal/http/market.go`, `account_stash.go`, `account_stash_merge.go`, `account_stash_renew.go` | wrap stash-writing POST routes (same lines); counterpart notify in purchase/accept |
| Modify | `client/scripts/main.gd` | drop the v234 local stash upsert after offer accept/cancel |
| Modify | docs (`CODEMAP.md`, as-built, lifecycle, `PROGRESS.md`) | close-out |

## Maintenance ratchet

- [x] `client/scripts/main.gd` — touched; net shrink (lower baseline)
- [x] `server/internal/game/game_test.go` — not touched
- [x] `tools/bot/run.py` — not touched (no protocol-bot step, spec non-goal)
- [x] `tools/validate_shared.py` — not touched
- [x] Other over-limit: none; `http/account_stash.go` (560) route edits stay on the same lines

```bash
make maintainability
```

## Task 1 — Sim: account stash sync input

- [x] 1.1 `SystemAccountStashSyncInputType = "system_account_stash_sync"`; `AccountStashSync{PlayerID, Items []PersistedStashItem, Gold int}`; `Input.StashSync`.
- [x] 1.2 `splitAccountStashSyncInputs`: keep only actor-less rows with a payload.
- [x] 1.3 `applyAccountStashSyncs(ctx)`: per sync (input order), `usePlayer`; build the new stash via `LoadAccountStash(items, gold, current capacity)`; emit removes (sorted old IDs not in new), then adds/updates (sorted new IDs absent or whose marshalled `StashItemView` differs), then `stash_gold_update` if gold changed; `savePlayer`; restore the default player.
- [x] 1.4 `PlayerIDsForAccount(accountID)` sorted.
- [x] 1.5 Wire into `TickResultsProfiled` right after `newSimTickCtx`.

```bash
cd server && go test ./internal/game/ -run 'AccountStashSync' -count=1 && make -C .. lint-determinism
```

## Task 2 — Decode

- [x] 2.1 Type constant, `EncodeStoredAccountStashSync(messageID, sync)`, decode branch.
- [x] 2.2 Tests: stored round-trip; `IsClientIntent` false; client `Decode` fails.

```bash
cd server && go test ./internal/inputdecode/ -count=1
```

## Task 3 — Realtime loop + hub

- [x] 3.1 `sessionLoop.stashDirty map[string]bool` (under `l.mu`); `markAccountStashDirty`.
- [x] 3.2 `Hub.NotifyAccountStashChanged(accountIDs ...string)`: snapshot loops under `h.mu`, mark each.
- [x] 3.3 `stashSyncInputsLocked(tick)`: for sorted dirty accounts, read `ListAccountStashItems` + `GetOrCreateAccountStashGold` (250 ms ctx); on error log, meter and drop; for each player of the account build the input, record via `systemInputRowLocked(tick, …)` + `persistSystemInput`; clear the flag.
- [x] 3.4 `session_tick.go`: append the returned inputs to the tick's inputs after `flushDeferredPersist`.
- [x] 3.5 Tests (in-memory repo): owner-only stash ops next tick; deposit-then-sync keeps the deposit; `replay.Verify` match; red variant (apply without recording → mismatch); absent account no-op.

```bash
cd server && go test ./internal/realtime/ -run 'StashSync' -count=1 && go test -race ./internal/realtime/ -run 'StashSync' -count=1
```

## Task 4 — HTTP

- [x] 4.1 `withStashSync(h)`: wraps the handler, records the status, and on 2xx notifies the caller's account.
- [x] 4.2 Wrap the stash-writing POST routes; `handlePurchaseMarketListing` also notifies `listing.SellerAccountID`, `handleAcceptMarketOffer` `offer.BidderAccountID`.
- [x] 4.3 Test: 2xx notifies, 4xx does not.

```bash
cd server && go test ./internal/http/ -run 'StashSync' -count=1
```

## Task 5 — Client

- [x] 5.1 Remove the local `_upsert_stash_item` loop and the related refresh in the offer accept/cancel branch of `main.gd`.

```bash
make client-unit
```

## Task 6 — Bot proof

- [x] `make bot-client SCENARIO=buyer_offer_cancel_ui HEADLESS=1` (red before, green after).
- [x] Regression: `make bot-client SCENARIO=account_stash_panel HEADLESS=1`, `SCENARIO=market_board_ui`.

## Task 7 — Lifecycle docs and CI

- [x] As-built, lifecycle row, `PROGRESS.md` (v486 follow-up: `buyer_offer_cancel_ui` closed; `origin` credential owner-blocked by network infra), CODEMAP.

```bash
make ci
```

## Execution notes (2026-09-30)

- The dirty set lives on the `Hub` (`stashDirty`, keyed by loop, cleared in `removeLoop`) instead of
  a `sessionLoop` field, to keep grandfathered `session_loop.go` untouched.
- HTTP reaches the hub through a one-method `stashNotifier` seam (testable, no typed-nil `*Hub`),
  reusing the existing `statusRecorder`.
- Client: besides the v234 offer accept/cancel upsert, four more market branches (stash listing
  create, listing cancel, stash offer, purchase) patched local stash state after HTTP; all removed,
  plus the now-dead `_remove_market_stash_item` and `_upsert_stash_item` wrappers.
- Finding for the follow-up: `publish_inventory` / `offer_inventory` change character inventory over
  HTTP with the same live-sim staleness (as-built follow-up).
