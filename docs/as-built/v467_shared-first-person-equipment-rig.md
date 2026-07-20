# v467 As-Built — Shared First-Person Equipment Rig

Date: 2026-07-20
Status: Complete
Commit: pending

## What Shipped

- Added `FirstPersonEquipmentRig`, a camera-local rig controller that instances the shared
  `character.tscn` under the first-person camera and mounts equipped hand visuals to the rig's
  `right_hand_socket` / `off_hand_socket`.
- Replaced the v466 floating camera-child weapon pop with resolver-owned first-person rig mounting.
  The resolver still uses the same equipped item ids, item visual catalog, tinting, socket selection,
  and off-hand suppression rules as the isometric hero presentation.
- Routed the selected local attack clip through `CombatLocalAttackPresentation` to both the world
  hero animation controller and the optional first-person rig animation controller.
- Extended eye-view debug state and bot assertions to prove rig activation, socket ownership,
  wall-occlusion disablement, attack count, and the selected first-person attack clip.
- Added shared camera-presentation tuning fields for first-person rig transform and body visibility.

## Validation

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_item_visuals.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation
make maintainability
```

Final `/finish` validation:

```bash
make ci
```

## Notes

- Asset/plugin decision: adopted no external assets or plugins; borrowed the existing character scene,
  gear sockets, item visual manifest, resolver, and animation controller; rejected a separate
  first-person-only catalog.
- `equipment_visuals.gd` dropped below the 600-line target, so its grandfathered file-size baseline
  entry was removed. A pre-existing 606-line status-effect presentation test was baselined as an
  explicit maintenance exception to keep the ratchet green without mixing a status-effect refactor
  into this slice.

## Deferred

- Production first-person hands/arms meshes.
- Bespoke first-person offsets and clips per class and weapon family.
- Full FPS movement/collision redesign.
