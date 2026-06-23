# v329 Plan — Camera Mode Options

**Status:** Ready for implementation  
**Goal:** Ship three discrete client camera modes (isometric, third-person, chest view) with Settings
persistence, `V` cycling, and mouse-look perspective combat.  
**Architecture:** Extract a typed `PlayerCameraController` (v216 context pattern) fed by new
`camera_presentations.v0.json` tuning data. Keep isometric click combat untouched behind explicit
mode guards; perspective modes use a focused `PerspectiveCombatInput` helper for center-ray aim and
disable click-to-move. No server or protocol changes.  
**Tech stack:** Shared JSON schema + Godot 4 client (`Camera3D`, `SpringArm3D`) + Python/Godot bot
scenario proof.

---

## Baseline and shortcut decision

| Item | Value |
|------|-------|
| Builds on | v328 `camera-impact-feedback` complete |
| Branch | Current checkout only |
| Process note | `PROGRESS.md` still recommends `$review` / `$refactor` before the next feature batch; this slice may proceed in parallel if accepted |

**Asset / plugin decision (from spec):**

| Choice | Decision |
|--------|----------|
| External camera addons | **Reject** |
| New character GLBs / FP viewmodels | **Reject** |
| `chest_socket` on `CharacterVisual` | **Borrow** — `FALLBACK_SOCKETS["chest_socket"]` at `(0, 1.08, 0)` plus `find_child` discovery |
| Item mount socket pattern | **Borrow** for anchor resolution |
| `camera_presentations.v0.json` | **Adopt** — presentation tuning owner |

**Reuse:**

- Settings persistence pattern from v266 (`map_opacity`, monster health bar mode buttons).
- `DirectionalAttackInput.payload()` for perspective LMB attacks (new aim source only).
- `character_visual.gd` socket nodes for chest anchor.
- `collision_lab` world preset for focused client bot scenario (same as scenario 74).

