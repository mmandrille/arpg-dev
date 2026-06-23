# v329 Spec — Camera Mode Options

**Status:** Draft  
**Date:** 2026-06-23  
**Codename:** camera-mode-options

---

## Purpose

Give players three discrete camera modes for gameplay:

1. **Isometric** (default) — current orthographic follow camera with click-to-move and
   click-to-attack.
2. **Third-person** — perspective over-the-shoulder camera behind the hero with mouse look,
   camera-relative WASD movement, and aim-based attacks.
3. **Chest view** — perspective camera anchored at the character chest, looking forward with the
   player body (torso, arms, equipped gear) still visible.

Players select the active mode from **Settings** and can cycle modes at runtime with **`V`**
(`isometric → third_person → chest_view → isometric`). The choice persists locally in
`user://settings.json` and applies immediately in-session. Mouse-wheel zoom adjusts zoom only
**within** the active mode; there is no scroll-driven morphing between modes.

Perspective modes capture the mouse during gameplay, show a center-screen aim reticle, and use
mouse look plus center-ray aiming for basic attacks and skills that depend on facing direction.
Isometric mode keeps the existing free cursor and click combat path unchanged.

---

## Non-goals

- Scroll-wheel morphing or continuous blending between camera modes.
- Click-to-move, sustained click, or ground-plane click picking in third-person or chest view.
- Hiding the full player body in chest view; no separate first-person arms/weapon viewmodel rig.
- Mouse look in isometric mode.
- Server, protocol, replay, or shared-rules authority changes.
- Account-synced or server-persisted camera settings.
- Full controls remapping UI (hotkey is fixed to `V` for this slice).
- Per-class camera offset tuning beyond a shared data-driven default rig (defer per-class
  overrides).
- Fog-of-war art rebalance, dungeon lighting rebalance, or production camera/VFX assets.
- Co-op remote-player camera sync (local client only).
- Extracting the entire `main.gd` input stack beyond the camera controller and the minimum
  perspective-input branch required for this slice.

---

## Acceptance Criteria

### Settings and persistence

- Settings includes a localized **Camera mode** control with exactly three options:
  `isometric`, `third_person`, and `chest_view`.
- The selected mode persists in `user://settings.json`, reloads on client start, and applies
  immediately when changed from Settings.
- Players without the new saved field keep **isometric** as the default.

### Runtime cycling

- Pressing **`V`** during gameplay cycles modes in order:
  `isometric → third_person → chest_view → isometric`.
- Cycling updates the saved setting immediately (same persistence behavior as `L` loot-filter
  cycling).
- `V` is ignored while text input has focus or while gameplay input is blocked by menus,
  inventory, or pause UI.

### Isometric mode (regression)

- Camera uses orthographic projection with the current follow offset and zoom min/max behavior.
- Left-click move/attack, sustained click, entity picking, and mouse-wheel orthographic zoom
  behave as they do today.
- Mouse remains visible (not captured).

### Third-person mode

- Camera uses perspective projection with an over-the-shoulder follow rig (`SpringArm3D` or
  equivalent) that pulls in on wall/ceiling collision.
- Mouse is captured during gameplay; menus/inventory/pause release capture.
- Mouse look controls camera yaw (and bounded pitch).
- WASD movement remains camera-relative.
- Left-click basic attacks aim along a flat direction derived from a **center-screen ray**, not
  ground-plane mouse picking.
- Mouse wheel adjusts follow distance within configured bounds for this mode.
- Click-to-move and sustained-click auto pathing are disabled.

### Chest view mode

- Camera uses perspective projection anchored at the character **`chest_socket`** (existing rig
  socket used by item visuals) with a small forward offset defined in shared camera data.
- Torso, arms, belt gear, and weapons remain visible; the slice may hide only the head/helmet
  submesh if required to prevent camera clipping.
- Input behavior matches third-person mode (mouse look, WASD, center-ray aim, captured mouse,
  wheel zoom within chest-mode bounds).
- Chest view is visually distinct from third-person (chest-forward framing, not behind-the-shoulder).

### Shared presentation

- A thin center reticle is visible in third-person and chest view only.
- Camera impact feedback from v328 continues to work in isometric and does not leave the rig in
  a broken offset state in perspective modes.
- Screen-attached UI that already follows `_camera` (floating combat text, health bars, fog
  overlay) remains functional without crashes in all three modes; minor perspective placement drift
  is acceptable in v1.

### Bot and debug

- Client debug/bot state exposes the active `camera_mode` string and whether mouse capture is
  active.
- `bot_controller.gd` accepts `KEY_V` in `press_key` steps.

---

## Scope and Likely Files

### New client modules

- `client/scripts/player_camera_controller.gd` — mode state machine, rig setup, follow sync, per-mode
  zoom, mouse capture transitions, projection switching. Uses a narrow typed context object (v216
  pattern); no `helpers=globals()`.
