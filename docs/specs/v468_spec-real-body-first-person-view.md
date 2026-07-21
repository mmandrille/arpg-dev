# v468 Spec: Real-Body First-Person View

Status: Draft
Date: 2026-07-21
Codename: `real-body-first-person-view`
Baseline: v467 `shared-first-person-equipment-rig`

## Purpose

Delete the fake `FirstPersonEquipmentRig` and replace it with the real character body already in
the scene. In chest_view mode the camera sits at eye height looking forward — the real character
with all its mounted equipment (boots, belt, chainmail, helmet, weapons, rings — everything) is
simply visible to that camera. No duplicate character instance, no fake hand proxies, no parallel
equipment mounting pipeline.

This consolidates the two-world problem introduced in v466-v467: any animation improvement,
equipment visual fix, or data-driven config change now applies to both isometric and first-person
automatically because there is only one character.

## Decision record

`base_human.glb` is a single unified mesh — no per-body-part render layer separation is possible
without art changes. The camera sits at eye_height 1.42; the head socket is at 1.55 (~0.13 units
above camera). The body geometry around the head will clip into the near plane.

**Near-clip approach:** Set `Camera3D.near` to a data-driven value for chest_view (default ~0.25).
This clips head/face geometry that would fill the camera interior while leaving hand geometry
(right_hand_socket at y=0.92, 0.5 units below eye) and all socket items well beyond the clip plane.
This is the same technique used by every mainstream first-person engine — Skyrim, Unreal first-person
templates, etc. — when using a unified character mesh.

If in-slice testing shows the body clips too aggressively or leaves a visible mesh seam, the
fallback is Godot visibility layers: assign the character body mesh to layer 2 and the FP camera to
cull layer 2, while leaving socket items on the default layer (visible to both cameras). This
requires no art asset change. Prefer near-clip first; add visibility layer masking only if needed.

## Non-goals

- No server, protocol, combat, inventory, or movement changes.
- No art asset rework (no mesh separation, no new hand/arm art).
- No full FPS controller rewrite, collision model change, or multiplayer camera sync.
- No bespoke per-class or per-weapon-family FP animation pass.

## What goes away

| Removed | Where |
|---------|-------|
| `FirstPersonEquipmentRig` class and file | `client/scripts/first_person_equipment_rig.gd` — deleted |
| Fake hand proxies (CapsuleMesh + SphereMesh) | gone with the class |
| `set_eye_view()` / `first_person_animation_controller()` / `record_first_person_attack()` | `equipment_visuals.gd` — removed |
| `_eye_view_rig` field and all references | `equipment_visuals.gd` — removed |
| `_sync_eye_view_equipment()` | `main.gd` — removed |
| FP rig animation controller arguments in `CombatLocalAttackPresentation.present_result` and `present_local_start` calls | `main.gd` — removed |
| `hide_local_body`, `first_person_rig_*`, `first_person_body_visible`, `first_person_hands_visible`, `view_model_*` config keys | `camera_presentations.v0.json` and schema — removed |

## What changes

**`camera_presentations.v0.json`** — add `near_clip` (float, default 0.25) to chest_view mode. Remove
all `first_person_*`, `view_model_*`, and `hide_local_body` keys.

**`player_camera_controller.gd`** — in `_setup_chest_view`, apply `_cfg.get("near_clip", 0.25)` to
`_camera.near`. Restore `_camera.near` to its default (0.05) when tearing down chest_view.

**`equipment_visuals.gd`** — remove `_eye_view_rig`, `_eye_view_camera`, `_eye_view_enabled`,
`_eye_view_cfg`, `set_eye_view`, `first_person_animation_controller`, `record_first_person_attack`,
`_refresh_eye_view`. Remove the `FirstPersonEquipmentRigScript` preload. The `get_debug_state` eye_view
key is replaced with a simple `{"enabled": false}` placeholder or removed from the schema entirely.
`equipment_visuals.gd` was already below 600 lines and stays there after removal.

