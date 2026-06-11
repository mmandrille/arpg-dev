# v75 Spec — Persistent Window Layout

Status: Complete
Date: 2026-06-11
Codename: `persistent-window-layout`

## Purpose

Persist custom positions for the draggable gameplay windows so players can customize their in-game UI layout and keep it across restarts.

## Non-goals

- No server/account persistence; layout is local to the client machine.
- No protocol or shared schema changes.
- No migration of non-draggable menu windows in this slice.
- No reset-layout UI yet.

## Acceptance Criteria

- Stats, skills, inventory, shop, and stash windows each have a stable local layout key.
- Dragging a persisted window saves its clamped position to local user storage.
- Reopening/recreating the window restores the saved position.
- Unit/smoke tests do not pollute the player’s local saved layout.
- Existing panel behavior and item interactions remain unchanged.

## Scope And Likely Files

- `client/scripts/draggable_window.gd`
- `client/scripts/character_stats_panel.gd`
- `client/scripts/skills_panel.gd`
- `client/scripts/inventory_panel.gd`
- `client/scripts/shop_panel.gd`
- `client/scripts/stash_panel.gd`
- Focused client tests
- Lifecycle docs

## Test And Bot Proof

- `make client-unit` is the primary proof.
- No protocol bot scenario is required because layout persistence is local client presentation state.

## Open Questions And Risks

- Risk: writing user files in tests can leak state; disable persistence under `CLIENT_UNIT_ONLY`.
- Risk: saved positions can become invalid after resolution changes; always clamp loaded positions to the current viewport.