---

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/assets/camera_presentations.v0.json` | Per-mode projection, zoom bounds, offsets, sensitivity, pitch clamps, spring-arm length |
| Create | `shared/assets/camera_presentations.v0.schema.json` | Schema validation |
| Create | `client/scripts/camera_presentations_loader.gd` | `class_name` + `ensure_loaded()` singleton |
| Create | `client/scripts/player_camera_controller.gd` | Mode machine, rig nodes, follow sync, zoom, capture, projection switch |
| Create | `client/scripts/player_camera_context.gd` | Narrow typed context for controller (player anchor, settings, gameplay camera ref) |
| Create | `client/scripts/perspective_combat_input.gd` | Center-ray flat aim, perspective attack guards |
| Create | `client/scripts/aim_reticle_overlay.gd` | Thin code-native center reticle (perspective modes only) |
| Create | `client/tests/test_camera_mode_settings.gd` | Settings, panel, controller, cycle, capture tests |
| Create | `tools/bot/scenarios/client/83_camera_mode_setting.json` | Bot/visual proof |
| Modify | `client/scripts/client_settings.gd` | `camera_mode` normalize/load/save/cycle |
| Modify | `client/scripts/settings_panel.gd` | Camera mode button row + signal |
| Modify | `client/scripts/main.gd` | Wire controller, `V`, mode-gated input, bot API; net shrink vs baseline |
| Modify | `client/scripts/client_constants.gd` | Keep only literals not owned by camera data |
| Modify | `client/scripts/camera_impact_feedback.gd` | Shake rig pivot, not hard-coded offset |
| Modify | `client/scripts/combat_event_presentation.gd` | Bind active gameplay `Camera3D` from controller |
| Modify | `client/scripts/character_visual.gd` | Optional head-submesh visibility toggle for chest view clipping |
| Modify | `client/scripts/bot_controller.gd` | `KEY_V` parse + `set_camera_mode` dispatch |
| Modify | `client/scripts/bot_step_catalog.gd` | Register new steps |
| Modify | `client/scripts/bot_action_step_validator.gd` | Validate new steps |
| Modify | `client/scripts/bot_scenario_runner.gd` | `select_camera_mode`, `assert_camera_mode`, `set_camera_mode` handlers |
| Modify | `scripts/client_smoke.sh` | Register `test_camera_mode_settings.gd` |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Camera mode labels |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `main.gd` baseline after extraction |

**Retest only (fix if broken):** `fog_of_war_overlay.gd`, `damage_number.gd`, `monster_health_bar.gd`,
`input_shadow_overlay.gd`

**Out of scope:** `server/`, `shared/protocol/`, `shared/rules/`, replay tools

---

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [ ] `client/scripts/main.gd` (baseline **5788**)
- [ ] `server/internal/game/game_test.go` — not touched
- [ ] `tools/bot/run.py` — not touched
- [ ] `tools/validate_shared.py` — not touched
- [ ] Other over-limit file: none expected

Decision:

- [x] Extract `player_camera_controller.gd`, `player_camera_context.gd`, `perspective_combat_input.gd`,
  `camera_presentations_loader.gd`, and `aim_reticle_overlay.gd` in this slice.
- [ ] Defer extraction with rationale: N/A

Touch-to-shrink requirement: `main.gd` must end **≤ 5788 lines** (preferably lower). Move
`_sync_camera_to_player`, `_adjust_camera_zoom`, mode state, rig setup, and capture helpers into the
controller; leave only thin delegation and mode gates in `main.gd`.

Verification:

```bash
make maintainability
```

---

## Task 1 — Shared camera presentation contract

Files:

- Create: `shared/assets/camera_presentations.v0.json`
- Create: `shared/assets/camera_presentations.v0.schema.json`
- Create: `client/scripts/camera_presentations_loader.gd`

- [ ] Step 1.1: Add schema-backed catalog with three modes (`isometric`, `third_person`,
  `chest_view`). Include at minimum: `projection`, `zoom_default`, `zoom_min`, `zoom_max`,
  `follow_offset` or `spring_arm_length`, `shoulder_offset`, `chest_forward_offset`,
  `mouse_sensitivity`, `pitch_min_degrees`, `pitch_max_degrees`, `reticle_enabled`.
- [ ] Step 1.2: Implement loader (`class_name CameraPresentationsLoader extends RefCounted`,
  `static func ensure_loaded()`, `static func mode(name: String) -> Dictionary`).
- [ ] Step 1.3: Add focused unit assertions in `test_camera_mode_settings.gd` for loader defaults
  and unknown-mode fallback to isometric tuning.

```bash
make validate-shared
```

---

## Task 2 — Client settings and Settings panel

Files:

- Modify: `client/scripts/client_settings.gd`
- Modify: `client/scripts/settings_panel.gd`
- Modify: `shared/i18n/en.json`, `shared/i18n/es.json`
- Modify: `client/scripts/main.gd` (settings callbacks only in this step)

- [ ] Step 2.1: Add `camera_mode` with constants
  `CAMERA_MODE_ISOMETRIC`, `CAMERA_MODE_THIRD_PERSON`, `CAMERA_MODE_CHEST_VIEW`, normalize unknown →
  `isometric`, persist in `user://settings.json`.
- [ ] Step 2.2: Add `cycle_camera_mode()` returning the next mode in order
  `isometric → third_person → chest_view → isometric`; save on cycle.
- [ ] Step 2.3: Add Settings panel button row (mirror monster health bar pattern) + localized labels
  + `camera_mode_selected(mode)` signal.
- [ ] Step 2.4: Wire `main.gd` settings callback to apply mode through the camera controller and
  `_sync_settings_panel()`.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
