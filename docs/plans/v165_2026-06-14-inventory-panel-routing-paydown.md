# v165 Plan — Inventory Panel Routing Paydown

Status: Complete
Goal: Move inventory action decision glue into `InventoryTransferRouter` while preserving behavior.
Architecture: This is a client-only maintainability slice. The server remains authoritative; the
router only decides which existing intent payload the panel should emit.
Tech stack: Godot GDScript client and headless client unit tests.

## Baseline and shortcut decision

Builds on v163 `inventory-transfer-router` and v164 `session-browser-filters`. Godot plugin
adoption: reject, because this is internal GDScript routing extraction with no new UI/art/camera.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/inventory_transfer_router.gd` | Add focused helpers for action/equip/hotbar decision payloads |
| Modify | `client/scripts/inventory_panel.gd` | Delegate extracted routing decisions |
| Modify | `client/tests/test_inventory_transfer_router.gd` | Pin extracted decision behavior |
| Modify | `client/tests/test_stash_panel.gd` | Keep panel integration coverage on router-owned slot parsing |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower panel baseline if line count drops |
| Modify | `PROGRESS.md` | Slice lifecycle and summary |
| Add | `docs/as-built/v165_inventory-panel-routing-paydown.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/inventory_panel.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused router helpers as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 — Router extraction

Files:
- Modify: `client/scripts/inventory_transfer_router.gd`
- Modify: `client/scripts/inventory_panel.gd`

- [x] Step 1.1: Move payload construction for double-click, shift-click, and drop routing into
  focused router helpers where the panel currently assembles context and emits the decision.
- [x] Step 1.2: Keep the panel responsible for UI state lookups, eligibility checks, and signal
  emission only.
```bash
make client-unit
```

## Task 2 — Router unit coverage

Files:
- Modify: `client/tests/test_inventory_transfer_router.gd`
- Modify: `client/tests/test_stash_panel.gd`

- [x] Step 2.1: Add focused assertions for extracted priority order and weapon-set payloads.
- [x] Step 2.2: Verify existing shop/stash/corpse/unique-chest/blacksmith routes still match.
```bash
make client-unit
```

## Task 3 — Lifecycle docs and CI

Files:
- Modify: `.maintainability/file-size-baseline.tsv`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v165_inventory-panel-routing-paydown.md`

- [x] Step 3.1: Lower the inventory panel baseline if the final file is smaller.
- [x] Step 3.2: Update lifecycle docs and write the as-built note.
- [x] Step 3.3: Run final verification.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make ci`
