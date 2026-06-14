# v141 As-built: Market Store Extraction

Spec: [`docs/specs/v141_spec-market-store-extraction.md`](../specs/v141_spec-market-store-extraction.md)
Plan: [`docs/plans/v141_2026-06-13-market-store-extraction.md`](../plans/v141_2026-06-13-market-store-extraction.md)

## What shipped

- Moved market listing, offer, expiration, audit, and summary store methods from
  `server/internal/store/repos.go` into `server/internal/store/market_repo.go`.
- Moved market-only offer/refund/audit/scanner helpers into
  `server/internal/store/market_helpers.go`, while leaving shared stash helpers in `repos.go`.
- Kept market purchase behavior in `server/internal/store/market_purchase.go` and sharing the same
  package-private market helper surface.
- Updated the Market CODEMAP row to point directly at the focused market store files.
- Lowered the `server/internal/store/repos.go` maintainability baseline from 3052 to 2315 lines.
- Recorded the approved ratchet exception-cap policy in `CLAUDE.md`.

## Proof

- `cd server && go test ./internal/store -count=1`
- `cd server && go test ./internal/http -run 'Market' -count=1`
- `.venv/bin/python tools/validate_codemap.py`
- `make maintainability`
- `make ci`

## Deferred

- Time-based background market sweeps remain deferred.
- Market audit browsing/admin UI remains deferred.
