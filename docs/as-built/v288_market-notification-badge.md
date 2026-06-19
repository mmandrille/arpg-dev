# v288 As Built: Market Notification Badge

Date: 2026-06-19
Spec: [`docs/specs/v288_spec-market-notification-badge.md`](../specs/v288_spec-market-notification-badge.md)
Plan: [`docs/plans/v288_2026-06-19-market-notification-badge.md`](../plans/v288_2026-06-19-market-notification-badge.md)

## What shipped

- Added `market_board_badges.gd` to own market-board badge formatting, count clamping, hidden-zero
  visibility, active colors, and read-only debug state.
- Market board badge nodes now start hidden until the existing market summary path applies counts.
- `main.gd` delegates badge updates to the helper and exposes `market_board_badges` in bot state.
- Added `wait_market_board_badges` and `assert_market_board_badges` client bot steps.
- Added client bot scenario `68_market_notification_badge`, which preflights a seller-owned
  listing with a bidder offer and proves the in-world board shows one incoming bid and one
  published listing before the market panel opens.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_market_board_badges.gd
godot --headless --path client --script res://tests/test_item_visuals.gd
make bot-client scenario=68_market_notification_badge HEADLESS=1
make maintainability
```

Result: green on 2026-06-19.

Full verification:

```bash
make ci
```

Result: deferred until the end of the selected autoloop queue.

## Manual visual command

```bash
make bot-visual scenario=68_market_notification_badge
```

## Deferred

- Periodic polling, realtime push, notification inboxes, unread persistence, and market-panel
  redesign remain deferred.
