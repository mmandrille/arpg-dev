# Inventory Market Blacksmith Flow - Implementation Plan

Goal: Let market and blacksmith actions stage character inventory items directly.
Architecture: The server remains authoritative. Client panels only stage inventory rows and send
HTTP requests containing character-owned item instance ids; the server reserves those items into
account stash before invoking existing market/upgrade store operations. Blacksmith upgrade payment
uses character gold first, then account stash gold for the remainder.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/http/market.go` | Accept inventory-origin publish and offer bodies |
| Modify | `server/internal/http/account_stash.go` | Add inventory-origin upgrade route |
| Modify | `server/internal/store/repos.go` | Spend upgrade cost from character gold before stash gold |
| Modify | `server/internal/http/auth_session_test.go` | Cover inventory-origin HTTP flows |
| Modify | `client/scripts/net_client.gd` | Store selected character id and send new request bodies |
| Modify | `client/scripts/inventory_panel.gd` | Expose market/blacksmith double-click contexts |
| Modify | `client/scripts/market_panel.gd` | Stage publish item and 2x5 offer grid from inventory |
| Modify | `client/scripts/blacksmith_panel.gd` | Stage one inventory item and show upgrade preview |
| Modify | `client/scripts/main.gd` | Open inventory with services and route staged actions |
| Modify | `client/tests/test_shop_panel.gd` | Add focused panel regression coverage |

## Tasks

- [x] Record plugin decision and contract shape in spec.
- [x] Add server acceptance tests for inventory-origin market listing, offer, and upgrade.
- [x] Implement HTTP request handling without removing existing stash fields.
- [x] Update NetClient and main coordinator for inventory-origin requests.
- [x] Update InventoryPanel contexts for service double-click routing.
- [x] Rework MarketPanel staging UI for publish and 2x5 offers.
- [x] Rework BlacksmithPanel staging UI and preview.
- [x] Apply blacksmith payment from inventory gold first, then stash gold.
- [x] Add client unit coverage.
- [x] Run targeted verification.

## Maintainability Exception

This slice crosses existing large UI surfaces (`main.gd`, `inventory_panel.gd`, `market_panel.gd`,
`blacksmith_panel.gd`), the broad authenticated HTTP route test, and the existing store repository
transaction file. Splitting these files cleanly would be a separate extraction slice because the
service panels still rely on existing debug-state, bot, and draggable-window patterns, while the
wallet-aware upgrade transaction belongs beside the existing stash upgrade code for this slice. The
file-size baseline is intentionally updated for the touched files, plus the currently drifting
`stash_panel.gd` baseline reported by the ratchet, and a follow-up should extract shared
market/blacksmith inventory staging widgets before adding more town-service UI.

## Final verification

- [x] `cd server && go test ./internal/http`
- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make bot-client scenario=35_market_board_ui`
- [x] `make bot-client scenario=39_blacksmith_upgrade_ui`
