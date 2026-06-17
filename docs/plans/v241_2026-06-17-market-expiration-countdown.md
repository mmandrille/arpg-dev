# v241 Plan - Market Expiration Countdown

Status: Ready for implementation
Goal: Show expiration timing on market listing rows without changing market ownership semantics.
Architecture: The store already owns `MarketListing.ExpiresAt`; the HTTP layer will expose it as
RFC3339 text and the Godot client will render display-only expiration labels. Expiration remains
server-enforced by the existing read/sweep paths. Client bot proof reads row debug state rather than
depending on pixel inspection.
Tech stack: Go HTTP/store response serialization, Godot market UI, client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v240. Market listings already have `expires_at` in `server/internal/store/models.go` and
active reads already filter/sweep expired rows. Asset/plugin decision: reject external assets and
plugins; borrow the existing market row panel and text styling.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/http/market.go` | Include `expires_at` in market listing responses. |
| Modify | `client/scripts/market_panel.gd` | Render expiration labels and delegate extracted listing helpers. |
| Create | `client/scripts/market_listing_rows.gd` | Focused listing title/detail/expiration/debug helpers. |
| Create | `client/tests/test_market_listing_rows.gd` | Cover expiration label and debug fields. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Match market listing rows by expiration-label presence/text. |
| Modify | `client/scripts/bot_step_catalog.gd` | Validate expiration expectations for market panel steps. |
| Create | `tools/bot/scenarios/client/58_market_expiration_countdown.json` | Client-bot proof for listing expiration row. |
| Modify | `docs/specs/v241_spec-market-expiration-countdown.md` | Mark complete if implementation differs. |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Slice lifecycle close-out. |
| Create | `docs/as-built/v241_market-expiration-countdown.md` | As-built proof summary. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/market_panel.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline or avoid growth through extraction?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:

```bash
make maintainability
```

## Task 1 - Server response field

Files:
- Modify: `server/internal/http/market.go`

- [x] Step 1.1: Add `expires_at` to `marketListingResponse` and populate it from `MarketListing.ExpiresAt`.

```bash
cd server && go test ./internal/http -run Market -count=1
```

## Task 2 - Client listing expiration display

Files:
- Modify: `client/scripts/market_panel.gd`
- Create: `client/scripts/market_listing_rows.gd`
- Create: `client/tests/test_market_listing_rows.gd`

- [x] Step 2.1: Extract listing row helper logic from `market_panel.gd` into `market_listing_rows.gd`.
- [x] Step 2.2: Render an expiration line for listings with a valid `expires_at`; fall back gracefully for missing/malformed values.
- [x] Step 2.3: Expose `expires_at`, `expiration_label`, and `expiration_visible` in listing debug rows.
- [x] Step 2.4: Add unit coverage for label formatting and debug payloads.

```bash
godot --headless --path client --script res://tests/test_market_listing_rows.gd
godot --headless --path client --script res://tests/test_shop_panel.gd
```

## Task 3 - Client bot proof

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Create: `tools/bot/scenarios/client/58_market_expiration_countdown.json`

- [x] Step 3.1: Add market listing row matchers for `expiration_visible` and `expiration_contains`.
- [x] Step 3.2: Add a focused client bot scenario that opens a preflight listing and asserts expiration text.

```bash
make bot-client scenario=58_market_expiration_countdown.json HEADLESS=1
```

## Task 4 - Lifecycle docs and close-out

Files:
- Modify: `docs/specs/v241_spec-market-expiration-countdown.md`
- Modify: `docs/plans/v241_2026-06-17-market-expiration-countdown.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Create: `docs/as-built/v241_market-expiration-countdown.md`

- [x] Step 4.1: Mark spec/plan complete and add lifecycle/as-built documentation.
- [x] Step 4.2: Update `PROGRESS.md` current status to v241 with final batch CI pending.

```bash
make maintainability
```

## Final verification

- [x] `cd server && go test ./internal/http -run Market -count=1`
- [x] `godot --headless --path client --script res://tests/test_market_listing_rows.gd`
- [x] `godot --headless --path client --script res://tests/test_shop_panel.gd`
- [x] `make bot-client scenario=58_market_expiration_countdown.json HEADLESS=1`
- [x] `make maintainability`
- [x] Batch-level `make ci` deferred to `$autoloop` after the selected queue completes.
