# v288 Plan — Market Notification Badge

Status: Complete
Goal: Make market-board badge counters reflect account market summary counts and prove them before
the market panel opens.
Architecture: Keep server/store behavior unchanged. Move badge presentation into a focused client
helper and expose read-only debug state for bot assertions.
Tech stack: Godot client helper/unit test, existing market summary HTTP route, client bot scenario.

## Baseline and shortcut decision

The market board already contains `IncomingBidBadge` and `PublishedListingBadge` nodes, and
`main.gd` already fetches `/v0/market/summary`. This slice should finish that surface rather than
adding a new notification system.

Asset/plugin decision:

- Adopt: existing badge geometry, existing summary route, existing market preflight helper.
- Borrow: existing market client bot scenario structure.
- Reject: external assets/plugins, realtime notification transport, unread persistence, polling, and
  market-panel redesign.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/market_board_badges.gd` | Apply and read market board badge counts/visual state. |
| Modify | `client/scripts/main.gd` | Delegate badge application and expose `market_board_badges` bot state. |
| Modify | `client/scripts/town_node_factory.gd` | Start badge nodes hidden until nonzero counts arrive. |
| Create | `client/tests/test_market_board_badges.gd` | Headless helper proof for hidden/active badge states. |
| Modify | `client/scripts/bot_step_catalog.gd` | Register market-badge wait/assertion fields. |
| Modify | `client/scripts/bot_wait_handlers.gd` | Evaluate wait steps against badge debug state. |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Evaluate direct badge assertions. |
| Create | `client/scripts/bot_market_badge_assertions.gd` | Focused matcher for badge debug state. |
| Create | `tools/bot/scenarios/client/68_market_notification_badge.json` | Client bot proof. |
| Create during finish | `docs/as-built/v288_market-notification-badge.md` | Record proof and commands. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [x] `client/scripts/main.gd` — only replace local badge logic with helper delegation and expose
  debug state.

Decision:

- [x] Extract helper/module: `market_board_badges.gd` owns badge formatting, colors, visibility, and
  debug extraction.
- [x] Defer extraction with rationale: no broader market-panel or town-node refactor in this slice.

Verification:

```bash
make maintainability
```

## Task 1 — Badge helper and node defaults

Files:

- Create: `client/scripts/market_board_badges.gd`
- Modify: `client/scripts/town_node_factory.gd`
- Modify: `client/tests/test_market_board_badges.gd`

- [x] Step 1.1: Add helper constants for inactive, incoming-active, and listing-active colors.
- [x] Step 1.2: Add `apply_to_board(node, incoming_bids, published_listings)` that sets label text,
  active color, and badge visibility.
- [x] Step 1.3: Add `debug_state(node)` exposing existence, counts, text, visibility, and colors.
- [x] Step 1.4: Start badge root nodes hidden until counts are applied.
- [x] Step 1.5: Add a headless test for zero hidden state and nonzero active state.

Verify:

```bash
godot --headless --path client --script res://tests/test_market_board_badges.gd
```

## Task 2 — Main delegation and bot state

Files:

- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Preload `MarketBoardBadges`.
- [x] Step 2.2: Replace direct label mutation in `_update_market_board_badges` with helper calls.
- [x] Step 2.3: Add `_market_board_badge_debug_state` and expose it as `market_board_badges` in
  `get_bot_state`.

Verify:

```bash
godot --headless --path client --script res://tests/test_item_visuals.gd
```

## Task 3 — Bot assertion and scenario

Files:

- Create: `client/scripts/bot_market_badge_assertions.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_wait_handlers.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Create: `tools/bot/scenarios/client/68_market_notification_badge.json`

- [x] Step 3.1: Add `wait_market_board_badges` and `assert_market_board_badges` support.
- [x] Step 3.2: Match expected incoming/published counts and optional visibility/text/color fields.
- [x] Step 3.3: Add a client bot scenario using market preflight as the seller with a bidder offer.
- [x] Step 3.4: Assert badge counts before opening the market panel, then open the panel to preserve
  the workflow proof.

Verify:

```bash
make bot-client scenario=68_market_notification_badge HEADLESS=1
```

## Task 4 — Docs and lifecycle

Files:

- Existing: `docs/specs/v288_spec-market-notification-badge.md`
- Existing: `docs/plans/v288_2026-06-19-market-notification-badge.md`
- Create during finish: `docs/as-built/v288_market-notification-badge.md`
- Modify during finish: `PROGRESS.md`

- [x] Step 4.1: Record focused checks and bot proof in the as-built note.
- [x] Step 4.2: Update lifecycle/current status during finish.

## Task 5 — Final verification

- [x] `godot --headless --path client --script res://tests/test_market_board_badges.gd`
- [x] `godot --headless --path client --script res://tests/test_item_visuals.gd`
- [x] `make bot-client scenario=68_market_notification_badge HEADLESS=1`
- [x] `make maintainability`
