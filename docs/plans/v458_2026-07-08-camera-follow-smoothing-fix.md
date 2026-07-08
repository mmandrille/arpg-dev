# v458 Plan: Camera Follow Smoothing Fix

Status: Implemented, awaiting `/finish`  
Spec: [`docs/specs/v458_spec-camera-follow-smoothing-fix.md`](../specs/v458_spec-camera-follow-smoothing-fix.md)

## Architecture

Keep `PlayerAnchor` as the exact predicted/authoritative gameplay position and keep
`CharacterVisual` as its presentation-only smoothed child. Stop routine reconciliation from
calling the camera's explicit snap API; the existing per-frame `tick_follow(delta)` becomes the
only ordinary isometric follow path. While damping, orient the camera from the configured
`follow_offset` so position lag does not alter the isometric angle.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/main.gd` | Stop snapping the camera during ordinary player reconciliation. |
| Modify | `client/scripts/player_camera_controller.gd` | Preserve fixed isometric orientation during damped follow. |
| Modify | `client/tests/test_camera_mode_settings.gd` | Prove follow advances without snapping and preserves orientation. |
| Create | `docs/specs/v458_spec-camera-follow-smoothing-fix.md` | Define defect, scope, and acceptance criteria. |
| Create | `docs/plans/v458_2026-07-08-camera-follow-smoothing-fix.md` | Record implementation and proof steps. |

## Tasks

- [x] Remove routine camera snap from `_reconcile_player()` while retaining explicit setup/mode
  synchronization behavior.
- [x] Keep the isometric camera basis aligned with `follow_offset` as its position damps.
- [x] Add focused camera follow regression coverage.
- [x] Run camera and movement unit tests.
- [x] Run client unit, click-to-move visual bot proof, and maintainability checks.
- [x] Run the normal-speed visible movement scenario; macOS capture was unavailable in the agent
  session, so final perceptual confirmation remains available through the manual command below.

## Verification Result

Green on 2026-07-08:

```text
test_camera_mode_settings: 35 passed, 0 failed
test_movement_visual_smoothing: 7 passed, 0 failed
make client-unit: PASS
movement_visual_smoothing: 1 passed, 0 failed (headless and visible)
click_to_move: 1 passed, 0 failed
make maintainability: PASS
```

Manual perceptual confirmation:

```bash
make bot-visual scenario=80_movement_visual_smoothing
```

## Risk Controls

- Do not change gameplay positions or network reconciliation.
- Do not change chest-view parenting or rotation ownership.
- Do not change data-driven damping values until the integration defect is removed and observed.
- Keep explicit snap behavior available for initial setup, camera mode changes, and discontinuities.
