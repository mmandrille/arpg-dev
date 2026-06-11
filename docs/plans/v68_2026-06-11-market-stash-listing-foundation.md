# v68 Plan: Market Stash Listing Foundation

## Spec

[`docs/specs/v68_spec-market-stash-listing-foundation.md`](../specs/v68_spec-market-stash-listing-foundation.md)

## File Map

- `server/migrations/0016_market_listings.sql` — durable listings table.
- `server/internal/store/models.go`, `interfaces.go`, `repos.go`, `store_test.go` — listing model,
  atomic stash-to-listing and cancel-to-stash moves, tests.
- `server/internal/http/market.go`, `server/internal/http/server.go`, `auth_session_test.go` —
  authenticated market routes and HTTP proof.
- `PROGRESS.md`, `docs/as-built/v68_market-stash-listing-foundation.md` — close-out docs.

## Tasks

- [x] Add market listing persistence and store methods.
- [x] Add authenticated HTTP routes for list, create, and cancel.
- [x] Add store and HTTP tests for listing, browse, cancel, and foreign-cancel rejection.
- [x] Update docs and run focused verification plus `make ci`.

## Verification

```bash
(cd server && go test ./internal/store ./internal/http)
make ci
```
