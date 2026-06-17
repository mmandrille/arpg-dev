# v231 Plan - Market Cancel Offer UI

Status: Complete
Goal: Add seller-facing market listing cancellation to the Godot market board.
Architecture: Reuse the existing authenticated cancel-listing HTTP route and refresh the current
market panel after success; keep all ownership and refund semantics in the Go store.
Tech stack: Godot client scripts, existing HTTP market API, client bot scenario, SDD docs.

## Baseline and shortcut decision

Do not change backend market semantics. The route, store refund behavior, and tests already exist.
This slice only wires a player-facing action and bot proof. Asset/plugin decision: adopt the
existing market UI and bot/preflight framework; reject external assets/plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/net_client.gd` | Add `cancel_market_listing` transport helper |
| Modify | `client/scripts/market_panel.gd` | Add seller-owned cancel action and bot hook |
| Modify | `client/scripts/main.gd` | Route cancel action, status, refresh, stash upsert |
| Modify | `client/scripts/bot_controller.gd` | Add bot action dispatch and description |
| Modify | `client/scripts/bot_scenario_runner.gd` | Allow cancel action in market scenario catalog |
| Add | `tools/bot/scenarios/client/48_market_cancel_listing_ui.json` | Prove listing cancel UI |
| Modify | `PROGRESS.md` | Current status after completion |
| Modify | `docs/progress/slice-lifecycle.md` | Lifecycle row |
| Add | `docs/as-built/v231_market-cancel-offer-ui.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` is currently at its allowed ceiling; offset any added lines.
- [x] `client/scripts/market_panel.gd` has limited +25 growth budget.
- [x] `client/scripts/bot_controller.gd` has limited growth budget; offset if needed.
- [x] `client/scripts/bot_scenario_runner.gd` has limited growth budget; keep to one-line catalog change.
- [x] Did every touched grandfathered file stay within the ratchet?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice: not expected because this is a
  narrow UI wiring change.
- [x] Defer extraction with rationale: backend market extraction already exists; client hotspot
  growth will be offset locally.

Verification:
```bash
make maintainability
```

## Task 1 - Client cancel transport and routing

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Add `cancel_market_listing(listing_id)`.
- [x] Step 1.2: Route `market_action_requested("cancel_listing")` through the HTTP helper, refresh
  inventory/market state, and show success/failure status.

## Task 2 - Market panel action and bot hook

Files:
- Modify: `client/scripts/market_panel.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`

- [x] Step 2.1: Add a seller-owned row cancel button next to `View Offers`.
- [x] Step 2.2: Add a bot method/action to click cancel for a matching seller-owned listing.
- [x] Step 2.3: Keep debug state assertions driven by existing owned listing rows/status.

## Task 3 - Scenario and lifecycle proof

Files:
- Add: `tools/bot/scenarios/client/48_market_cancel_listing_ui.json`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v231_market-cancel-offer-ui.md`

- [x] Step 3.1: Add a scenario that preflights a seller listing with a bidder offer, cancels it,
  asserts the seller row disappears, and asserts the listed item returns to stash.
- [x] Step 3.2: Record v231 as complete with focused proof and note final batch CI is pending.

## Final verification

- [x] `make client-unit`
- [x] `make bot-client scenario=48_market_cancel_listing_ui.json HEADLESS=1`
- [x] `make maintainability`
- [x] Batch-level `make ci` is deferred to `$autoloop` after the selected queue commits.
