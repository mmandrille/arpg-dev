# v165 Spec — Inventory Panel Routing Paydown

Status: Complete
Date: 2026-06-14
Codename: `inventory-panel-routing-paydown`

## Purpose

Continue the v163 inventory routing extraction by moving the remaining inventory action decision
glue out of `client/scripts/inventory_panel.gd` and into focused router helpers. The player-facing
inventory behavior must stay unchanged: double-click, shift-click, drag/drop, equip, unequip, use,
shop, stash, corpse, unique chest, market, blacksmith, and hotbar actions keep producing the same
client intent payloads.

## Non-goals

- No new inventory features, server intents, protocol schema changes, or shared-rule changes.
- No redesign of the inventory panel visuals or item tooltip presentation.
- No changes to authoritative server inventory behavior.

## Acceptance Criteria

- `inventory_panel.gd` delegates routing payload construction to `InventoryTransferRouter` and is
  smaller than its current maintainability baseline.
- `InventoryTransferRouter` has direct unit coverage for the extracted equip/use/hotbar/drop
  decision paths, including weapon-set payload preservation.
- Existing inventory transfer, stash, shop, market, blacksmith, corpse, unique chest, and hotbar
  client behavior remains compatible with current client tests.
- `make client-unit`, `make maintainability`, and `make ci` pass.

## Scope and Likely Files

- Client: `client/scripts/inventory_panel.gd`, `client/scripts/inventory_transfer_router.gd`
- Client tests: `client/tests/test_inventory_transfer_router.gd`, `client/tests/test_stash_panel.gd`
- Docs: `PROGRESS.md`, `docs/as-built/v165_inventory-panel-routing-paydown.md`
- Maintainability: `.maintainability/file-size-baseline.tsv` if the panel baseline can be lowered.

## Test and Bot Proof

This is a behavior-preserving client maintainability slice, so no new gameplay bot scenario is
required. Proof comes from focused router unit coverage plus the existing client unit and CI gates.

## Open Questions and Risks

- Risk: routing extraction can accidentally change priority order for contexts such as shop before
  equip or market before use. The router unit test must pin those priorities.
- Risk: `inventory_panel.gd` remains over 600 lines; the slice must still lower the grandfathered
  baseline rather than grow it.
