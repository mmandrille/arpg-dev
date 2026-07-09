# v461 Spec: Entity Locomotion Polish

Status: Complete
Date: 2026-07-09
Codename: `entity-locomotion-polish`
Baseline: v460 `leveled-potions`

## Purpose

Make ordinary dungeon traversal read as grounded walking instead of visible hopping for both the
local hero and nearby moving monsters. The slice keeps server authority, pathfinding, collision,
and protocol unchanged while polishing how existing client-side movement visual smoothing, entity
tick smoothing, and isometric camera follow interact during short authoritative corrections and
chase movement.

The immediate player-facing goal is that normal click-to-move and monster pursuit look smoother in
the dungeon without introducing misleading latency, delayed collisions, or presentation drift from
the authoritative simulation.

## Non-goals

- No protocol or schema version bump.
- No server-side movement, collision, or pathfinding rewrite.
- No new movement intents, input cadence changes, or prediction model changes.
- No leap, charge, teleport, or other discontinuous-movement presentation redesign.
- No broad gameplay speed slowdown by default; a tiny data-driven retune is allowed only if
  presentation-only polish is insufficient and the plan justifies it.
- No new external plugins, camera addons, or imported assets.

## Acceptance criteria

1. Ordinary local click-to-move traversal in the dungeon no longer shows obvious hero
   back-and-forth or snap-like walking during small authoritative corrections.
2. Nearby monsters and other non-local walking entities visually traverse in smoother short
   segments during chase movement and read as walking rather than hopping between updates.
3. The isometric camera no longer amplifies perceived snapping during ordinary traversal; routine
   follow behavior remains continuous across local prediction and authoritative correction.
4. Gameplay authority remains unchanged: authoritative positions, collision, pathfinding, and
   server-owned outcomes behave the same as before this slice.
5. Existing movement presentation proofs stay green, and at least one focused visual proof covers
   monster-side locomotion if the existing scenarios are not sufficient to judge that behavior.

## Scope and likely files touched

- Client movement integration:
  - `client/scripts/main.gd`
  - `client/scripts/movement_visual_smoothing.gd`
  - `client/scripts/entity_tick_smoothing_runtime.gd`
  - `client/scripts/player_camera_controller.gd`
  - `client/scripts/player_movement_feel.gd`
- Client presentation tuning:
  - `shared/assets/movement_presentation.v0.json`
  - `shared/assets/camera_presentations.v0.json`
  - `shared/assets/combat_feel_presentation.v0.json` only if an existing presentation-owned field
    is the correct owner for the final tuning
- Client tests:
  - `client/tests/test_movement_visual_smoothing.gd`
  - `client/tests/test_entity_tick_smoothing.gd`
  - `client/tests/test_camera_mode_settings.gd`
- Bot / visual proof:
  - existing `tools/bot/scenarios/client/80_movement_visual_smoothing.json`
  - existing `tools/bot/scenarios/client/84_entity_tick_smoothing.json`
  - existing `tools/bot/scenarios/client/05_click_to_move.json`
  - optional new focused monster locomotion proof scenario if needed
- Docs:
  - `docs/plans/v461_2026-07-09-entity-locomotion-polish.md`
  - `docs/as-built/v461_entity-locomotion-polish.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

## Test and bot proof

Focused verification is expected during implementation:

```bash
make validate-shared
make client-unit
HEADLESS=1 make bot-visual scenario=80_movement_visual_smoothing
HEADLESS=1 make bot-visual scenario=84_entity_tick_smoothing
HEADLESS=1 make bot-visual scenario=05_click_to_move
make maintainability
```

If the implementation adds a dedicated monster locomotion scenario, run that scenario as part of
focused verification too. Manual visual confirmation should remain available via:

```bash
make bot-visual scenario=80_movement_visual_smoothing
```

or the monster-focused scenario chosen in the plan.

## Asset and plugin decision

- Adopt: existing in-repo movement smoothing, tick smoothing, camera presentation, and client-bot
  visual proof systems.
- Borrow: prior movement slices v299, v349, v367, v368, and the in-flight v458 camera follow
  smoothing fix behavior as the baseline integration to preserve.
- Reject: external camera plugins, new asset pipelines, and production art changes; this slice is
  purely about improving the existing locomotion presentation stack.

## Open questions and risks

1. `v458` (`camera-follow-smoothing-fix`) has a spec and plan but no lifecycle completion row. The
   plan must confirm whether that behavior already landed in code and whether this slice should
   explicitly close it out or only build on it.
2. If the hero still looks jumpy after camera/follow integration is verified, the plan must decide
   whether to tune visual smoothing only or allow a very small shared-rules movement-speed retune.
3. Existing scenarios may prove smoothing activation/settling but still be weak for judging monster
   chase readability. The plan should add a focused monster visual scenario if current proof is not
   perceptual enough.
