# v231 Spec: Market Cancel Offer UI

Status: Complete
Date: 2026-06-16
Codename: market-cancel-offer-ui

## Purpose

Let sellers cancel their own active market listing from the Godot market board, using the existing
server-owned cancel route so the listed item and any active offer items are refunded by the store.

## Baseline

Builds on v93 market offer ownership, v114 market board UI, v115 direct purchase UI, v117 seller
offer inspection/acceptance, v128/v139 expiration freshness, and v130 audit records. The HTTP and
store cancel path already exists: `POST /v0/market/listings/{listing_id}/cancel`.

ADR alignment:
- ADR-0011: canceled listings must release reserved items predictably and remain server-owned.
- ADR-0014 D12: market polish should keep advancing the player-facing endgame loop.

Asset/plugin decision: adopt the existing market panel, authenticated HTTP client, preflight helper,
and bot scenario framework; reject external assets/plugins.

## Non-goals

- No new store/HTTP cancel semantics, listing edit, confirmation modal, audit UI, search, sorting,
  pagination, taxes, expiration timers, or notification inbox.
- No buyer-side offer cancel UI; this slice is seller listing cancel only.
- No realtime stash reload beyond the returned canceled listing payload and existing market refresh.

## Acceptance Criteria

- `NetClient` exposes authenticated `POST /v0/market/listings/{listing_id}/cancel`.
- Seller-owned listing rows show a cancel action next to existing offer inspection.
- Cancel success removes the seller-owned listing from the market panel after refresh and shows a
  clear status message.
- Cancel failure leaves the panel open and reports a clear warning status.
- Client debug state exposes enough row/status data for bot assertions.
- A client bot scenario prepares a seller-owned listing with a bidder offer, opens the market board,
  cancels the listing, and verifies the listing row disappears and the seller stash receives the
  listed item back.

## Scope and Likely Files

- Client: `client/scripts/net_client.gd`, `client/scripts/market_panel.gd`,
  `client/scripts/main.gd`, `client/scripts/bot_controller.gd`,
  `client/scripts/bot_scenario_runner.gd`
- Bot scenario: `tools/bot/scenarios/client/48_market_cancel_listing_ui.json`
- Docs: `docs/plans/v231_2026-06-16-market-cancel-offer-ui.md`,
  `docs/as-built/v231_market-cancel-offer-ui.md`, `PROGRESS.md`,
  `docs/progress/slice-lifecycle.md`

## Test and Bot Proof

- `make client-unit`
- `make bot-client scenario=48_market_cancel_listing_ui.json HEADLESS=1`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=48_market_cancel_listing_ui.json
```

## Open Questions and Risks

- No blocking questions. The existing backend route already owns refunds and authorization.
- Main client script is at its maintainability growth ceiling, so implementation must offset any
  added action handling lines.
