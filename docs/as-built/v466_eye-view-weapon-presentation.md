# v466 As-Built — Eye-View Weapon Presentation

Date: 2026-07-14
Spec: [`docs/specs/v466_spec-eye-view-weapon-presentation.md`](../specs/v466_spec-eye-view-weapon-presentation.md)
Plan: [`docs/plans/v466_2026-07-14-eye-view-weapon-presentation.md`](../plans/v466_2026-07-14-eye-view-weapon-presentation.md)

## What shipped

- Perspective `chest_view` now follows a head/eye transform while the gameplay camera stays mounted
  under the scene root, avoiding camera invalidation when the local body is hidden.
- Added data-owned eye-view framing in `shared/assets/camera_presentations.v0.json`: eye height,
  local body hiding, and view-model weapon transform.
- Added a local-only camera view-model mount for equipped hand items, reusing the existing item
  visual catalog, GLB assets, rarity tinting, and two-handed/off-hand blocking rules.
- Added bounded eye-view weapon attack motion and debug state (`attack_count`) without changing
  server authority, protocol, combat math, or inventory state.
- Added `105_eye_view_weapon_presentation`, a Godot-client visual proof that switches to
  perspective mode, equips the training bow, verifies the camera-local weapon, attacks, and checks
  the view-model motion counter.

## Proof

Focused verification:

```bash
make validate-shared
godot --headless --path client --script res://tests/test_camera_mode_settings.gd
godot --headless --path client --script res://tests/test_item_visuals.gd
.venv/bin/pytest tools/test_ci_pack.py tools/test_scenario_movement_audit.py -q
make client-unit
HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation
make maintainability
```

## Manual visual command

```bash
make bot-visual scenario=105_eye_view_weapon_presentation
```

## Deferred

- Bespoke first-person offsets and attack animations per class and weapon family.
- Production first-person hand/weapon art, VFX, and audio.
- Full FPS controller behavior, new collision/movement rules, or multiplayer camera synchronization.
