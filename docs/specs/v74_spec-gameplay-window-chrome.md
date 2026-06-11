# v74 Spec — Gameplay Window Chrome

Status: Complete
Date: 2026-06-11
Codename: `gameplay-window-chrome`

## Purpose

Extend the v73 draggable titlebar shell to the gameplay item windows: inventory, shop, and stash. These windows should gain a darker titlebar, close `X`, titlebar-only dragging, and viewport clamping while preserving existing item drag/drop, buy/sell, stash, and keyboard open/close behavior.

## Non-goals

- No persistence of customized window positions.
- No migration of waypoint, settings, character select, multiplayer, pause, or main menu windows.
- No server, protocol, shared rules, or persistence change.

## Acceptance Criteria

- Inventory, shop, and stash panels render the reusable darker titlebar chrome.
- The `X` button hides the corresponding panel using the existing close semantics.
- Dragging works only from the titlebar and does not interfere with inventory/shop/stash slot drag/drop.
- Each panel clamps inside the viewport.
- Existing item interaction tests still pass.
- Debug state exposes window title, position, close availability, and dragging support for all three migrated panels.

## Scope And Likely Files

- `client/scripts/inventory_panel.gd`
- `client/scripts/shop_panel.gd`
- `client/scripts/stash_panel.gd`
- Focused existing panel tests under `client/tests/`
- `docs/as-built/`, `docs/plans/`, `PROGRESS.md`

## Test And Bot Proof

- `make client-unit` is the primary proof because this is client presentation and local UI interaction behavior.
- No protocol bot scenario is required; server-owned gameplay outcomes and wire contracts are unchanged.

## Open Questions And Risks

- Risk: inventory/shop/stash panels have existing drag/drop controls; titlebar drag must not capture body slot events.
- Risk: shop/stash titles are dynamic; titlebar labels must update when the panel title changes.

