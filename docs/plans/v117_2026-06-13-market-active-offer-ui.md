# v117 Plan — Market Active Offer UI

Status: Complete
Goal: Let sellers inspect and accept active item offers from the Godot market board.
Architecture: Keep market ownership in the existing HTTP/store routes. Godot loads seller-owned
offers on demand, renders the returned rows, and posts accept intents through `NetClient`; it does
not mutate item ownership locally. The client bot preflight creates both sides of the market setup
through normal backend routes before the seller scenario starts.
Tech stack: Godot client, existing Go HTTP market routes, Python preflight tooling, Godot client bot,
docs.

## Baseline and shortcut decision

the right local surface.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/net_client.gd` | Add list/accept market offer HTTP helpers. |
| Modify | `client/scripts/market_panel.gd` | Render seller offer rows and accept action. |
| Modify | `client/scripts/main.gd` | Route market offer actions, refresh listings, and status. |
| Modify | `client/scripts/bot_controller.gd` | Add bot controls for seller offer inspection/acceptance. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Add assertions/steps for market offer rows. |
| Modify | `tools/bot/client_market_preflight.py` | Optionally create a foreign bidder and active offer. |
| Modify | `scripts/bot_client.sh` | Pass new preflight metadata/options through existing harness. |
| Create | `tools/bot/scenarios/client/38_market_active_offer_ui.json` | Seller UI acceptance proof. |
| Create | `docs/as-built/v117_market-active-offer-ui.md` | As-built proof summary. |
| Modify | `PROGRESS.md` | Mark v117 complete and carry deferred scope. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] `client/scripts/bot_controller.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: check before finish

Decision:
- [x] Documented maintenance exception: the slice extends existing market UI action plumbing and
  bot command dispatch in `main.gd`, `bot_controller.gd`, and `bot_scenario_runner.gd`. Those
  grandfathered baselines were updated for this focused registration work; `market_panel.gd` was
  kept under the 600-line target.

Verification:
```bash
make maintainability
```

## Task 1 — Client HTTP and Panel

Files:
- Modify: `client/scripts/net_client.gd`
- Modify: `client/scripts/market_panel.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Add authenticated list/accept market offer helpers.
- [x] Step 1.2: Add seller-owned listing offer inspection and active offer rows with Accept action.
- [x] Step 1.3: Refresh active listings and status after acceptance.
```bash
make client-unit
```

## Task 2 — Bot Preflight and Scenario

Files:
- Modify: `tools/bot/client_market_preflight.py`
- Modify: `scripts/bot_client.sh`
- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `tools/bot/scenarios/client/38_market_active_offer_ui.json`

- [x] Step 2.1: Extend market preflight to create an optional foreign active offer.
- [x] Step 2.2: Add bot controls/assertions for loading and accepting seller offers.
- [x] Step 2.3: Prove seller accepts a preflight offer and the active listing disappears.
```bash
make bot-client scenario=38_market_active_offer_ui
```

## Task 3 — Lifecycle Docs and CI

Files:
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v117_market-active-offer-ui.md`

- [x] Step 3.1: Record completed slice, test proof, and deferred expiration/audit/notifications.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make bot-client scenario=38_market_active_offer_ui`
- [x] `make maintainability`
- [x] `make ci`
