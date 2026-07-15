# v466 Plan — Eye-View Weapon Presentation

Status: Ready for implementation
Goal: Make perspective camera combat feel like the hero's eyes by showing equipped hand items and attack motion.
Architecture: This is a client presentation slice. It reuses the existing perspective camera mode,
equipment visual resolver, hand sockets, and client-only attack feedback while keeping all
authoritative combat, movement, inventory, and protocol state unchanged.
Tech stack: Godot client, shared asset JSON/schema, client bot scenario, lifecycle docs.

## Baseline and shortcut decision

Builds on v465 `combat-impact-confirmation`, v464 `combat-input-flow-polish`, v461 camera/movement
smoothing, and the existing `chest_view` perspective mode. Adopt existing camera/equipment/attack
presentation systems; borrow recent client visual bot patterns; reject external first-person
plugins, controller rewrites, and new asset pipelines.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/assets/camera_presentations.v0.schema.json` | Add optional eye-view/view-model framing fields. |
| Modify | `shared/assets/camera_presentations.v0.json` | Own eye-height, body visibility, and view-model offset tuning as data. |
| Modify | `client/scripts/player_camera_controller.gd` | Position perspective camera at eye height and expose debug state. |
| Modify | `client/scripts/equipment_visuals.gd` | Mount a local eye-view hand item when perspective mode needs a camera-visible weapon. |
| Modify | `client/tests/test_camera_mode_settings.gd` | Cover eye-view camera data and debug state. |
| Modify | `client/tests/test_item_visuals.gd` | Cover eye-view mount state without regressing ordinary equipment visuals. |
| Add | `tools/bot/scenarios/client/105_eye_view_weapon_presentation.json` | Visual proof for perspective weapon visibility and attack motion. |
| Modify | `docs/progress/slice-lifecycle.md` | Mark v466 lifecycle row complete at close-out. |
| Modify | `PROGRESS.md` | Advance current status and deferred scope at close-out. |
| Add | `docs/as-built/v466_eye-view-weapon-presentation.md` | Summarize shipped proof and limits. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `server/internal/game/game_test.go`
- [x] `tools/bot/run.py`
- [x] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: touched files are currently below 600 lines and the slice is scoped to existing focused presentation modules.

Verification:
```bash
make maintainability
```

## Task 1 — Shared camera presentation data

Files:
- Modify: `shared/assets/camera_presentations.v0.schema.json`
- Modify: `shared/assets/camera_presentations.v0.json`

- [x] Step 1.1: Add optional data fields for perspective eye height, body visibility policy, and view-model weapon offset.
```bash
make validate-shared
```

## Task 2 — Eye-view camera controller

Files:
- Modify: `client/scripts/player_camera_controller.gd`
- Modify: `client/tests/test_camera_mode_settings.gd`

- [x] Step 2.1: Place perspective mode at eye/head height using shared presentation data.
- [x] Step 2.2: Add debug state for active mode, camera projection, configured eye height, and body visibility policy.
- [x] Step 2.3: Add focused camera tests.
```bash
godot --headless --path client --script res://tests/test_camera_mode_settings.gd
```

## Task 3 — Camera-visible equipped weapon

Files:
- Modify: `client/scripts/equipment_visuals.gd`
- Modify: `client/tests/test_item_visuals.gd`

- [x] Step 3.1: Add a local presentation-only view-model mount for equipped main-hand/off-hand visuals when eye-view is active.
- [x] Step 3.2: Keep ordinary isometric/skeleton-mounted equipment state unchanged.
- [x] Step 3.3: Add focused equipment visual tests for mounted eye-view item state.
```bash
godot --headless --path client --script res://tests/test_item_visuals.gd
```

## Task 4 — Visual bot proof

Files:
- Add: `tools/bot/scenarios/client/105_eye_view_weapon_presentation.json`

- [x] Step 4.1: Add a Godot-client scenario that switches to perspective mode, confirms eye-view weapon state, attacks a target, and observes attack motion.
- [x] Step 4.2: Run the focused visual scenario.
```bash
HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation
```

## Task 5 — Lifecycle docs and final focused verification

Files:
- Modify: `docs/specs/v466_spec-eye-view-weapon-presentation.md`
- Modify: `docs/plans/v466_2026-07-14-eye-view-weapon-presentation.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v466_eye-view-weapon-presentation.md`

- [x] Step 5.1: Update spec status, plan checkboxes, lifecycle row, progress status, and as-built proof.
- [x] Step 5.2: Run focused verification before finish.
```bash
make validate-shared
godot --headless --path client --script res://tests/test_camera_mode_settings.gd
godot --headless --path client --script res://tests/test_item_visuals.gd
make client-unit
HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation
make maintainability
```

## Final verification

- [x] `make validate-shared`
- [x] `godot --headless --path client --script res://tests/test_camera_mode_settings.gd`
- [x] `godot --headless --path client --script res://tests/test_item_visuals.gd`
- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation`
- [x] `make maintainability`
- [x] `$finish` runs `make ci` before commit if focused verification does not fully cover the changed surface.
