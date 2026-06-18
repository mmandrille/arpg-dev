# v257 As-built - Full-Screen Map Overlay

Date: 2026-06-18

## What shipped

- Extended `DiscoveryMinimap` from a binary hidden/visible widget to three display modes:
  `hidden`, `compact`, and `fullscreen`.
- `TAB` now cycles `hidden -> compact -> fullscreen -> hidden` during gameplay.
- Compact mode keeps the v256 208x208 HUD map, top-right placement, transparency, explored cells,
  known walls, centered player marker, and optional elite-objective pin.
- Full-screen mode centers a 568x568 map overlay, uses the same session-local discovery state, and
  widens the world radius so it is useful for inspection rather than only magnification.
- Bot debug state now exposes `display_mode`, `full_screen`, and current map dimensions.
- Updated the old discovery minimap client scenario and added `70_full_screen_map_overlay`.

## Proof

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=70_full_screen_map_overlay
make maintainability
```

## Manual visual check

```bash
make bot-visual scenario=70_full_screen_map_overlay
```

## Scope limits

- No server, shared protocol, replay, persistence, database, or reconnect/resume behavior changed.
- No new map markers, route line, pathing, search, panning, or click-to-navigate shipped.
- No imported map art, shader plugin, Godot addon, or asset pipeline change shipped.
