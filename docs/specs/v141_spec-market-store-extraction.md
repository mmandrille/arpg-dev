# v141 Spec: Market Store Extraction

Status: Complete
Date: 2026-06-13
Codename: `market-store-extraction`

## Purpose

Move the market listing, offer, expiration, audit, and summary persistence methods out of
`server/internal/store/repos.go` into a focused market repository file. This is a behavior-preserving
maintainability slice that repays the v139 ratchet exception and makes future market work load the
market domain directly instead of a broad store coordinator.

## Non-goals

- No market behavior, SQL contract, HTTP route, protocol schema, or migration changes.
- No time-based market sweep beyond the existing read-triggered `ExpireMarketListings` behavior.
- No market audit UI or admin inspection surface.
- No extraction of unrelated stash, session, replay, or character progression repository methods.

## Acceptance Criteria

- `server/internal/store/repos.go` no longer contains the public market listing/offer/expiration/audit
  method implementations that currently span the v140-reviewed market block.
- The moved code lives in `server/internal/store/market_repo.go` and
  `server/internal/store/market_helpers.go` alongside `market_purchase.go`, remains in package
  `store`, and preserves the same method names and receiver signatures.
- Shared helpers that are already used by market purchase code may move with the market repository
  code, while genuinely generic helpers such as `rowScanner` stay available to the package.
- Existing store and HTTP market tests pass without acceptance-test rewrites.
- The maintainability baseline for `server/internal/store/repos.go` is lowered to the new line count.
- `docs/CODEMAP.md` points the Market domain at `market_repo.go` and `market_purchase.go`.
- The approved ratchet exception-cap policy is recorded in `CLAUDE.md`.

## Scope And Likely Files

- `server/internal/store/repos.go` - remove market-specific repository methods and helpers.
- `server/internal/store/market_repo.go` - new focused market repository file.
- `server/internal/store/market_helpers.go` - package-private market helper file.
- `server/internal/store/market_purchase.go` - compile against moved shared market helpers.
- `.maintainability/file-size-baseline.tsv` - lower `repos.go` baseline.
- `docs/CODEMAP.md` - update Market store row.
- `CLAUDE.md` - approved pre-task policy edit.
- `docs/plans/v141_2026-06-13-market-store-extraction.md` and `docs/as-built/v141_market-store-extraction.md`.

## Test And Bot Proof

This slice is a store-internal refactor with no gameplay/protocol change, so no new bot scenario is
required. Verification:

- `cd server && go test ./internal/store ./internal/http`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking product or design questions.
- Risk: accidentally moving a helper still used outside market. Mitigation: compile the full server
  packages and keep package-private helpers in package `store`.
