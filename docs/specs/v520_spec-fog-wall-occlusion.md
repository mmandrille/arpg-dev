# v520 Spec - Fog Wall Occlusion

- **Status:** Complete — focused slice checks and combined batch CI passed; scenario 77's T-face readability limit is documented.
- **Date:** 2026-10-01
- **Codename:** fog-wall-occlusion

## Purpose

Close the visible light gaps where the client fog's independently projected shadow polygons meet
along straight walls and at connected wall corners or T-junctions. The wall presentation should
remain readable while the floor and objects behind the connected blocker stay dark.

The change is limited to the existing Godot fog presentation geometry and cache. The server's
authoritative LOS, fog radii, wall data, protocol, replay, aggro, and gameplay behavior remain
unchanged.

## Non-goals

- No authoritative LOS, visibility radius, monster awareness/aggro, combat, replay, or gameplay
  behavior changes.
- No protocol, schema, wall-layout format, shared tuning contract, or server changes.
- No new blocker types, non-rectangular generated geometry, vertical occlusion, or door-state
  behavior changes.
- No imported art, assets, add-ons, shader plugins, or new asset pipeline.
- No performance claim; this slice verifies bounded presentation geometry and captures only.

## Acceptance Criteria

- Adjacent and overlapping rectangular LOS blockers that form one wall run are treated as a
  connected silhouette for visual shadow construction.
- The visual shadow remains continuous behind a straight wall run, a right-angle corner, and a
  T-junction. Geometry fixtures prove each joint has no uncovered shadow wedge in the projected
  region immediately behind the join.
- Disconnected blockers remain independent so a distant wall does not create a shadow between
  unrelated structures.
- The region behind connected blockers remains dark; the near wall face remains visible/readable
  in the actual isometric camera captures.
- A low-alpha, deterministically irregular outer shadow rim softens the otherwise straight edge;
  the opaque/dark core remains tied to the blocker geometry.
- The rim does not fill large concave openings or bridge detached blocker groups.
- Closed doors continue to use the existing supplied-occluder path and their current shadow
  behavior is unchanged.
- Cache invalidation still rebuilds geometry on wall-layout updates; static frames still reuse the
  cached result.
- Existing bot visual scenarios `68_fog_los_shadow_mask` and
  `77_line_of_sight_blocker_shadow` pass and produce inspectable real-camera captures.
- No authoritative server, protocol, replay, radius, aggro, or gameplay files or semantics change.

## Scope and Likely Files

- Client geometry:
  - `client/scripts/hero_visibility_field.gd` - group connected rectangular blockers and build a
    shared visual shadow silhouette while retaining the current polygon projection path; derive a
    bounded organic soft-edge polygon around (not instead of) the hard core.
  - `client/scripts/fog_of_war_overlay.gd` and `client/scripts/fog_presentation_loader.gd` - render
    the subtle outer rim under the existing gloom/core layers using catalog-owned values.
  - `shared/assets/fog_presentation.v0.json` and
    `shared/assets/fog_presentation.v0.schema.json` - own and bound the soft-edge alpha, scale, and
    variation.
  - `client/scripts/fog_los_shadow_cache.gd` - preserve hard invalidation for layout changes and
    cache behavior for stable geometry.
- Tests:
  - `client/tests/test_fog_of_war_overlay.gd` and/or a focused geometry test - add deterministic
    straight-run, corner, T-junction, and disconnected-blocker fixtures.
  - `client/tests/test_fog_los_shadow_cache.gd` - verify layout invalidation still covers combined
    wall geometry if cache behavior changes.
- Existing bot scenarios:
  - `tools/bot/scenarios/client/68_fog_los_shadow_mask.json`
  - `tools/bot/scenarios/client/77_line_of_sight_blocker_shadow.json`
- Focused test-world framing:
  - `shared/rules/worlds.v0.json` - add a compact connected T-junction to the existing blocker lab
    fixture so the actual camera can capture the joint; preserve its existing column, rock, water,
    and hole behavior.
- Documentation:
  - `docs/plans/v520_2026-10-01-fog-wall-occlusion.md`
  - `docs/as-built/v520_fog-wall-occlusion.md`
  - `docs/progress/slice-lifecycle.md`
  - `docs/CODEMAP.md`

No new scenario is planned unless the existing scenarios cannot provide camera-visible coverage
of a corner or T-junction. No `main.gd`, protocol, or shared tuning edits are expected.

## Existing Implementation and Asset Decision

v255 introduced screen-space shadows per rectangular blocker; v262 routed supplied closed-door
occluders through that path; v264 added the organic fog edge while preserving those shadows. The
current project also has `FogLosShadowCache`, bot debug counts, and two established camera
scenarios. This slice reuses the project-native geometry, cache, shader, and bot harness.

- **Adopt:** existing Godot `HeroVisibilityField` geometry and fog shadow cache.
- **Borrow:** current CanvasLayer/Polygon2D rendering, debug state, and scenarios 68 and 77.
- **Reject:** external assets, add-ons, shader plugins, and external geometry libraries.

## Focused Verification

- Godot unit fixtures for straight, corner, T-junction, disconnected blockers, and door behavior.
- `make client-unit`
- `make bot-visual scenario=68_fog_los_shadow_mask`
- `make bot-visual scenario=77_line_of_sight_blocker_shadow`
- Inspect the scenario captures with the actual isometric camera; verify the joint stays dark and
  the wall remains readable.
- `make maintainability`

The worker does not run `make ci`, `make ci-full`, or `/finish`; the coordinator owns the combined
batch gate and closeout.

## Risks and Open Questions

- A connected concave wall cluster may have open space inside its convex silhouette. Any shared
  silhouette approximation must avoid producing broad unrelated shadows; captures must inspect
  both the blocker face and the immediate behind-wall region.
- Projection is height-aware at wall edges. Shared geometry must preserve the current wall-height
  projection and fog underlay/core treatment.
- There are no blocking product questions in the accepted slice brief.