**`main.gd`** — remove all call sites that pass `resolver.first_person_animation_controller()` and
`Callable(resolver, "record_first_person_attack")` to `CombatLocalAttackPresentation` calls (there
are four such sites). Remove `_sync_eye_view_equipment()` and its three call sites. Remove the
`character_visual.visible` assignment that hid the body. `main.gd` baseline shrinks; update the
maintainability baseline accordingly.

**`_sync_chest_view` in `player_camera_controller.gd`** — currently calls nothing about body
visibility; that was in `main.gd`'s `_sync_eye_view_equipment`. After removing that call, the
camera controller itself needs no body-visibility changes.

## Acceptance criteria

- In chest_view mode the real character node (`character_visual`) is visible; no `FirstPersonEquipmentRig`
  node or fake hand proxy exists anywhere in the scene tree.
- All equipped items — including those not directly in view (boots, belt, helmet, amulet) — are
  mounted on the real character's sockets and therefore exist in the scene's lighting and shadow
  computation while in chest_view.
- The hand socket items (main_hand, off_hand) are visible from the chest_view camera at the
  character's arm position.
- The camera near-clip in chest_view is data-driven from `camera_presentations.v0.json` and
  prevents visible head-mesh intrusion into the camera interior.
- Isometric equipment mounting is identical to before: zero regressions.
- `make validate-shared` passes.
- Headless unit tests for camera settings and item visuals pass.
- Bot scenario 105 (`eye_view_weapon_presentation`) passes; its assertions are updated to check
  real-body equipment state (slot mounted on `character_visual`, not on an `eye_view` rig).
- `make ci` green.

## Scope and likely files

| File | Change |
|------|--------|
| `client/scripts/first_person_equipment_rig.gd` | **Delete** |
| `client/scripts/equipment_visuals.gd` | Remove `_eye_view_rig` and all FP rig methods |
| `client/scripts/player_camera_controller.gd` | Apply `near_clip` in chest_view setup/teardown |
| `client/scripts/main.gd` | Remove FP rig call sites and `_sync_eye_view_equipment` |
| `shared/assets/camera_presentations.v0.json` | Remove FP rig keys; add `near_clip` |
| `shared/assets/camera_presentations.v0.schema.json` | Mirror schema change |
| `client/tests/test_camera_mode_settings.gd` | Update assertions: real body visible, near_clip set |
| `client/tests/test_item_visuals.gd` | Remove FP rig assertions; verify isometric slots unchanged |
| `tools/bot/scenarios/client/105_eye_view_weapon_presentation.json` | Update to real-body assertions |
| `.maintainability/file-size-baseline.tsv` | Lower `main.gd` baseline (net removal) |
| `docs/as-built/v468_real-body-first-person-view.md` | Close-out |
| `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Lifecycle |

## Test and bot proof

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_item_visuals.gd
make client-unit
HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation
make maintainability
make ci
```

Manual visual check (confirm real equipment visible, no fake hands, no head geometry clipping):
```bash
make bot-visual scenario=105_eye_view_weapon_presentation
```

## Asset and plugin decision

- Adopt: existing real character scene, gear sockets, equipment visual resolver, camera controller
  eye-height rig, Godot `Camera3D.near` property.
- Borrow: near-clip technique from mainstream FPS engines (no external dependency needed).
- Reject: visibility layer masking as first resort (try near-clip first; add layers only if needed
  after in-slice visual testing). Reject any art asset split in this slice.

## Maintainability impact

- `first_person_equipment_rig.gd` (208 lines): deleted outright.
- `equipment_visuals.gd`: net reduction (~20 lines).
- `main.gd`: net reduction (~15 lines from removed call sites and `_sync_eye_view_equipment`);
  lower the grandfathered baseline to the new actual count.

## Open questions

- Near-clip value: 0.25 is the starting point. In-slice testing will confirm whether head mesh
  clips cleanly or whether a value between 0.15 and 0.40 is needed. The value is data-driven in
  `camera_presentations.v0.json` so no code change is required to tune it.
- If the body mesh around the neck/collar is still visible at any reasonable near-clip, add a
  `first_person_body_layer` visibility layer mask as the fallback — assign the character mesh to
  layer 2 and the FP camera to cull layer 2, while socket item nodes remain on the default layer.
  This stays in scope for this slice if needed; document the decision in the as-built.
