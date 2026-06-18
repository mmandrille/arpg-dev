# v264 Spec - Organic Fog Silhouette

Status: Complete
Date: 2026-06-18
Codename: organic-fog-silhouette

## Purpose

Make the dungeon fog presentation feel less artificial by breaking up the perfectly circular
light/gloom/darkness boundary around the hero. The client should keep the same server-authoritative
visibility radius and LOS occlusion behavior, but render the fog edge with a deterministic irregular
silhouette so the darkness reads more like uneven dungeon gloom than a clean UI circle.

## Non-goals

- No server gameplay visibility, monster awareness, aggro, combat, protocol, or persistence change.
- No durable fog snapshots, explored-map persistence, minimap routefinding, or map memory changes.
- No imported fog art, shader plugin, Godot addon, particle system, production lighting pass, or
  full dungeon art treatment.
- No change to wall/door LOS shadow semantics from v255/v262.

## Acceptance Criteria

- The Godot fog overlay uses deterministic angular edge variation for the light/gloom/darkness
  transition instead of a perfectly circular distance boundary.
- The transition from gloom into full darkness has a softened feather instead of an abrupt outer
  cutoff.
- The visual variation is presentation-only: debug light/gloom radii and server visibility semantics
  stay unchanged.
- The organic edge is stable frame-to-frame for the same hero/camera state and does not flicker with
  wall shadow updates.
- Existing LOS shadow polygons still render over the organic fog mask, with a gloomy underlay and
  less-than-opaque dark core instead of a single full-black polygon.
- Bot/debug state exposes that the organic edge is active plus its configured intensity/segment
  values, darkness feather, and shadow gloom/core alpha values for scenario assertions.
- Existing fog overlay and LOS shadow client bot scenarios continue to pass with added organic-edge
  assertions.

## Scope and Likely Files

- Client:
  - `client/scripts/fog_of_war_overlay.gd`
  - `client/tests/test_fog_of_war_overlay.gd`
  - `client/scripts/bot_assertion_handlers.gd`
- Bot:
  - `tools/bot/scenarios/client/67_fog_of_war_overlay.json`
  - `tools/bot/scenarios/client/68_fog_los_shadow_mask.json`
- Docs:
  - `docs/plans/v264_2026-06-18-organic-fog-silhouette.md`
  - `docs/as-built/v264_organic-fog-silhouette.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, imported fog art, shader plugins, and Godot addons.
Borrow the existing in-repo `FogOfWarOverlay` shader path and extend its code-native mask.

## Test and Bot Proof

```bash
make client-unit
HEADLESS=1 make bot-visual scenario=67_fog_of_war_overlay
HEADLESS=1 make bot-visual scenario=68_fog_los_shadow_mask
make maintainability
```

Manual visual proof, if desired:

```bash
make bot-visual scenario=67_fog_of_war_overlay
make play
```

## Open Questions and Risks

- No required questions.
- Risk: too much edge variation can make the radius feel mechanically inaccurate. Keep debug radii
  unchanged and use modest code-native variation.
- Risk: shader math can be hard to prove visually in headless tests. Expose stable debug values and
  keep bot proof on existing fog scenarios.
