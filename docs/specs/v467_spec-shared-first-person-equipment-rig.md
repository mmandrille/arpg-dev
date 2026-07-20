# v467 Spec — Shared First-Person Equipment Rig

Status: Draft
Date: 2026-07-20
Codename: `shared-first-person-equipment-rig`

## Purpose

Replace the v466 eye-view floating weapon presentation with a camera-local first-person equipment
rig that uses the same equipped-item data, socket names, and attack animation clip selection as the
isometric hero. In perspective `chest_view`, the player should see equipped hand/off-hand visuals
mounted to rig sockets instead of standalone items popped in front of the camera.

This is a client-presentation slice only. The server still owns inventory, combat outcomes, and
authoritative events; the client only renders the already-equipped state.

## Non-goals

- No server, realtime protocol, combat math, cooldown, damage, loot, or inventory authority changes.
- No protocol or golden fixture version bump.
- No production first-person hand art, weapon art, VFX, or audio pass.
- No full FPS movement/collision/controller redesign.
- No class-specific bespoke first-person animation library beyond reusing the current character
  clips and socket contract.
- No external assets or Godot plugins.

## Asset / Plugin Decision

- **Adopt:** none.
- **Borrow:** existing `client/scenes/character.tscn`, `CharacterVisual` socket creation,
  `gear_sockets.v0.json`, `item_visuals.v0.json`, `assets/manifests/assets.v0.json`,
  `EquipmentVisualResolver` item resolution, and `AnimationController` attack clip playback.
- **Reject:** imported hand/arms asset packs, new GLB generation, marketplace plugins, shader packs,
  and a separate first-person-only weapon catalog for this slice.

## Acceptance Criteria

- In `chest_view`, hand equipment is mounted under first-person rig sockets named
  `right_hand_socket` and `off_hand_socket`; the primary hand/off-hand nodes are not mounted
  directly under `PlayerCamera`.
- The first-person hand/off-hand item ids and asset ids match the same authoritative equipped state
  used by the world/isometric equipment resolver.
- Two-handed main-hand equipment suppresses the first-person off-hand visual using the same rule as
  world equipment.
- Local attack presentation routes the same selected attack clip/speed to both the world
  `AnimationController` and the first-person rig animation controller.
- The old v466 camera-local floating item tween is no longer the primary attack motion path.
- Existing isometric equipment mounting and local attack animation behavior remain unchanged.
- Eye-view camera height/light and perspective wall-occlusion fixes remain intact.
- Bot/debug state exposes enough data to prove first-person rig active state, socket mount parent,
  equipped item id, asset id, attack clip, and attack count.

## Scope And Likely Files

- Client:
  - `client/scripts/equipment_visuals.gd`
  - `client/scripts/first_person_equipment_rig.gd`
  - `client/scripts/main.gd`
  - `client/scripts/combat_local_attack_presentation.gd`
  - `client/scripts/player_camera_controller.gd`
  - `client/scripts/character_visual.gd`
- Shared presentation data:
  - `shared/assets/camera_presentations.v0.json`
  - `shared/assets/camera_presentations.v0.schema.json`
- Bot and tests:
  - `client/tests/test_item_visuals.gd`
  - `client/tests/test_camera_mode_settings.gd`
  - `tools/bot/scenarios/client/105_eye_view_weapon_presentation.json`
  - `docs/progress/scenario-movement-audit.tsv` if scenario metadata changes require it
- Docs:
  - `docs/CODEMAP.md`
  - `docs/plans/v467_2026-07-20-shared-first-person-equipment-rig.md`
  - `docs/as-built/v467_shared-first-person-equipment-rig.md`
  - `PROGRESS.md`
  - `docs/progress/slice-lifecycle.md`

## Test And Bot Proof

Focused proof:

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_item_visuals.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_camera_mode_settings.gd
HEADLESS=1 make bot-visual scenario=105_eye_view_weapon_presentation
make maintainability
```

Finish proof:

```bash
make ci
```

Manual visual proof:

```bash
make bot-visual scenario=105_eye_view_weapon_presentation
```

## Open Questions And Risks

| # | Question / Risk | Default |
|---|-----------------|---------|
| Q-1 | Full body meshes may clip the first-person camera. | Hide or heavily crop first-person rig body meshes while keeping sockets/equipment active; defer production hands. |
| Q-2 | Current clips may not read perfectly from first person. | Reuse the existing clip selection now; defer bespoke first-person animation authoring. |
| Q-3 | `main.gd` is over the maintainability target. | Keep edits minimal and put new rig logic in a focused file; if `main.gd` grows past allowance, extract or document per ratchet policy. |

## ADR Alignment

- ADR-0001 D2: preserves thin client/server authority; presentation only.
- ADR-0001 D7 and ADR-0006: reuses the manifest and mount-socket equipment pipeline.
- ADR-0007: animation remains client-only and uses existing authoritative combat events/input
  triggers as presentation signals.
