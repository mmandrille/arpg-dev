# v458 Spec: Camera Follow Smoothing Fix

Status: Implemented, awaiting `/finish`  
Date: 2026-07-08  
Codename: `camera-follow-smoothing-fix`  
Baseline: v457 `live-combat-transport-stability`

## Purpose

Remove the visible back-and-forth hero motion during ordinary floor traversal by letting the
existing isometric camera damping run continuously across local prediction and authoritative
reconciliation steps.

## Problem

`_reconcile_player()` currently snaps the camera to `PlayerAnchor` after every predicted or
authoritative position update. At the same time, `MovementVisualSmoothing` preserves the hero
mesh at its prior world position and eases it toward the moved anchor. The snapped camera and
offset hero therefore move in opposite screen-space directions before settling, producing the
recorded back-and-forth effect.

The damped camera also calls `look_at()` on the exact player position while its own position is
still catching up. That changes the isometric viewing angle during follow lag instead of keeping
the configured isometric orientation stable.

## Acceptance Criteria

- Ordinary local prediction and server reconciliation update `PlayerAnchor` without snapping the
  isometric camera.
- Initial setup, mode changes, and other explicit `sync_to_player()` calls retain snap behavior.
- Isometric damping preserves a fixed viewing direction derived from `follow_offset` while the
  camera position catches up.
- Player gameplay position, server authority, input cadence, collision, and movement tuning remain
  unchanged.
- Focused camera, movement smoothing, client unit, and click-to-move verification pass.

## Non-Goals

- No server, protocol, pathfinding, collision, or movement-speed changes.
- No new interpolation layer or external camera plugin.
- No chest-view behavior changes.
- No retuning of `follow_damping_seconds` or movement presentation values in this fix.

## Asset And Plugin Decision

- Adopt: the existing `PlayerCameraController`, `MovementVisualSmoothing`, and data-driven camera
  presentation catalog.
- Borrow: existing camera mode and movement smoothing unit/scenario coverage.
- Reject: external camera plugins and new assets; the defect is an integration conflict between
  existing in-repo presentation layers.

## Verification

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
godot --headless --path client --script res://tests/test_movement_visual_smoothing.gd
make client-unit
HEADLESS=1 make bot-visual scenario=80_movement_visual_smoothing
HEADLESS=1 make bot-visual scenario=05_click_to_move
make maintainability
```

Manual visual verification:

```bash
make bot-visual scenario=80_movement_visual_smoothing
```
