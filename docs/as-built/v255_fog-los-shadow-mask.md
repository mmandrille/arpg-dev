# v255 As-Built - Fog LOS Shadow Mask

Date: 2026-06-17

## What shipped

- Extended the Godot fog overlay with wall-layout input and screen-space LOS shadow polygons.
- Rectangular walls inside the hero's light/gloom area now cast opaque black visual shadows over
  the floor and objects behind them.
- The blocking wall geometry remains visible/readable because the mask starts beyond the wall
  silhouette rather than replacing wall rendering.
- Kept the change presentation-only: no server, protocol, shared rule, replay, combat, aggro, or
  monster AI behavior changed.
- Exposed fog debug state for wall count, occluder count, shadow count, and representative shadow
  polygon points/bounds.
- Extended `assert_fog_of_war` with optional wall/occluder/shadow count expectations.
- Added the `68_fog_los_shadow_mask` client bot scenario on `collision_lab` to prove a wall layout
  produces LOS shadows.
- Follow-up: tightened the shadow start offset and projected the near mask edge from the primitive
  wall mesh height so diagonal wall angles do not leave visible floor slivers behind obstacles.

## Proof

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
make maintainability
make ci
```

Manual visual proof, if desired:

```bash
make bot-visual scenario=68_fog_los_shadow_mask
```

## Scope limits

- No server gameplay visibility, combat, aggro, monster AI awareness, or protocol changes shipped.
- No durable explored-map memory, minimap memory, or session-persistent reveal shipped.
- No doorway, high-obstacle, non-rectangular, destructible, secret, or vertical occluder semantics
  shipped.
- No production fog art, production dungeon lighting, imported shader plugin, Godot addon, or asset
  pipeline change shipped.
