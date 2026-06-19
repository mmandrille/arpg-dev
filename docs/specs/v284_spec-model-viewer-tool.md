# v284 Spec: Model Viewer Tool

Status: Draft
Date: 2026-06-19
Codename: `model-viewer-tool`

## Purpose

Add a developer-facing mini tool for inspecting committed 3D character and monster models without
starting a gameplay session. The tool should make the current runtime model catalog discoverable
from repository data and open one selected model in an isolated Godot preview scene that cycles its
available animations.

The intended workflows are:

```bash
make model-list
make model model=<asset_id>
```

`make model-list` must be dynamic. It should derive rows from `assets/manifests/assets.v0.json`,
`shared/assets/class_presentations.v0.json`, and `shared/assets/monster_visuals.v0.json`, not from a
separate hardcoded list. Rows should use manifest asset IDs as selectable names and include the
runtime path plus known in-game usage labels, for example:

```text
character_paladin_v0  client/assets/characters/paladin/paladin.glb  used_by=paladin
monster_tiny_flyer_v0 client/assets/monsters/tiny_flyer/monster_tiny_flyer.glb used_by=dungeon_bat
```

`make model model=<asset_id>` should launch a focused Godot viewer that loads the selected character
or monster model alone, frames it under the existing camera/art direction, and cycles supported
animation clips such as `idle`, `walk`, `attack`, `hit`, and `death` when those clips exist.

## Non-goals

- Do not preview equipment, inventory items, mounted weapons, or paper-doll combinations in this
  slice. Item/equipment model inspection is deferred to a separate slice.
- Do not add new GLB assets, source art, textures, external plugins, or external asset pipelines.
- Do not change combat, movement, AI, server state, persistence, or realtime protocol behavior.
- Do not retarget broken animation tracks or redesign animation libraries. Missing clips should be
  reported or skipped cleanly.
- Do not build a production in-game model browser UI.
- Do not require a database, server, auth token, or gameplay session to inspect a model.

## Acceptance Criteria

- `make model-list` prints every selectable character and monster asset ID declared in the asset
  manifest and referenced by class or monster presentation data.
- `make model-list` includes each row's runtime GLB path and at least one `used_by=` label when the
  asset is referenced by `class_presentations.v0.json` or `monster_visuals.v0.json`.
- `make model-list` excludes equipment/item-only assets for this slice.
- `make model model=<asset_id>` accepts manifest asset IDs such as `character_paladin_v0` and
  `monster_tiny_flyer_v0`.
- `make model model=<asset_id>` fails fast with a clear message for an unknown asset ID and points
  the user to `make model-list`.
- The viewer loads the selected model or existing presentation scene in isolation, with no server or
  database required.
- The viewer auto-cycles available clips and visibly exercises locomotion and reaction animations
  when the selected model's `AnimationPlayer` exposes them.
- Missing optional clips do not crash the viewer; they are skipped or surfaced as concise status
  output.
- A headless/check mode can verify that at least one character model and one monster model resolve,
  instantiate, and expose a usable preview root.

## Scope And Likely Files

Contracts and data:

- `assets/manifests/assets.v0.json` remains the source of truth for `asset_id -> runtime_path`.
- `shared/assets/class_presentations.v0.json` supplies character usage labels such as `paladin`.
- `shared/assets/monster_visuals.v0.json` supplies monster usage labels such as `dungeon_bat`.
- No shared protocol or gameplay schema bump is expected.

Tooling:

- `Makefile` or `make/subfiles.mk` for `model-list` and `model` targets.
- A focused Python helper under `tools/assets/` or equivalent for dynamic catalog discovery,
  validation, and command-line output.
- Tests near the helper, likely under `tools/assets/`.

Client:

- A new isolated Godot scene/script such as `client/scenes/model_viewer.tscn` and
  `client/scripts/model_viewer.gd`, or an equivalent headless-friendly viewer entrypoint.
- Existing model presentation sources should be reused where practical:
  `client/scenes/character.tscn`, `client/scenes/monster_*.tscn`,
  `client/scripts/class_presentations_loader.gd`, `client/scripts/monster_visuals_loader.gd`, and
  existing `client/animations/*.tres` libraries.
- Avoid adding model-viewer responsibilities to `client/scripts/main.gd`.

Docs:

- Add or update concise developer command documentation if the new make targets are not self-evident
  from `make help`.
- As-built notes should record the exact sample models verified manually or headlessly.

## Test And Bot Proof

Expected focused checks:

- Python unit coverage for catalog discovery:
  - includes character and monster assets referenced by presentation data;
  - groups multiple `used_by` labels for shared assets such as dummy monsters;
  - excludes equipment/item assets;
  - rejects unknown model IDs with actionable output.
- Godot headless/check coverage for model instantiation:
  - one character asset, for example `character_paladin_v0`;
  - one monster asset, for example `monster_tiny_flyer_v0`;
  - verifies an `AnimationPlayer` or equivalent preview animation surface is found when expected.
- Run `make validate-assets` to prove the manifest/runtime GLB baseline still resolves.

Bot proof:

- No protocol bot scenario is required because this is offline client tooling, not gameplay or
  network behavior.
- Visual verification command to document for humans/agents:

```bash
make model model=monster_tiny_flyer_v0
```

If the implementation adds a headless check flag, document the exact command in the plan/as-built,
for example:

```bash
make model model=monster_tiny_flyer_v0 CHECK=1
```

## Asset And Plugin Decision

- Adopt: existing committed GLB runtime assets, `assets/manifests/assets.v0.json`, and the existing
  shared class/monster presentation metadata.
- Borrow: existing Godot presentation scenes, loaders, and `AnimationPlayer`/`AnimationController`
  conventions from the client.
- Reject: new external assets, new Godot plugins, new asset formats, network asset fetching, or a
  Blender/export-pipeline change for this slice.

## Open Questions And Risks

- The post-v283 engineering review is due per `PROGRESS.md`. Planning should either run that review
  first or explicitly confirm this tooling slice is allowed to proceed before the review gate.
- Some runtime GLBs are loaded through presentation scenes with animation libraries, while raw GLB
  asset IDs may not carry the same `AnimationPlayer` setup by themselves. The plan should choose the
  least duplicated mapping from asset ID to preview scene and keep it dynamic enough to avoid a
  parallel hardcoded catalog.
- Existing model scales vary, for example character presentation scale differs per class. The viewer
  should use presentation metadata where available and keep framing robust across unusually scaled
  models.
- Animation availability differs by model/profile. The viewer should treat the visible set of clips
  as model-owned facts, not as a contract that every model must support every clip.
