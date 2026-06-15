# v114 Plan — Market Board UI

Status: Complete
Goal: Prove the town market board can publish and browse priced listings in Godot.
Architecture: Keep market ownership and listing creation server-owned through existing HTTP routes.
The client sends a chosen `price_gold`, renders the returned/listed price, and exposes debug rows
for bot assertions.
Tech stack: Godot client, authenticated HTTP client helper, client bot, lifecycle docs.

## Baseline and shortcut decision

Builds on v68 market listing foundation, v93 offers, and v111 direct priced purchase backend.
## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/net_client.gd` | Send optional `price_gold` on listing create. |
| Modify | `client/scripts/market_panel.gd` | Add price input, display listing prices, expose debug rows. |
| Modify | `client/scripts/main.gd` | Pass publish price to HTTP and expose bot helpers if needed. |
| Modify | `client/scripts/bot_controller.gd` | Add market bot action dispatch if needed. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add wait/assert support for market panel rows. |
| Create | `tools/bot/scenarios/client/35_market_board_ui.json` | Client proof for publish + browse. |
| Modify | `PROGRESS.md` | Mark v114 complete during finish. |
| Create | `docs/as-built/v114_market-board-ui.md` | Record shipped behavior and proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`

Decision:
- [x] Keep over-limit edits surgical; market-specific UI code stays in `market_panel.gd`.
- [x] Documented maintenance exception: market bot actions and step validation must register in the
  existing grandfathered bot controller/runner until those registries are split out.
- [x] Run `make maintainability` before final CI.

Verification:
```bash
make maintainability
```

## Task 1 — Price-aware publish and browse rows

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/market_panel.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Send `price_gold` when publishing a market listing.
- [x] Step 1.2: Add a simple publish price control with default 25 gold.
- [x] Step 1.3: Render `price_gold` in browse listing rows.
```bash
make client-unit
```

## Task 2 — Bot assertions

Files:
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `tools/bot/scenarios/client/35_market_board_ui.json`

- [x] Step 2.1: Expose market listing rows/prices in debug state.
- [x] Step 2.2: Add wait/assert support for market panel visible listing counts/prices.
- [x] Step 2.3: Add a client scenario that collects loot, deposits to stash, opens market,
  publishes a priced listing, and asserts browse visibility.
```bash
make bot-client scenario=35_market_board_ui
```

## Task 3 — Lifecycle docs and CI

Files:
- Modify: `docs/plans/v114_2026-06-13-market-board-ui.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v114_market-board-ui.md`

- [x] Step 3.1: Mark plan tasks complete as they pass.
- [x] Step 3.2: Update `PROGRESS.md` latest slice, next slice, lifecycle row, and recently closed note.
- [x] Step 3.3: Add the v114 as-built note.
```bash
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make bot-client scenario=35_market_board_ui`
- [x] `make maintainability`
- [x] `make ci`
