# v73 Spec — Draggable Window Foundation

Status: Complete
Date: 2026-06-11
Codename: `draggable-window-foundation`

## Purpose

Add a reusable client-side window chrome pattern with a darker titlebar, close button, and titlebar-only dragging. Prove the behavior on the character stats and skills panels before migrating the larger inventory/shop/stash surfaces.

## Non-goals

- No persistence of custom positions; positions reset to defaults on client restart.
- No migration of inventory, shop, stash, waypoint, settings, character select, or multiplayer panels in this slice.
- No protocol, server, shared rule, or persistence change.

## Acceptance Criteria

- Character stats and skills panels render a darker titlebar above their content.
- Each migrated titlebar has a right-aligned `X` button that hides the panel without triggering gameplay input behind it.
- Dragging starts only from the titlebar, not from panel body controls.
- Dragged panels clamp inside the viewport so the titlebar remains reachable.
- Existing keyboard toggles still open/close stats and skills panels.
- Debug state exposes enough window data for tests: title, position, close availability, and dragging support.

## Scope And Likely Files

- `client/scripts/draggable_window.gd` — new reusable chrome helper/control.
- `client/scripts/character_stats_panel.gd` — migrate stats panel to window chrome.
- `client/scripts/skills_panel.gd` — migrate skills panel to window chrome.
- `client/tests/test_skills_panel.gd` and/or `client/tests/test_coop_client.gd` — focused coverage for titlebar, close, drag, and clamping.
- `docs/plans/`, `docs/as-built/`, `PROGRESS.md` — lifecycle docs.

## Test And Bot Proof

- `make client-unit` is the primary proof for GDScript panel behavior.
- No protocol bot scenario is required because this slice is client presentation only and does not change gameplay, protocol, server authority, or replay.
- Manual follow-up check: `make play`, open stats/skills, drag each titlebar, close each with `X`.

## Open Questions And Risks

- Risk: current tests assert fixed stats/skills relative positions; keep defaults unchanged and only test movement when explicitly invoked.
- Risk: panel body mouse handling can conflict with drag if the helper captures too broadly; drag must be titlebar-only.
