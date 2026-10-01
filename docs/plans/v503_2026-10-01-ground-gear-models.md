# v503 Plan — Ground gear models at game scale

- **Status:** Complete; integrated and combined batch `make ci` passed in 11m41s.

- **Visual limit:** Player-camera gear readability remains limited by a very dark single-frame capture.
- **Spec:** [`v503_spec-ground-gear-models.md`](../specs/v503_spec-ground-gear-models.md)
- **Goal:** Tune the existing manifest-backed armor and jewelry ground models so all equipment families read clearly at the player camera scale, using schema-backed ground-pose data.
- **Baseline:** `5365832b9029e0d9e178d57f2a95a6a697b01acc` (detached HEAD in the assigned worktree).
- **Prerequisites:** None for spec/plan or implementation. v487 ground-model resolution is present in the baseline. Renderer use was serialized with sibling visual work and has now been released.
- **Sibling ordering:** v504 and v508 may spec/plan now. Integrate this slice before their implementation; v504 may extend the item-to-ground-model surface, while v508 owns rarity treatment. v507 may implement independently unless it touches this plan's test/scenario files.

## Review findings

- The seven existing ground gear families already point to GLBs through
  `shared/assets/item_presentations.v0.json` and `assets/manifests/assets.v0.json`.
- `LootNodeFactory` already consumes the schema-backed `equipment_display.v0.json` pose map for
  ground models. There is no need to add runtime code, a new loader, or a new asset format.
- Existing fallback model assets are the documented v487 choice because the KayKit character packs
  do not provide detachable body armor meshes. This slice borrows those existing assets and adds no
  download, dependency, or manifest edit.
- The baseline `floor-item` captures from v487 are available under
  `docs/as-built/assets/v487/`. They show model silhouettes in a floor-item preview. That screenshot
  suite discovers items from equipped visuals and omits rings and amulets, so execution must capture
  those directly with `render_focus.py` and additionally capture the existing `equipment_lab` loot
  from the player camera before and after tuning.
- The security skill marks pure presentation as out of scope: no API, authorization, database,
  external input, or HTML rendering changes are planned.

## File map and ownership

| Action | Path | Responsibility |
|---|---|---|
| Modify | `shared/assets/equipment_display.v0.json` | Per-asset ground scale/rotation/floor-rest values for the existing armor and jewelry models. Do not change shared rarity settings or gameplay values. |
| Modify | `client/tests/test_loot_node_factory.gd` | Catalog- and manifest-derived gear model resolution and pose assertions. |
| Modify | `client/tests/test_item_visuals.gd` | Integration assertion that representative dropped gear uses a visible manifest-backed model. |
| Modify | `tools/bot/scenarios/client/10_full_equipment.json` | Capture the initial `equipment_lab` loot at game camera scale before the scenario picks up items; mark it `skip_if_headless: true`. |
| Create | `docs/as-built/v503_ground-gear-models.md` | Implementation evidence and limits. |
| Create | `docs/as-built/assets/v503/` | Representative focused and player-camera before/after captures. |

No production GDScript, schema, asset, manifest, world, server, or protocol changes are expected.
If an implementation finds that code changes are essential, update the spec and request coordinator
review before expanding scope.

## Ordered tasks

### 1. Capture baseline at player camera scale

- [x] Add one named `capture_frame` step to `client_full_equipment` after it confirms the initial
  ground-loot count and before the first pickup. Use `skip_if_headless: true`; do not add movement.
- [x] Run `make bot-visual scenario=client_full_equipment` on the untouched presentation data and
  save the emitted frame under `docs/as-built/assets/v503/` as the baseline capture.
- [x] Run `make regen-screenshots SUITE="floor-item"` before tuning. Inspect helm, chest, belt,
  boots, and other captured armor output. Capture `ring` and `amulet` individually because they are
  excluded from the screenshot catalog's equipped-visual item discovery.

```bash
make bot-visual scenario=client_full_equipment
make regen-screenshots SUITE="floor-item"
python3 skills/showme/scripts/render_focus.py --focus floor-item --items ring
python3 skills/showme/scripts/render_focus.py --focus floor-item --items amulet
```

### 2. Tune ground poses in the shared catalog

- [x] Add `ground_pose.assets` entries for the existing fallback gear assets, deriving IDs from the
  current manifest. Include only the ring asset referenced by the ground presentation catalog;
  leave the unused right-ring asset untouched.
- [x] Tune scale, rotation, and floor-rest/centering only from inspected captures. Keep the ring and
  amulet recognizable at the game camera without making them dominate the ground loot.
- [x] Leave `item_presentations.v0.json`, `item_visuals.v0.json`, rarity color/beam/glow settings,
  asset files, and manifest entries unchanged.

```bash
make validate-shared
make validate-assets
```

- [x] `make validate-shared` passed (2,208 checks).
- [x] `make validate-assets` passed (451 checks).

### 3. Extend focused coverage

- [x] Add catalog-derived client checks for every gear family: the model asset exists in the
  manifest, is instantiated under the expected ground-model node name, and follows the effective
  pose from `EquipmentDisplayLoader.ground_pose_for`.
- [x] Assert the tuned model bounds meet the floor-rest tolerance and remain within the model's
  configured size cap. Keep expected scale values sourced from JSON rather than copied into test
  literals.
- [x] Keep the existing loot collider/pickup proof intact; do not change collision dimensions.

```bash
godot --headless --path client --script res://tests/test_loot_node_factory.gd
godot --headless --path client --script res://tests/test_item_visuals.gd
make bot-client SCENARIO=client_full_equipment HEADLESS=1
```

### 4. Inspect after captures and handoff evidence

- [x] Re-run `make regen-screenshots SUITE="floor-item"` and capture the `ring` and `amulet`
  focuses individually. Inspect jewelry and armor PNGs, and copy representative before/after
  samples into `docs/as-built/assets/v503/`.
- [x] Re-run `make bot-visual scenario=client_full_equipment` and inspect the player-camera capture
  at the original view. The dark frame limits its value for confirming family silhouettes.
- [x] Write the as-built summary with verification results, camera/renderer limits, and exact bot
  command. Report every ignored capture artifact that the coordinator needs to preserve.
- [x] Run `make maintainability` and `git diff --check`. Do not run `make ci` or `make ci-full` in
  this batch slice session.

## Verification and handoff gate

The focused proof and before/after images are ready for coordinator review. The player-camera
capture is dark enough that gear-family recognition at that distance remains a visual limitation;
the frame confirms placement but does not conclusively establish silhouette recognition under
normal gameplay lighting. The coordinator runs combined `make ci` after accepted slices are
integrated; this plan has no per-slice full-CI task.

At handoff report the recorded base SHA, dependencies used, worktree dirty state, every changed,
deleted, and untracked path, ignored image evidence to retain, commands and outcomes, visual
captures, acceptance gaps, and likely shared-file overlap. Do not commit, push, run `/finish`, or
modify the coordinator's main checkout.