```

---

## Task 3 — Player camera controller and rigs

Files:

- Create: `client/scripts/player_camera_context.gd`
- Create: `client/scripts/player_camera_controller.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/camera_impact_feedback.gd`

- [ ] Step 3.1: Define `PlayerCameraContext` with: `player_anchor`, `character_visual`,
  `client_settings`, accessor to gameplay `Camera3D`, and callbacks for menu/input-lock queries.
- [ ] Step 3.2: Build scene rig under controller ownership:
  - Isometric: orthographic `Camera3D` at `CAMERA_FOLLOW_OFFSET` equivalent from data.
  - Third-person: `SpringArm3D` + perspective `Camera3D` with shoulder offset; collision pull-in on
    environment layer used by dungeon walls.
  - Chest view: perspective `Camera3D` parented to resolved `chest_socket` on `character_visual`
    with forward offset from data.
- [ ] Step 3.3: Implement `apply_mode(mode)`, `sync_to_player()`, `adjust_zoom(delta)`, mouse
  capture release/restore, and `get_gameplay_camera()`.
- [ ] Step 3.4: Replace `main.gd` `_sync_camera_to_player` / `_adjust_camera_zoom` bodies with
  controller delegation; instantiate controller in `_build_scene()`.
- [ ] Step 3.5: Update `CameraImpactFeedback` to shake a rig pivot node supplied by the controller
  instead of assigning `ClientConstants.CAMERA_FOLLOW_OFFSET` directly.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
make maintainability
```

---

## Task 4 — Perspective input, reticle, and isometric guards

Files:

- Create: `client/scripts/perspective_combat_input.gd`
- Create: `client/scripts/aim_reticle_overlay.gd`
- Modify: `client/scripts/main.gd`

- [ ] Step 4.1: Add `PerspectiveCombatInput.flat_aim_direction(camera, player_anchor)` using
  viewport center ray projected to ground plane (or camera forward flattened to XZ).
- [ ] Step 4.2: In `main.gd` `_unhandled_input` / `_handle_input`, gate paths:
  - **Isometric:** existing LMB click pick, sustained click, ground aim — unchanged.
  - **Third-person / chest view:** skip click-to-move and sustained click; on LMB use center-ray aim
    with `DirectionalAttackInput.payload()`; accumulate mouse motion for yaw/pitch when captured.
- [ ] Step 4.3: Bind **`V`** in `_unhandled_input` to `client_settings.cycle_camera_mode()` +
  controller `apply_mode()`; ignore when text focus or `_menu_blocks_gameplay_input()`.
- [ ] Step 4.4: Capture mouse in perspective modes during gameplay; release on pause/inventory/
  settings open; restore visible cursor on isometric.
- [ ] Step 4.5: Add `AimReticleOverlay` on gameplay UI layer; visible only when
  `camera_presentations` reticle flag true and mode is perspective.
- [ ] Step 4.6: Optional chest-view head clip fix: toggle head/helmet submesh visibility on
  `character_visual` only while `chest_view` active if visual clipping occurs (skip if models are
  clean in `make play` spot-check).

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
```

---

## Task 5 — Dependent presentation systems smoke check

Files:

- Retest: `client/scripts/fog_of_war_overlay.gd`, `damage_number.gd`, `monster_health_bar.gd`,
  `input_shadow_overlay.gd`, `combat_event_presentation.gd`

- [ ] Step 5.1: Ensure all screen-attached overlays still receive the active gameplay camera from
  controller; fix crashes or null refs only (placement drift acceptable).
- [ ] Step 5.2: Run existing fog/damage unit or smoke tests that bind a `Camera3D` if any regress.

```bash
make client-unit
```

---

## Task 6 — Unit tests and client smoke registration

Files:

- Create: `client/tests/test_camera_mode_settings.gd`
- Modify: `scripts/client_smoke.sh`

- [ ] Step 6.1: Cover `ClientSettings` normalize/save/reload/`cycle_camera_mode()`.
- [ ] Step 6.2: Cover Settings panel selected button sync for all three modes (en + es label spot
  check optional).
- [ ] Step 6.3: Cover `PlayerCameraController` projection per mode and mouse-capture policy using a
  minimal scene tree harness (no full `main.gd` scene-graph paths).
- [ ] Step 6.4: Assert isometric mode keeps `Input.MOUSE_MODE_VISIBLE` policy in unit helper.
- [ ] Step 6.5: Register test in `client_smoke.sh`.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
make client-unit
```

