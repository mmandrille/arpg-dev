---
name: 3dmodel
description: Integrate GLB/glTF 3D models into the arpg-dev Godot client as monster, companion, summon, boss, or prop visual replacements. Use when a user provides a 3D model file and asks to use it in-game, replace an existing visual, make it walk/attack, fix model orientation/scale, or create a repeatable model-import workflow.
---

# 3D Model Integration

Use this skill to turn a supplied `.glb`/`.gltf` into a working arpg-dev visual. Keep gameplay authority on the server; most model work is client presentation plus shared visual metadata.

## Start Here

1. Read project baseline first: `PROGRESS.md` current status/open gaps/checklist, `CLAUDE.md`, `docs/CODEMAP.md`, and relevant ADRs:
   - `docs/adr/0001-technology-stack.md`
   - `docs/adr/0006-asset-pipeline.md`
   - `docs/adr/0007-animation-state-model.md`
2. Inspect the requested source and target:
   - Source model path, e.g. `assets/monsters/wolf/wolf.glb`.
   - Replacement target, e.g. ranger summon `black_wolf_companion`, monster `dungeon_wolf`, boss model, or prop/interactable.
3. Prefer the existing manifest/data path. Do not hardcode model paths in gameplay code.
4. Record an adopt/borrow/reject decision in the spec or plan when the task is client-side work using outside assets.

## Probe A Model Before Integrating

Run the bundled sandbox tool from repo root:

```bash
python3 skills/3dmodel/scripts/create_model_probe.py \
  --model assets/monsters/wolf/wolf.glb \
  --key wolf \
  --yaw-degrees -90
```

The tool:

- Inspects the GLB JSON chunk: bounds, nodes, skins, embedded animations.
- Copies the model to `client/assets/_model_probe/<key>/`.
- Creates `client/scenes/model_probe_<key>.tscn`.
- Creates `client/animations/model_probe_<key>_anims.tres` with basic `idle`, `walk`, `attack`, and `death` clips.
- Runs a headless Godot sanity test when `GODOT` is available.

Use this as a disposable import/orientation check. After learning the right yaw/scale/animation approach, port the result into canonical files and delete or ignore the `_model_probe` sandbox files.

## Canonical Integration Workflow

1. Copy runtime bytes under `client/assets/...`.
   - Keep the original source under `assets/...` if supplied by the user.
   - Runtime path must be under the Godot project, e.g. `client/assets/monsters/wolf/wolf.glb`.
   - Let Godot import once so `.import` files and extracted embedded textures appear.
2. Register the asset in `assets/manifests/assets.v0.json`.
   - Add `type`, `source_path`, `runtime_path`, `format: "glb"`, `scale_unit: "meters"`, `required_nodes`, and `provenance`.
   - Use `required_nodes: []` for static unskinned meshes; do not fake skin joints.
   - Use `shasum -a 256 <runtime-glb>` for provenance.
3. Wire shared visual metadata.
   - Monsters/companions: `shared/assets/monster_visuals.v0.json` and schema enum in `monster_visuals.v0.schema.json`.
   - Class/player model: `shared/assets/class_presentations.v0.json`.
   - Equipment: `shared/assets/item_visuals.v0.json` plus socket/mount transforms.
   - Skill summon visual model: `shared/rules/skills.v0.json` `companion.visual_model` and `companion.visual_scale`.
4. Add a Godot scene.
   - Create `client/scenes/monster_<key>.tscn` or matching target scene.
   - Use a correction parent for imported models:

```text
MonsterKey (Node3D)
  ModelRoot (Node3D)          # owns fixed yaw/scale/axis correction
    Model (instance GLB)      # imported model
  AnimationPlayer             # root_node = ".."
```

   - Animate `ModelRoot/Model`, not `ModelRoot`, so walk/attack clips do not overwrite the fixed yaw correction.
5. Register the scene in client code.
   - Add preload and match arm in `client/scripts/main.gd::_monster_scene_for_visual`.
   - Add `monster_visuals_loader.gd` allowlist entry if the visual model can arrive over protocol.
6. Keep server changes minimal.
   - For summons, change only rules data if possible.
   - If code couples visual size to gameplay scaling, decouple it: combat stats and `visual_scale` are separate concerns.
7. Update tests and bot scenarios.
   - Client: `client/tests/test_animation.gd` should instantiate the scene, require `idle/walk/hit/death` or `attack` as applicable, and assert walk does not reset yaw correction.
   - Server: target skill/monster test should assert configured `visual_model` and `visual_scale`.
   - Bot scenario: update `visual_model` expectations for the target.

## Orientation Rules Learned From The Wolf

- arpg-dev facing helpers assume visual front is parent local `+Z`.
- Imported GLBs may use `+X`, `-X`, or another axis as their nose/front.
- Put the permanent axis correction on `ModelRoot`.
- Put generated walk/attack bob/roll tracks on the child imported model (`ModelRoot/Model`).
- If the model walks sideways, adjust `ModelRoot.rotation.y` by 90-degree increments.
- If it is aligned but walks backward, flip `ModelRoot.rotation.y` by 180 degrees.
- Add a test that plays `walk`, seeks into the clip, and confirms `ModelRoot.rotation.y` is unchanged.

## Verification Checklist

Run the smallest meaningful set first:

```bash
make validate-shared
make validate-assets
cd server && go test ./internal/game -run '<target test>'
GODOT=/opt/homebrew/bin/godot make client-unit
make bot scenario=<target_scenario>
GODOT=/opt/homebrew/bin/godot GODOT_FLAGS=--headless make bot-visual scenario=<target_scenario>
```

For visual work, always tell the user the exact manual replay command, usually:

```bash
make bot-visual scenario=<target_scenario>
```

## Common Failure Modes

- **Scene loads, but model is invisible:** missing runtime copy under `client/assets`, missing `.import`, or extracted texture not committed.
- **Faces target but appears side-on:** model-local forward axis is wrong; adjust `ModelRoot.rotation.y`.
- **Looks correct idle, wrong while walking:** animation track is overwriting the correction node; animate a child node instead.
- **Scale data has no effect:** server may override visual scale from gameplay stat scaling; decouple `visual_scale`.
- **Asset validation fails on joints:** static meshes should declare `required_nodes: []`; only rigged/skinned assets should list required skin joints.
