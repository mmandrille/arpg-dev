# v257 Spec - Full-Screen Map Overlay

Status: Complete
Date: 2026-06-18
Codename: full-screen-map-overlay

## Purpose

Extend the v256 discovery minimap into a larger full-screen map overlay so players can inspect the
known area without squinting at the compact HUD widget. Repeated `TAB` presses cycle through hidden,
compact HUD minimap, and full-screen overlay modes.

## Non-goals

- No server, shared protocol, replay, persistence, or database change.
- No reconnect/resume or cross-session explored-map memory.
- No new markers beyond the existing player marker and elite-objective pin.
- No route line, quest pathing, marker search, click-to-navigate, or map panning.
- No imported map art, shader plugin, Godot addon, or asset pipeline change.

## Acceptance Criteria

- `TAB` cycles map modes in order: hidden -> compact HUD -> full-screen overlay -> hidden.
- The compact HUD mode preserves the v256 size, transparency, player marker, known walls, explored
  cells, and optional elite-objective pin.
- The full-screen mode centers a much larger map overlay and displays the same explored cells,
  known walls, player marker, and optional elite-objective pin from the existing discovery state.
- The full-screen mode shows a wider world radius than the compact HUD map so it is useful for
  inspection rather than only magnification.
- Debug state exposes the current display mode, full-screen flag, current map size, and explored
  map counts for unit and client bot proof.
- Existing fog overlay and server-authoritative visibility behavior remain unchanged.

## Scope and Likely Files

- Client presentation:
  - `client/scripts/discovery_minimap.gd` - add display modes, larger centered layout, current-size
    drawing, and debug mode fields.
  - `client/scripts/main.gd` - switch the `TAB` handler from binary toggle to mode cycling.
- Client tests and bot:
  - `client/tests/test_discovery_minimap.gd` - cover the three-mode cycle and full-screen sizing.
  - `client/scripts/bot_assertion_handlers.gd` - allow display-mode assertions.
  - `tools/bot/scenarios/client/70_full_screen_map_overlay.json` - client visual proof for compact
    and full-screen map modes.
- Docs:
  - `docs/plans/v257_2026-06-18-full-screen-map-overlay.md`
  - `docs/as-built/v257_full-screen-map-overlay.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, imported map art, shader plugins, and Godot addons.
Borrow the existing v256 `DiscoveryMinimap` drawing and debug-state patterns.

## Test and Bot Proof

- `make client-unit`
- `HEADLESS=1 make bot-visual scenario=70_full_screen_map_overlay`
- `make maintainability`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=70_full_screen_map_overlay
```

## Open Questions and Risks

- User answered on 2026-06-18 that `TAB` should cycle map styles.
- Risk: `client/scripts/main.gd` is a large grandfathered coordinator. Keep the change to the
  existing TAB handler only.