---

## Task 7 — Bot scenarios

Files:

- Modify: `client/scripts/bot_controller.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_action_step_validator.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/scripts/main.gd` (`get_bot_state`, `bot_set_camera_mode`)
- Create: `tools/bot/scenarios/client/83_camera_mode_setting.json`

- [ ] Step 7.1: Add `KEY_V` to `bot_controller._parse_keycode`.
- [ ] Step 7.2: Add bot steps:
  - `set_camera_mode` `{ "mode": "third_person" }`
  - `select_camera_mode` `{ "mode": "chest_view" }` (clicks Settings panel button when visible)
  - `assert_camera_mode` `{ "mode": "isometric", "projection": "orthogonal", "mouse_captured": false }`
- [ ] Step 7.3: Expose in `get_bot_state()`:
  - `camera_mode` (string)
  - `camera_projection` (`orthogonal` | `perspective`)
  - `mouse_captured` (bool)
- [ ] Step 7.4: Add scenario `83_camera_mode_setting.json`:

```json
{
  "id": "camera_mode_setting",
  "runner": "godot_client",
  "world_id": "collision_lab",
  "seed": "camera_mode_setting_seed",
  "title": "Camera mode setting and V cycle",
  "client_steps": [
    { "type": "wait_ws_open", "timeout_s": 10.0 },
    {
      "type": "assert_camera_mode",
      "mode": "isometric",
      "projection": "orthogonal",
      "mouse_captured": false
    },
    { "type": "press_key", "keycode": "KEY_ESCAPE" },
    { "type": "wait_pause_menu", "timeout_s": 3.0 },
    { "type": "click_menu_button", "button": "settings" },
    { "type": "wait_settings_panel", "timeout_s": 3.0 },
    { "type": "select_camera_mode", "mode": "third_person" },
    {
      "type": "assert_camera_mode",
      "mode": "third_person",
      "projection": "perspective",
      "mouse_captured": false
    },
    { "type": "click_menu_button", "button": "back" },
    { "type": "click_menu_button", "button": "resume" },
    {
      "type": "assert_camera_mode",
      "mode": "third_person",
      "projection": "perspective",
      "mouse_captured": true
    },
    { "type": "press_key", "keycode": "KEY_V" },
    {
      "type": "assert_camera_mode",
      "mode": "chest_view",
      "projection": "perspective",
      "mouse_captured": true
    },
    { "type": "press_key", "keycode": "KEY_V" },
    {
      "type": "assert_camera_mode",
      "mode": "isometric",
      "projection": "orthogonal",
      "mouse_captured": false
    }
  ]
}
```

- [ ] Step 7.5: Do **not** migrate existing visual scenarios; they remain isometric by default.

```bash
HEADLESS=1 make bot-visual scenario=83_camera_mode_setting
```

---

## Task 8 — Lifecycle docs and CI

- [ ] Step 8.1: On slice completion, update `PROGRESS.md` and add `docs/as-built/v329_camera-mode-options.md`.
- [ ] Step 8.2: Update `docs/progress/slice-lifecycle.md` row when shipping.

```bash
make ci
```

---

## Final verification

- [ ] `make maintainability`
- [ ] `make validate-shared`
- [ ] `make client-unit`
- [ ] `HEADLESS=1 make bot-visual scenario=83_camera_mode_setting`
- [ ] `make ci`
- [ ] Manual: `make play` — verify Settings + `V`, mouse look in perspective modes, visible gear in
  chest view, isometric click combat unchanged after returning to isometric

---

## Deferred (explicit)

- Scroll-wheel morphing between modes
- Click-to-move in perspective modes
- Full controls remapping for `V`
- Per-class camera offsets in `class_presentations.v0.json`
- Fog-of-war art polish for perspective cameras
- Mouse sensitivity slider in Settings UI
- Headless proof of mouse-look aim accuracy (visual/manual only per v37)
