# v75 Plan — Persistent Window Layout

Status: Complete
Goal: Persist draggable positions for the migrated gameplay windows in local client storage.
Architecture: Extend `DraggableWindow` with optional local layout persistence keyed per panel. Panels opt in by setting stable keys after their default positions are established. Persistence is local-only and disabled in unit-test mode to avoid contaminating developer/user layouts.
Tech stack: Godot GDScript client, client unit tests, lifecycle docs.

## Baseline And Shortcut Decision

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/draggable_window.gd` | Save/load window positions by key |
| Modify | stats/skills/inventory/shop/stash panel scripts | Set stable persistence keys |
| Modify | focused client tests | Verify persistence API without writing test layout files |
| Add | `docs/as-built/v75_persistent-window-layout.md` | As-built proof |
| Modify | `PROGRESS.md` | Lifecycle update |

## Task 1 — Persistence Helper

Files:
- Modify: `client/scripts/draggable_window.gd`

- [x] Step 1.1: Add optional layout key, save-on-drag, load-on-enable, and viewport clamp on loaded positions.
- [x] Step 1.2: Disable file persistence when `CLIENT_UNIT_ONLY=1`.
```bash
make client-unit
```

## Task 2 — Panel Opt-in

Files:
- Modify: `client/scripts/character_stats_panel.gd`
- Modify: `client/scripts/skills_panel.gd`
- Modify: `client/scripts/inventory_panel.gd`
- Modify: `client/scripts/shop_panel.gd`
- Modify: `client/scripts/stash_panel.gd`

- [x] Step 2.1: Assign stable layout keys after each panel computes its default position.
- [x] Step 2.2: Keep all existing debug, close, and drag helpers intact.
```bash
make client-unit
```

## Task 3 — Tests And Docs

Files:
- Modify: focused client tests
- Add: `docs/as-built/v75_persistent-window-layout.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Add a focused persistence helper test using an in-memory/test-disabled path.
- [x] Step 3.2: Mark spec/plan complete and update lifecycle docs.
```bash
make client-unit
```

## Final Verification

- [x] `make client-unit`

`make ci` is intentionally deferred unless targeted client verification exposes integration risk.
