# v467 Plan — Shared First-Person Equipment Rig

Status: Complete
Goal: Replace floating eye-view item nodes with a camera-local first-person rig that reuses the
same equipment socket and attack animation contract as the world hero.
Architecture: Keep authority unchanged. Add a focused first-person rig controller that owns a
camera-local `character.tscn` instance, mounts equipped hand items through the existing resolver
data path, and exposes an injected `AnimationController` for shared attack playback. The world hero
resolver remains the isometric presentation path; the first-person rig is a parallel render target
fed by the same authoritative inventory/equipped state.
Tech stack: Godot 4 GDScript client, shared JSON presentation data, Godot client bot scenario, SDD
docs.

## Baseline and Shortcut Decision

Builds on v466 `eye-view-weapon-presentation`, which proved camera mode, eye-view lighting, and
equipped-item visibility but used camera-child view models. This slice keeps the v466 camera and
lighting fixes, then replaces the view-model ownership model.

Asset/plugin decision:
- Adopt: none.
- Borrow: `client/scenes/character.tscn`, `CharacterVisual`, `gear_sockets.v0.json`,
  `item_visuals.v0.json`, `EquipmentVisualResolver` catalog/rules helpers, and
  `AnimationController`.
- Reject: external assets/plugins, new generated GLBs, and a separate first-person-only item
  visual catalog.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/first_person_equipment_rig.gd` | Camera-local rig, socket-mounted hand/off-hand visuals, debug, animation controller |
| Modify | `client/scripts/equipment_visuals.gd` | Delegate eye-view presentation to the rig and expose shared debug state |
| Modify | `client/scripts/combat_local_attack_presentation.gd` | Route selected attack clips to an optional first-person animation controller |
| Modify | `client/scripts/main.gd` | Wire first-person attack presentation without adding rig logic to the coordinator |
| Modify | `shared/assets/camera_presentations.v0.json` | Add first-person rig tuning fields if needed |
| Modify | `shared/assets/camera_presentations.v0.schema.json` | Validate any new tuning fields |
| Modify | `client/tests/test_item_visuals.gd` | Unit proof for rig socket mounting and two-handed/off-hand behavior |
| Modify | `client/tests/test_camera_mode_settings.gd` | Keep camera/light regression coverage green |
| Modify | `tools/bot/scenarios/client/105_eye_view_weapon_presentation.json` | Assert first-person rig/socket/debug behavior |
| Modify | `docs/CODEMAP.md` | Index new first-person rig file |
| Add | `docs/as-built/v467_shared-first-person-equipment-rig.md` | Slice proof and deferred work |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Lifecycle close-out |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/equipment_visuals.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: check before final
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [ ] Defer extraction with rationale: `<not expected>`

Verification:

```bash
make maintainability
```

## Task 1 — First-Person Rig Controller

Files:
- Create: `client/scripts/first_person_equipment_rig.gd`
- Modify: `shared/assets/camera_presentations.v0.json`
- Modify: `shared/assets/camera_presentations.v0.schema.json`

- [x] Step 1.1: Create a rig controller that instantiates `client/scenes/character.tscn` as a
  camera child while perspective eye-view is enabled.
- [x] Step 1.2: Mount hand/off-hand visuals under `right_hand_socket` / `off_hand_socket` inside
  the first-person rig using the same item visual data and transform semantics.
- [x] Step 1.3: Hide or crop obstructive first-person body geometry while leaving sockets and
  equipment visible.
- [x] Step 1.4: Expose debug state for rig active state, socket parent names, item ids, asset ids,
  clip, and attack count.

```bash
make validate-shared
```

## Task 2 — Resolver Integration

Files:
- Modify: `client/scripts/equipment_visuals.gd`
- Modify: `client/tests/test_item_visuals.gd`

- [x] Step 2.1: Replace the primary camera-child eye-view item nodes with the first-person rig
  controller.
- [x] Step 2.2: Keep world/isometric equipment mounting unchanged.
- [x] Step 2.3: Preserve two-handed/off-hand suppression for first-person rig visuals.
- [x] Step 2.4: Add focused unit tests for rig socket mount ownership and matching equipped ids.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_item_visuals.gd
```

## Task 3 — Shared Attack Animation Routing

Files:
- Modify: `client/scripts/combat_local_attack_presentation.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_item_visuals.gd`

- [x] Step 3.1: Let local attack presentation play the same selected clip/speed on an optional
  first-person rig animation controller.
- [x] Step 3.2: Remove the v466 eye-view pop tween as the primary motion path.
- [x] Step 3.3: Ensure isometric local attack behavior remains unchanged.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_item_visuals.gd
```

## Task 4 — Bot Proof And Debug

Files:
- Modify: `tools/bot/scenarios/client/105_eye_view_weapon_presentation.json`
- Modify: `client/scripts/bot_eye_view_assertions.gd` if existing assertion keys are insufficient
- Modify: `client/scripts/bot_assertion_handlers.gd` only if required by the assertion path

- [x] Step 4.1: Extend the existing scenario to assert first-person rig active state and socket
  ownership.
- [x] Step 4.2: Assert wall occlusion remains inactive in perspective mode.
- [x] Step 4.3: Assert attack count and clip/state after a local attack.

```bash
HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation
```

## Task 5 — Docs, CODEMAP, And Final Focused Verification

Files:
- Modify: `docs/CODEMAP.md`
- Add: `docs/as-built/v467_shared-first-person-equipment-rig.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: this plan

- [x] Step 5.1: Update `docs/CODEMAP.md` for the new rig controller.
- [x] Step 5.2: Add v467 as-built notes with proof, decisions, and deferred scope.
- [x] Step 5.3: Mark plan tasks complete after verification.

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_item_visuals.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation
make maintainability
```

## Final Verification

- [x] `make validate-shared`
- [x] `godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_item_visuals.gd`
- [x] `godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd`
- [x] `HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation`
- [x] `make maintainability`
- [x] `make ci`

## Deferred Scope

- Production hands/arms meshes.
- Bespoke first-person clip authoring per class and weapon family.
- Full FPS controller and collision redesign.
- External asset generation or imported art-pack evaluation.
