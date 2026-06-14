# v141 Plan: Market Store Extraction

Status: Ready for implementation
Goal: Move market store methods out of `repos.go` without changing behavior.
Architecture: This is a package-internal Go refactor. `*Store` keeps the same public method surface,
but listing, offer, expiration, audit, and summary implementations move to `market_repo.go`.
Purchase remains in `market_purchase.go` and continues using package-private market helpers.
Tech stack: Go store package, Postgres SQL, existing Go tests, lifecycle docs.

## Baseline and shortcut decision

Builds on v139 `market-expiration-read-freshness` and the v140 review recommendation to extract
market persistence before additional economy work. No Godot/client shortcut decision is required;
this slice does not touch client UI, camera, inventory presentation, or art. The approved pre-task
adds a ratchet exception-cap policy to `CLAUDE.md`.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `server/internal/store/market_repo.go` | Market listing/offer/expiration/audit/summary repository methods. |
| Create | `server/internal/store/market_helpers.go` | Market-only offer/refund/audit scanner helpers shared by market repo and purchase. |
| Modify | `server/internal/store/repos.go` | Remove moved market code while keeping generic store helpers available. |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower the `repos.go` grandfathered baseline after extraction. |
| Modify | `docs/CODEMAP.md` | Route Market store readers to the new focused file. |
| Modify | `CLAUDE.md` | Approved exception-cap policy pre-task. |
| Create | `docs/as-built/v141_market-store-extraction.md` | Close-out proof and deferred scope. |
| Modify | `PROGRESS.md` | Mark v141 complete and update current status. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/store/repos.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 - Extract market repository code

Files:
- Create: `server/internal/store/market_repo.go`
- Create: `server/internal/store/market_helpers.go`
- Modify: `server/internal/store/repos.go`

- [x] Step 1.1: Move public market methods from `repos.go` into `market_repo.go` with unchanged signatures.
- [x] Step 1.2: Move market-specific helpers needed by those methods and `market_purchase.go` into `market_helpers.go`.
- [x] Step 1.3: Keep generic helpers in `repos.go` when they are used outside the market domain.
```bash
gofmt -w server/internal/store/repos.go server/internal/store/market_repo.go server/internal/store/market_purchase.go
cd server && go test ./internal/store ./internal/http
```

## Task 2 - Update ratchet and CODEMAP

Files:
- Modify: `.maintainability/file-size-baseline.tsv`
- Modify: `docs/CODEMAP.md`
- Modify: `CLAUDE.md`

- [x] Step 2.1: Lower `server/internal/store/repos.go` baseline to the extracted file's current line count.
- [x] Step 2.2: Update the Market CODEMAP row to include `server/internal/store/market_repo.go`.
- [x] Step 2.3: Keep the approved CLAUDE.md exception-cap policy in the slice diff.
```bash
make maintainability
```

## Task 3 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v141_market-store-extraction.md`
- Modify: `docs/specs/v141_spec-market-store-extraction.md`
- Modify: `docs/plans/v141_2026-06-13-market-store-extraction.md`

- [x] Step 3.1: Mark plan checkboxes complete and update spec status.
- [x] Step 3.2: Add v141 lifecycle/as-built entries and current-status updates.
- [x] Step 3.3: Run final verification.
```bash
make ci
```

## Final verification

- [x] `cd server && go test ./internal/store -count=1`
- [x] `cd server && go test ./internal/http -run 'Market' -count=1`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Time-based background market sweeps remain deferred.
- Market audit browsing/admin UI remains deferred.
