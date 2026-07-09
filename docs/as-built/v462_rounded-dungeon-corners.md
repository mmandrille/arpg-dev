# v462 As-Built — Rounded Dungeon Corners

Date: 2026-07-09
Spec: [`docs/specs/v462_spec-rounded-dungeon-corners.md`](../specs/v462_spec-rounded-dungeon-corners.md)
Plan: [`docs/plans/v462_2026-07-09-rounded-dungeon-corners.md`](../plans/v462_2026-07-09-rounded-dungeon-corners.md)

## What shipped

- Added `client/scripts/dungeon_wall_corner_presentation.gd` as the focused client helper that
  detects eligible negative-level `room_wall` L-turns, excludes perimeter/town/obstacle layouts,
  and attaches presentation-only geometry under the wall root without changing authoritative wall
  layout or collision.
- Locked the approved style to the stronger `rounded_cap` corner treatment by default while
  preserving the temporary bevel-vs-rounded selector for focused capture/testing paths.
- Extended the approved look beyond just the join caps: eligible dungeon `room_wall` runs now also
  get softened rounded top edges that reuse the same generated wall material, making room interiors
  read less blocky without introducing freestanding props.
- Kept `client/scripts/wall_renderer.gd` shallow by delegating both corner detection and top-edge
  generation to the helper, so the renderer still emits the same normalized wall layout and node
  count for authoritative walls before the presentation overlay is attached.
- Added focused unit coverage in `client/tests/test_factories.gd` for eligible corner detection,
  eligible room-wall run detection, stable corner-cap meshes, stable top-rounding meshes, and
  perimeter exclusion.
- Added `tools/bot/scenarios/client/100_rounded_dungeon_corners.json` as the dedicated generated
  wall lab proof while keeping `wall_floor_dungeon_rollout` green as the broader regression check.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
HEADLESS=1 make bot-client scenario=100_rounded_dungeon_corners
HEADLESS=1 make bot-client scenario=wall_floor_dungeon_rollout
make maintainability
```

## Manual visual command

```bash
godot --windowed --path client --script res://scripts/rounded_dungeon_corner_capture.gd -- --style rounded_cap --output .artifacts/showme/rounded-dungeon-corner-current.png
```

## Deferred

- Perimeter/generated wall joins still use the existing hard-edge treatment; this slice intentionally
  limited the new softened geometry to eligible dungeon `room_wall` runs first.
- Non-rectangular walkable space, curved collision, and obstacle silhouette redesign remain out of
  scope.
