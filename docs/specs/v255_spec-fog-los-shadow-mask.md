# v255 Spec - Fog LOS Shadow Mask

Status: Complete
Date: 2026-06-17
Codename: fog-los-shadow-mask

## Purpose

Make the Godot fog-of-war presentation respect the hero's line of sight around rectangular walls.
v253 hides living monsters behind walls on the server, but the client still lights floor space with
a radial circle, so the far side of a wall can look visible even though the hero's eyes cannot see
it. This slice adds client-side visual LOS shadows from the current authoritative wall layout: walls
inside the light/gloom area cast opaque fog over the region behind them, while the blocking wall
itself remains readable.

## Non-goals

- No server gameplay visibility, combat, aggro, monster AI awareness, or protocol changes.
- No durable explored-map memory, minimap memory, or session-persistent reveal.
- No doorway, high-obstacle, non-rectangular, destructible, secret, or vertical occluder semantics.
- No production lighting/art/audio pass, imported fog art, shader plugin, Godot addon, or asset
  pipeline change.
- No attempt to hide wall layouts themselves; authoritative wall geometry remains visible.

## Acceptance Criteria

- The local hero's fog overlay receives the current wall layout and hero/camera state without a
  protocol or schema bump.
- Rectangular walls inside the light/gloom area generate visual LOS shadow polygons that extend
  away from the hero to the gloom boundary or viewport edge.
- The shadowed region behind a wall is rendered as opaque darkness even when it lies inside the
  radial light or gloom radius.
- The occluding wall remains visible/readable; the shadow starts from the far side/tangent edge of
  the wall rather than blanketing the wall mesh itself.
- Loot, interactables, monsters, projectiles, companions, and floor/ground behind the wall are
  visually obscured by the overlay, regardless of whether those entities still exist in client
  state.
- When no walls are present, existing radial light/gloom/darkness behavior remains unchanged.
- Multiple wall shadows can overlap without flicker or layout-dependent instability.
- Fog shadows update as the hero moves, the camera moves, or a snapshot/delta replaces the wall
  layout.
- Bot/debug state exposes enough data to prove LOS masking: wall occluder count, shadow polygon
  count, and at least representative polygon points or bounds.

## Scope and Likely Files

- Client presentation:
  - `client/scripts/fog_of_war_overlay.gd` - extend the overlay with wall-aware LOS mask geometry,
    shader inputs/debug state, and no-wall fallback behavior.
  - `client/scripts/main.gd` - pass `current_wall_layout` into the overlay after snapshots, wall
    layout deltas, world fallback rendering, and per-frame camera/hero updates.
  - `client/scripts/wall_renderer.gd` - reuse normalized wall views if needed; avoid coupling
    rendering nodes back into the overlay.
- Client tests and bot:
  - `client/tests/test_fog_of_war_overlay.gd` - add pure/debug coverage for no-wall fallback,
    one-wall shadow generation, wall visibility boundary, and overlap count behavior.
  - `client/scripts/bot_assertion_handlers.gd` and related bot validation if the existing
    `assert_fog_of_war` step needs LOS-shadow expectations.
  - `tools/bot/scenarios/client/68_fog_los_shadow_mask.json` - visual proof scenario using an
    existing wall lab or compact world with a wall between the hero and open floor.
- Docs:
  - `docs/plans/v255_2026-06-17-fog-los-shadow-mask.md`
  - `docs/as-built/v255_fog-los-shadow-mask.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, imported fog art, shader plugins, and Godot addons.
Borrow existing in-repo `FogOfWarOverlay`, `WallRenderer`, `input_shadow_overlay.gd`/CanvasLayer
overlay patterns, current wall-layout debug state, and client bot assertion conventions.

## Test and Bot Proof

- `make client-unit`
- `HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask`
- `HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay`
- `make maintainability`

Manual visual proof, if desired after implementation:

```bash
make bot-visual scenario=68_fog_los_shadow_mask
```

## Open Questions and Risks

- No required questions. Defaults accepted on 2026-06-17: LOS shadows are opaque, the blocking wall
  remains visible, and all behind-wall objects are visually obscured.
- Risk: screen-space polygon math can produce artifacts near camera edges or when a wall crosses
  the hero position. The implementation should clamp/extents defensively and expose debug geometry
  so bot/unit tests can catch empty or inverted polygons.
- Risk: `client/scripts/main.gd` is a large coordinator. The plan should keep changes narrow and
  consider a focused helper if wall-to-overlay synchronization would otherwise grow the file.
- Risk: this is presentation-only. Server-authoritative monster visibility remains the source of
  truth for creature state; the client mask must not introduce gameplay decisions.
