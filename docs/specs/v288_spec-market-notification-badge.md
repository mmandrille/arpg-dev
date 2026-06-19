# v288 Spec: Market Notification Badge

Status: Implemented
Date: 2026-06-19
Codename: `market-notification-badge`

## Purpose

Make the town market board advertise account-relevant market activity before the player opens the
panel. Existing market summary data already reports active published listings and incoming bids; this
slice turns the existing board badge nodes into readable notification badges and proves them through
the client bot.

## Non-goals

- Do not add a notification inbox, unread persistence, polling timers, or realtime push messages.
- Do not change market listing, offer, receipt, purchase, or expiration rules.
- Do not add new server routes or database columns.
- Do not redesign the market panel.

## Acceptance Criteria

- The market board hides both badge counters when the authenticated account has no active published
  listings or incoming bids.
- The incoming-bids badge shows the current active incoming bid count and uses the active alert color
  when the count is greater than zero.
- The published-listings badge shows the current active listing count and uses the active listing
  color when the count is greater than zero.
- The badge update path reuses `GET /v0/market/summary` and refreshes after market actions that
  already refresh the market panel.
- Bot debug state exposes market badge counts, visibility, text, and colors.
- A client bot scenario prepares a seller-owned listing with a bidder offer, enters town, and proves
  the market board badges show the expected counts before opening the market panel.

## Scope And Likely Files

- `client/scripts/market_board_badges.gd` owns badge apply/debug helpers.
- `client/scripts/main.gd` delegates badge updates and exposes badge debug state.
- `client/scripts/town_node_factory.gd` keeps constructing badge nodes but starts them hidden.
- `client/scripts/bot_*` files add a focused market-badge wait/assertion hook.
- `client/tests/test_market_board_badges.gd` proves helper behavior.
- `tools/bot/scenarios/client/68_market_notification_badge.json` proves the town-facing badge.

## Test And Bot Proof

Focused checks:

```bash
godot --headless --path client --script res://tests/test_market_board_badges.gd
make bot-client scenario=68_market_notification_badge HEADLESS=1
make maintainability
```

Visual verification command for humans/agents:

```bash
make bot-visual scenario=68_market_notification_badge
```

## Asset And Plugin Decision

- Adopt: existing code-native market board badge meshes and existing market summary route.
- Borrow: existing client bot market preflight setup for a seller listing plus bidder offer.
- Reject: external assets, plugins, new art pipelines, notification persistence, and websocket push.

## Outcome

- The market board now hides zero-count badges and shows incoming-bid and published-listing counts
  using the existing account market summary route.
- Bot debug state and a client bot scenario prove the seller sees one active incoming bid and one
  published listing on the in-world board before opening the panel.
- Periodic polling, realtime push, unread persistence, and a notification inbox remain future work.
