# v201 Spec: Item Level Tooltip

Status: Complete - `make ci` green on 2026-06-15
Date: 2026-06-15
Codename: item-level-tooltip

## Purpose

Surface authoritative `item_level` in shared item tooltips so players can inspect progression state on rolled, unique, set, market, stash, shop, inventory, and blacksmith items without relying on server/debug payloads.

## Non-goals

- No protocol, schema, store, or server mutation changes.
- No item-level scaling, upgrade odds changes, or loot-band rebalance.
- No new art assets, plugins, or tooltip redesign.

## Acceptance Criteria

- Tooltips prefer an explicit `Item level N` footer when an item carries `item_level`.
- Character level requirements remain visible as requirements when `item_level` is present.
- Items without `item_level` keep the existing requirement-level footer behavior.
- Inventory/shop/market/stash/blacksmith tooltip paths continue using the shared `ItemTooltipPanel`.
- `make client-unit` and `make ci` pass.

## Scope and Files Likely Touched

- `client/scripts/item_tooltip_panel.gd`
- `client/tests/test_shop_panel.gd`
- `docs/plans/v201_2026-06-15-item-level-tooltip.md`
- `docs/as-built/v201_item-level-tooltip.md`
- `PROGRESS.md`
- `docs/progress/slice-lifecycle.md`

## Test and Bot Proof

- Focused Godot unit coverage in `client/tests/test_shop_panel.gd` validates item-level footer text and requirement visibility.
- No new protocol bot is required because this is a client presentation-only tooltip rendering change on existing payload fields.

## Open Questions and Risks

- Risk: tooltip footer previously used the same slot for requirement level. Mitigation: prefer `Item level N` only when the payload has explicit item-level data, and keep requirement level in the requirements block for those items.
- Asset/plugin decision: rejected. This slice reuses the existing shared Godot tooltip component and introduces no external assets or plugins.
