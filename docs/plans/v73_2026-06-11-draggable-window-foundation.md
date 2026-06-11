# v73 Plan — Draggable Window Foundation

Status: Complete
Goal: Add reusable draggable titlebar chrome and prove it on stats and skills panels.
Architecture: Keep this entirely in the Godot client. A small reusable chrome helper owns titlebar, close, drag, and viewport clamping while existing panels keep their content and visibility APIs. No server, protocol, shared rules, or persistence changes are needed.
Tech stack: Godot GDScript client, client unit tests, lifecycle docs.

## Baseline And Shortcut Decision

Builds on v72. Godot plugin adoption checklist: reject plugin adoption for this slice because draggable window chrome is a small in-repo UI behavior and existing panels are custom GDScript controls; borrowing an inventory/window addon would add more surface than this proof needs.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/draggable_window.gd` | Reusable titlebar, close, drag, clamp helper |
| Modify | `client/scripts/character_stats_panel.gd` | Use chrome for stats panel |
| Modify | `client/scripts/skills_panel.gd` | Use chrome for skills panel |
| Modify | `client/tests/test_skills_panel.gd` | Verify skills titlebar, close, drag, clamp |
| Modify | `client/tests/test_coop_client.gd` | Verify stats titlebar and close behavior |
| Add | `docs/as-built/v73_draggable-window-foundation.md` | As-built proof |
| Modify | `PROGRESS.md` | Lifecycle update |

## Task 1 — Reusable Chrome

Files:
- Create: `client/scripts/draggable_window.gd`

- [x] Step 1.1: Add a reusable `DraggableWindow` helper with titlebar, title label, `X` close button, drag handling, viewport clamping, and debug state.
```bash
make client-unit
```

## Task 2 — Stats And Skills Migration

Files:
- Modify: `client/scripts/character_stats_panel.gd`
- Modify: `client/scripts/skills_panel.gd`

- [x] Step 2.1: Wrap the existing stats panel content in the reusable chrome without changing its default position, size, or public visibility methods.
- [x] Step 2.2: Wrap the existing skills panel content in the reusable chrome without changing its default position, size, or public visibility methods.
- [x] Step 2.3: Keep panel body controls interactive and restrict dragging to the titlebar.
```bash
make client-unit
```

## Task 3 — Focused Client Tests

Files:
- Modify: `client/tests/test_skills_panel.gd`
- Modify: `client/tests/test_coop_client.gd`

- [x] Step 3.1: Add focused assertions for titlebar state, close button behavior, drag movement, and viewport clamp.
- [x] Step 3.2: Confirm existing stats/skills panel tests still pass with default positions.
```bash
make client-unit
```

## Task 4 — Lifecycle Docs

Files:
- Add: `docs/as-built/v73_draggable-window-foundation.md`
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v73_spec-draggable-window-foundation.md`
- Modify: `docs/plans/v73_2026-06-11-draggable-window-foundation.md`

- [x] Step 4.1: Mark spec and plan complete.
- [x] Step 4.2: Update `PROGRESS.md` to v73 and defer full panel migration/persistence to the next slices.
- [x] Step 4.3: Add as-built summary.
```bash
make client-unit
```

## Final Verification

- [x] `make client-unit`

`make ci` is intentionally reserved for final pre-commit proof only if targeted tests leave integration risk; this slice is client presentation only and `make client-unit` covers the touched behavior.
