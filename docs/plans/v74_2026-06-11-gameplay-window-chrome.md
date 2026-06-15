# v74 Plan — Gameplay Window Chrome

Status: Complete
Goal: Apply reusable draggable titlebar chrome to inventory, shop, and stash panels.
Architecture: Reuse the v73 `DraggableWindow` helper. Keep each panel's existing public API, default positioning, and body controls intact; only replace the outer panel container and move title text into the titlebar.
Tech stack: Godot GDScript client, client unit tests, lifecycle docs.

## Baseline And Shortcut Decision

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/inventory_panel.gd` | Add window chrome, close, drag debug |
| Modify | `client/scripts/shop_panel.gd` | Add window chrome, close, dynamic title sync |
| Modify | `client/scripts/stash_panel.gd` | Add window chrome, close, dynamic title sync |
| Modify | `client/tests/test_shop_panel.gd` | Verify shop window chrome |
| Modify | `client/tests/test_stash_panel.gd` | Verify stash window chrome |
| Modify | `client/scripts/smoke.gd` or existing inventory tests | Verify inventory window chrome |
| Add | `docs/as-built/v74_gameplay-window-chrome.md` | As-built proof |
| Modify | `PROGRESS.md` | Lifecycle update |

## Task 1 — Inventory Chrome

Files:
- Modify: `client/scripts/inventory_panel.gd`
- Modify: inventory-focused test coverage

- [x] Step 1.1: Replace the inventory outer container with `DraggableWindow`, keeping default position and drag/drop body controls.
- [x] Step 1.2: Add close/drag bot helpers and window debug state.
```bash
make client-unit
```

## Task 2 — Shop And Stash Chrome

Files:
- Modify: `client/scripts/shop_panel.gd`
- Modify: `client/scripts/stash_panel.gd`
- Modify: `client/tests/test_shop_panel.gd`
- Modify: `client/tests/test_stash_panel.gd`

- [x] Step 2.1: Replace shop/stash outer containers with `DraggableWindow`.
- [x] Step 2.2: Move dynamic shop/stash titles into the reusable titlebar and preserve status/action rows.
- [x] Step 2.3: Add focused tests for close, drag, and window debug state.
```bash
make client-unit
```

## Task 3 — Lifecycle Docs

Files:
- Add: `docs/as-built/v74_gameplay-window-chrome.md`
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v74_spec-gameplay-window-chrome.md`
- Modify: `docs/plans/v74_2026-06-11-gameplay-window-chrome.md`

- [x] Step 3.1: Mark spec/plan complete and add as-built summary.
- [x] Step 3.2: Update `PROGRESS.md` to v74 and keep persistence/remaining menu windows deferred to v75.
```bash
make client-unit
```

## Final Verification

- [x] `make client-unit`

`make ci` is intentionally deferred unless targeted client verification exposes integration risk.
