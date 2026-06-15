# v163 Plan — Inventory transfer router

Status: Complete
Goal: Inventory transfer/staging routing leaves `inventory_panel.gd` and lives in a focused helper.
Architecture: `InventoryPanel` still owns rendering, item eligibility, and signal emission. A new
router returns inert routing dictionaries for existing intent names and payloads; the panel applies
those decisions. Server authority and protocol contracts are unchanged.
Tech stack: Godot GDScript, client unit tests, existing client bot scenarios.

## Baseline and shortcut decision

Builds on the v160 review recommendation to extract inventory transfer/staging routing. Godot
maintainability work.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/inventory_transfer_router.gd` | Pure route decisions for inventory transfer/staging |
| Modify | `client/scripts/inventory_panel.gd` | Delegate routing and emit returned decisions |
| Create | `client/tests/test_inventory_transfer_router.gd` | Focused route contract unit tests |
| Modify | `scripts/client_smoke.sh` | Include the new GDScript unit test |
| Create | `docs/as-built/v163_inventory-transfer-router.md` | As-built summary |
| Modify | `docs/specs/v163_spec-inventory-transfer-router.md` | Mark complete at closeout |
| Modify | `PROGRESS.md` | Lifecycle closeout and next slice |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/inventory_panel.gd`
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Move routing branches out of `inventory_panel.gd`; do not add new behavior.

Verification:

```bash
make maintainability
```

## Task 1 — Router extraction

Files:
- Create: `client/scripts/inventory_transfer_router.gd`
- Modify: `client/scripts/inventory_panel.gd`

- [x] Add router functions for double-click, shift-click, and drop decisions.
- [x] Preserve existing intent names and payload keys exactly.
- [x] Keep item eligibility and weapon-set slot payload calculation in `InventoryPanel`.
- [x] Keep blacksmith unstaging as a returned action applied by `InventoryPanel`.

## Task 2 — Unit coverage

Files:
- Create: `client/tests/test_inventory_transfer_router.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Cover shop sell/buy, market stage, blacksmith stage/unstage, stash withdraw/equip, corpse
  withdraw, unique chest take, equip, unequip, use, and hotbar assignment route outputs.
- [x] Keep existing stash/shop panel integration tests passing.

```bash
make client-unit
```

## Task 3 — Scenario proof

Files:
- Existing: client bot scenarios

- [x] Run focused scenarios that exercise stash, market staging, and blacksmith staging.

```bash
make bot-client scenario=23_account_stash_panel.json
make bot-client scenario=35_market_board_ui.json
make bot-client scenario=39_blacksmith_upgrade_ui.json
```

## Task 4 — Lifecycle docs and CI

Files:
- Create: `docs/as-built/v163_inventory-transfer-router.md`
- Modify: `docs/plans/v163_2026-06-14-inventory-transfer-router.md`
- Modify: `docs/specs/v163_spec-inventory-transfer-router.md`
- Modify: `PROGRESS.md`

- [x] Mark completed plan tasks.
- [x] Update spec status/as-built/progress.

```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=23_account_stash_panel.json`
- [x] `make bot-client scenario=35_market_board_ui.json`
- [x] `make bot-client scenario=39_blacksmith_upgrade_ui.json`
- [x] `make ci`
