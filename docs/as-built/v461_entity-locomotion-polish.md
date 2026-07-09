# v461 As-Built — Entity Locomotion Polish

Date: 2026-07-09
Spec: [`docs/specs/v461_spec-entity-locomotion-polish.md`](../specs/v461_spec-entity-locomotion-polish.md)
Plan: [`docs/plans/v461_2026-07-09-entity-locomotion-polish.md`](../plans/v461_2026-07-09-entity-locomotion-polish.md)

## What shipped

- Isometric camera follow now tracks the rendered local hero position during ordinary locomotion,
  so camera damping follows the visible body instead of exaggerating the gap between the
  authoritative `PlayerAnchor` and the smoothed `CharacterVisual`.
- Camera follow tuning was softened from `0.12` to `0.16` seconds to reduce visible hopiness
  during short authoritative corrections without changing gameplay authority.
- Remote entity locomotion smoothing now uses a gentler adaptive duration window
  (`0.11`–`0.18` seconds, `6.5` distance-per-second) so nearby monsters and other non-local
  walkers glide through 10 Hz updates more like walking than hopping.
- Local hero visual smoothing catch-up was relaxed slightly (`catch_up_speed` `18.0 -> 14.0`,
  `max_offset` `0.7 -> 0.8`) so small correction steps settle more naturally instead of snapping
  back too aggressively.
- Added focused regression coverage for camera follow bias toward the rendered hero plus runtime
  checks that remote adaptive smoothing clamps to the configured min/max durations.

## Proof

Focused verification:

```bash
make validate-shared
godot --headless --path client --script res://tests/test_camera_mode_settings.gd
godot --headless --path client --script res://tests/test_entity_tick_smoothing.gd
godot --headless --path client --script res://tests/test_movement_visual_smoothing.gd
make client-unit
HEADLESS=1 make bot-visual scenario=80_movement_visual_smoothing
HEADLESS=1 make bot-visual scenario=84_entity_tick_smoothing
HEADLESS=1 make bot-visual scenario=05_click_to_move
HEADLESS=1 make bot-visual scenario=25_ranged_monster_ai
make maintainability
make ci
```

## Manual visual commands

```bash
make bot-visual scenario=80_movement_visual_smoothing
make bot-visual scenario=25_ranged_monster_ai
```

## Deferred

- Broader shared-rules movement-speed or acceleration retuning remains deferred; v461 intentionally
  solved the reported locomotion feel issue through client-side presentation and smoothing first.
