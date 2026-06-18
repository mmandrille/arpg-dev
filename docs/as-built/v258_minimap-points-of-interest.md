# v258 As-built - Minimap Points of Interest

Date: 2026-06-18

## What shipped

- Added discovered point-of-interest marker derivation to `DiscoveryMinimapState`.
- The discovery map now classifies known client interactables as:
  - stairs: `stairs_down`, `stairs_up`
  - waypoint: `teleporter`
  - service: town vendor, mystery seller, stash, bishop, market board, blacksmith, mercenary board,
    and unique chest
  - objective: the existing elite-objective pin
- Interactable markers are gated by explored cells, so far points of interest are not revealed just
  because the server sent non-creature entities.
- `DiscoveryMinimap` draws code-native marker shapes/colors in both compact and full-screen modes.
- Bot debug state exposes total marker count plus service, stairs, waypoint, and objective marker
  counts.
- Added `71_minimap_points_of_interest` as a focused client scenario on `vendor_lab`.

## Proof

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=71_minimap_points_of_interest
make maintainability
```

## Manual visual check

```bash
make bot-visual scenario=71_minimap_points_of_interest
```

## Scope limits

- No server, shared protocol, replay, persistence, database, or hidden-state behavior changed.
- No marker labels panel, filtering, search, route line, panning, click-to-navigate, or tooltip
  interaction shipped.
- No imported icon art, shader plugin, Godot addon, or asset pipeline change shipped.