- `client/scripts/perspective_combat_input.gd` (or equivalent focused helper) — center-ray aim
  direction, perspective LMB attack dispatch, and guards that disable click-to-move paths outside
  isometric mode.
- `client/tests/test_camera_mode_settings.gd` — settings parse/save, panel sync, mode cycle order,
  projection/capture transitions.

### Shared data (data-driven tuning)

- `shared/assets/camera_presentations.v0.json`
- `shared/assets/camera_presentations.v0.schema.json`

Suggested fields: per-mode projection type, follow distance / ortho size defaults and zoom bounds,
shoulder offset, chest forward offset, mouse sensitivity, pitch clamps, spring-arm length/collision
mask, reticle enabled flag. Values are presentation-only.

### Client settings and UI

- `client/scripts/client_settings.gd` — `camera_mode` load/save/normalize
- `client/scripts/settings_panel.gd` — camera mode selector (button row pattern like monster health
  bars)
- `client/scripts/main.gd` — delegate camera sync/zoom to controller; branch input handling by mode;
  wire `V`, settings callbacks, and perspective attack aim
- `client/scripts/client_constants.gd` — remove or relocate camera literals superseded by shared
  camera data where practical
- `client/scripts/camera_impact_feedback.gd` — apply shake to rig pivot rather than hard-coded
  `CAMERA_FOLLOW_OFFSET`
- `client/scripts/combat_event_presentation.gd` — continue binding the active gameplay camera

### Systems to retest (minimal touch expected)

- `client/scripts/fog_of_war_overlay.gd`
- `client/scripts/damage_number.gd`
- `client/scripts/monster_health_bar.gd`
- `client/scripts/input_shadow_overlay.gd`

### Bot / tooling

- `client/scripts/bot_controller.gd` — `KEY_V` parsing
- `client/scripts/bot_step_catalog.gd` / `bot_action_step_validator.gd` — if new debug assertions
  are added
- `tools/bot/scenarios/client/NN_camera_mode_setting.json` — settings selection + `V` cycle proof

### Localization

- `shared/i18n/en.json`
- `shared/i18n/es.json`

### Docs / maintainability

- `.maintainability/file-size-baseline.tsv` — net non-positive growth for `main.gd` if touched

### Explicitly out of scope

- `server/`
- `shared/protocol/`
- `shared/rules/` gameplay catalogs

### Asset / plugin decision

| Option | Decision |
|--------|----------|
| External camera addons/plugins | **Reject** — use Godot `Camera3D`, `SpringArm3D`, and in-repo rig sockets |
| New production character GLBs / FP viewmodels | **Reject** for this slice |
| Existing `chest_socket` on character rigs | **Borrow** — chest-view anchor |
| Item mount socket pattern in `item_visuals.v0.json` | **Borrow** — same socket discovery/fallback approach |
| Dedicated `camera_presentations.v0.json` | **Adopt** — presentation tuning owner for zoom, offsets, sensitivity |

---

## Test and Bot Proof

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
make client-unit
HEADLESS=1 make bot-visual scenario=NN_camera_mode_setting
make maintainability
```

**Unit tests must cover:**

- Default / unknown `camera_mode` normalization → `isometric`
- Save/reload persistence shape in `user://settings.json`
- Settings panel reflects saved mode
- `V` cycle order and persistence side effect
- Projection type per mode (`orthogonal` vs `perspective`)
- Mouse capture enabled only in perspective modes during gameplay
- Isometric click-pick path remains reachable when mode is `isometric`

**Bot scenario must cover:**

- Open Settings and select `third_person`
- Assert debug `camera_mode == "third_person"`
- Press `V` twice and assert `chest_view`, then `isometric`
- Existing visual scenarios remain pinned to isometric by default metadata; no mass scenario churn
  required in this slice

**Manual visual check:**

```bash
make play
```

Verify in town and dungeon: mode switch via Settings and `V`, mouse look in perspective modes,
visible equipped gear in chest view, and isometric click combat unchanged after returning to
isometric.

---

## Open Questions and Risks

| # | Item | Plan default if unresolved |
|---|------|----------------------------|
| Q-1 | Head/helmet mesh hide in chest view | Hide head submesh only when `chest_view` is active if clipping occurs |
| Q-2 | Crosshair style | Thin code-native center reticle; no new texture assets |
| Q-3 | Mouse sensitivity owner | `camera_presentations.v0.json` single global default; not exposed in Settings yet |
| Q-4 | Fog overlay correctness in perspective | Fix only crash/incorrect total failure; polish pass deferred |
| R-1 | **Input split complexity** — largest slice risk; requires clean mode guards to avoid breaking isometric click combat | Mode checks live in extracted helpers, not scattered one-offs |
| R-2 | **main.gd size** — grandfathered coordinator | Camera extraction must be net non-positive line growth on `main.gd` |
| R-3 | **Headless mouse-capture proof** — limited per v37 | Bot asserts mode/capture state; full look/aim feel is visual/manual |
