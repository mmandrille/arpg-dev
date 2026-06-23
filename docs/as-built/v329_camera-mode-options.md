# v329 As Built - Camera Mode Options

Date: 2026-06-23

## What Shipped

### Shared contracts
- `shared/assets/camera_presentations.v0.schema.json` + `camera_presentations.v0.json` — data-driven camera rig definitions for isometric, third_person, and chest_view modes with per-mode zoom, offsets, and presentation flags.

### Server
- No protocol or server changes; all camera logic is client-owned per ADR-0001 D2.

### Client
- `CameraPresentationsLoader` — singleton that loads and caches camera presentation data.
- `PlayerCameraContext` — typed context object passed to camera controller with player anchor, client settings, and chest socket reference.
- `PlayerCameraController` — owns the camera rig lifecycle: creates one Camera3D at setup time, switches between three rigs (isometric orthographic, third_person SpringArm3D + perspective, chest_view parented to chest_socket) via `apply_mode()`, and routes sync calls to the active rig. Fixes: detaches camera from spring arm before freeing it during teardown to prevent use-after-free when cycling modes. Camera shake offset is applied directly in `_sync_isometric` and `_sync_third_person` via `CameraImpactFeedbackScript.get_offset()`; chest_view has no camera shake (camera is parented to the chest socket node, so positioning is handled by the scene graph).
- `PerspectiveCombatInput` — handles mouse-look (camera yaw/pitch with configurable bounds), WASD camera-relative movement, and center-ray aiming for perspective modes. Disabled in isometric.
- `AimReticleOverlay` — renders a center-screen reticle when a perspective mode is active during gameplay.
- `ClientSettings.camera_mode` — added persistent setting with cycle method; defaults to isometric.
- Settings panel camera mode button row — three mutually-exclusive buttons to select mode; reflects current setting.
- V key cycling — `KeyV` pressed during gameplay calls `cycle_camera_mode()`, which cycles `isometric → third_person → chest_view → isometric` and saves immediately.
- Mouse capture during gameplay — perspective modes capture mouse when gameplay is active and no menu blocks input; isometric keeps free cursor. Capture is released when menus open (pause, settings, inventory).

### Bot / Testing
- Bot steps: `set_camera_mode`, `select_camera_mode` (for settings panel), `assert_camera_mode` (checks mode, projection, optionally mouse capture).
- Bot scenario 83: `83_camera_mode_setting.json` — validates settings-based mode selection, resume-from-pause persistence, and V key cycling through all three modes and back to start.
- Unit test: 31 assertions in `test_camera_mode_settings.gd` covering settings load, cycle behavior, persistence, and rig switching.

## Key Architectural Decisions

1. **PlayerCameraController owns rig lifecycle.** Extracted from `main.gd` to keep the monolith within maintainability ratchet while centralizing camera state and mode switching. Camera object is created once at setup and reused across mode changes; only the rig (spring arm, chest socket parenting) is rebuilt.

2. **Data-driven tuning via `camera_presentations.v0.json`.** Follow offset, zoom bounds, shoulder offset, spring arm length, pitch/yaw bounds, and reticle visibility are all configurable per mode. Eliminates camera-specific hardcoding and enables future per-class overrides.

3. **Perspective input disabled in isometric.** `PerspectiveCombatInput` is only active when camera mode is third_person or chest_view; isometric mode reverts to existing click-based combat and mouse picking. No input layering or priority conflicts.

4. **Mouse capture only during gameplay.** Menus (pause, settings, inventory, character panel) suppress mouse capture even in perspective modes; capture resumes when menus close. This is enforced via `_menu_blocks_gameplay_input()` in `_update_mouse_capture()`.

5. **Chest view uses existing chest_socket from item visuals.** The same socket used by equipment models is reused for chest view camera parenting. Falls back to torso if socket is missing; head submesh can be hidden to prevent camera clipping (not implemented in v329).

## Proof

```bash
make client-unit       # 31 assertions in test_camera_mode_settings.gd pass
make validate-shared   # camera_presentations schema + data validate
HEADLESS=1 make bot-visual scenario=83_camera_mode_setting  # all steps pass
```

## Open Deferred Items

- Scroll-wheel morphing between modes.
- Click-to-move in third-person and chest view.
- Per-class camera offset overrides in `class_presentations.v0.json`.
- Fog-of-war art polish for perspective modes (currently recycles isometric fog).
- Mouse sensitivity slider in Settings UI.
- Full hotkey remapping UI (V is currently fixed).

## Notes

- Extracted camera controller from main.gd during this slice; main.gd is now 5788 → 5825 lines (+37 due to new feature code in Tasks 4 and 7).
- Camera object is never freed; only rig components (spring arm, chest socket fallback) are rebuilt. This prevents stale camera references when cycling modes.
- The initial mode is read from `ClientSettings` at setup time, so returning to a previous saved mode loads immediately (no async delay).
