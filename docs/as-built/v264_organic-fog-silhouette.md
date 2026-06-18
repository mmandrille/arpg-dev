# v264 As-Built - Organic Fog Silhouette

Date: 2026-06-18
Spec: [`docs/specs/v264_spec-organic-fog-silhouette.md`](../specs/v264_spec-organic-fog-silhouette.md)
Plan: [`docs/plans/v264_2026-06-18-organic-fog-silhouette.md`](../plans/v264_2026-06-18-organic-fog-silhouette.md)

## Shipped Behavior

- `FogOfWarOverlay` now offsets the visual light/gloom/darkness thresholds with deterministic
  angular noise, breaking the prior perfectly circular fog silhouette.
- Debug radii remain unchanged, so the organic edge is presentation-only and does not alter
  server-authoritative visibility or gameplay.
- LOS shadow polygons from walls and closed-door occluders still render as opaque darkness above the
  organic fog mask.
- Fog debug state exposes `organic_edge_enabled`, `organic_edge_px`,
  `organic_edge_world_amplitude`, and `organic_edge_segments`.
- Existing fog client bot scenarios now assert that the organic edge is active while preserving
  light radius, gloom radius, and LOS shadow expectations.

## Boundaries

- No server gameplay visibility, monster awareness, aggro, combat, protocol, database, replay, or
  persistence behavior changed.
- No imported fog art, shader plugin, Godot addon, particle system, production lighting pass, or
  dungeon art treatment shipped.
- No change to v255/v262 wall/door LOS shadow semantics shipped.

## Verification

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
make maintainability
```

All focused commands passed on 2026-06-18. The selected autoloop batch `make ci` gate remains
pending.

Manual visual check:

```bash
make bot-visual scenario=67_fog_of_war_overlay
make play
```
