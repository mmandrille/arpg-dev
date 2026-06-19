# v284 Plan — Model Viewer Tool

Status: Ready for implementation
Goal: Add dynamic `make model-list` and `make model model=<asset_id>` workflows for inspecting
character and monster GLB presentation assets in an isolated Godot viewer.
Architecture: Keep catalog discovery in a small Python tooling module that reads existing manifest
and presentation JSON. Keep preview behavior in a focused Godot scene/script, reusing current
presentation scenes and `AnimationPlayer` conventions without touching gameplay `main.gd`. The tool
is offline client/tooling only: no server, database, protocol, replay, or bot scenario changes.
Tech stack: Python asset tooling/tests, Makefile targets, Godot client scene/script/tests,
existing shared asset JSON, docs.

## Baseline and shortcut decision

Builds on v283 with the model/animation work from v273-v279 already present. Reuse
`assets/manifests/assets.v0.json` as the asset identity source, `shared/assets/class_presentations.v0.json`
and `shared/assets/monster_visuals.v0.json` for `used_by` labels, and existing Godot presentation
scenes/animation libraries for runtime truth.

Asset/plugin decision:

- Adopt: committed GLB runtime assets, asset manifest, class presentation metadata, and monster
  presentation metadata.
- Borrow: existing Godot character/monster scenes, loaders, and `AnimationPlayer`/animation-library
  patterns.
- Reject: external assets, new Godot plugins, new runtime asset formats, network fetches, and
  Blender/export-pipeline changes.

Review gate note: `PROGRESS.md` says the post-v283 engineering review is due before the next feature
batch. This plan is ready, but implementation should either run `$review` first or proceed only with
explicit owner approval to bypass that gate for this tooling slice.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `tools/assets/model_catalog.py` | Discover selectable character/monster assets from manifest and presentation JSON; provide CLI list/resolve/check behavior. |
| Create | `tools/assets/test_model_catalog.py` | Unit coverage for dynamic discovery, `used_by` labels, item exclusion, and unknown asset failures. |
| Modify | `make/client.mk` | Add `model-list` and `model` make targets. |
| Create | `client/scenes/model_viewer.tscn` | Minimal isolated viewer scene with camera/light/model root and script attachment. |
| Create | `client/scripts/model_viewer.gd` | Load selected model/scene, frame it, auto-cycle available clips, and support headless check mode. |
| Create | `client/tests/test_model_viewer.gd` | Headless Godot checks for one character and one monster preview load path. |
| Modify | `scripts/client_smoke.sh` | Add a focused model-viewer unit gate under `client-unit`. |
| Modify | `CLAUDE.md` | Document the new developer commands if they are not obvious from `make help`. |
| Modify | `PROGRESS.md` | Update lifecycle/current status during `/finish`, not during implementation. |
| Create | `docs/as-built/v284_model-viewer-tool.md` | Record implementation proof and sample model commands during `/finish`. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [x] `client/scripts/main.gd` — must not be touched for this slice.
- [x] `server/internal/game/game_test.go` — not in scope.
- [x] `tools/bot/run.py` — not in scope.
- [x] `tools/validate_shared.py` — not in scope.
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:

- [x] Extract focused helper/module/test file as part of this slice: create `tools/assets/model_catalog.py`,
  `client/scripts/model_viewer.gd`, and focused tests instead of adding to large coordinators.
- [x] Defer extraction with rationale: not expected.

Verification:

```bash
make maintainability
```

## Task 1 — Dynamic model catalog

Files:

- Create: `tools/assets/model_catalog.py`
- Create: `tools/assets/test_model_catalog.py`

- [x] Step 1.1: Implement repo-root-relative JSON loading for `assets/manifests/assets.v0.json`,
  `shared/assets/class_presentations.v0.json`, and `shared/assets/monster_visuals.v0.json`.
- [x] Step 1.2: Build catalog rows keyed by manifest `asset_id`, limited to `type=character` and
  `type=monster` assets that are referenced by class or monster presentation data.
- [x] Step 1.3: Group `used_by` labels per asset ID; class labels are class IDs, monster labels are
  monster definition IDs. Preserve deterministic sort order for stable CLI output.
- [x] Step 1.4: Exclude equipment/item-only assets and fallback equipment assets.
- [x] Step 1.5: Add CLI commands for list and resolve/check. Unknown asset IDs must exit non-zero
  with a concise message that points to `make model-list`.
- [x] Step 1.6: Add Python unit tests for character inclusion, monster inclusion, multiple
  `used_by` labels, item exclusion, and unknown asset handling.

Verify:

```bash
.venv/bin/pytest tools/assets/test_model_catalog.py -q
```

## Task 2 — Godot viewer scene and script

Files:

- Create: `client/scenes/model_viewer.tscn`
- Create: `client/scripts/model_viewer.gd`

- [x] Step 2.1: Create an isolated `Node3D` viewer scene with camera, light, and a dedicated model
  root. It must run without server/database dependencies.
- [x] Step 2.2: Read selected asset ID from an environment variable or project setting passed by
  the make target. Read check/headless mode from a simple flag such as `CHECK=1`.
- [x] Step 2.3: Resolve selected asset ID through the Python-generated catalog data or by reading
  the same manifest/presentation JSON directly in GDScript. Avoid a parallel hand-maintained asset
  list.
- [x] Step 2.4: Prefer existing presentation scenes when they provide animation libraries. For
  character assets, use class presentation metadata and `client/scenes/character.tscn` where it
  preserves current animation setup. For monster assets, map through `monster_visuals.scene` to
  existing `client/scenes/monster_*.tscn`.
- [x] Step 2.5: Apply presentation scale/height metadata where available and frame the loaded model
  robustly under the camera.
- [x] Step 2.6: Discover available clips from the model's `AnimationPlayer`, auto-cycle common clips
  (`idle`, `walk`, `attack`, `hit`, `death`) when present, and skip missing optional clips cleanly.
- [x] Step 2.7: In check mode, instantiate the requested model, assert a preview root exists, assert
  an `AnimationPlayer` exists when expected for current presentation scenes, print a stable PASS
  sentinel, and exit.

Verify:

```bash
GODOT="$(GODOT)" MODEL_ASSET_ID=character_paladin_v0 MODEL_VIEWER_CHECK=1 godot --headless --path client --scene res://scenes/model_viewer.tscn
GODOT="$(GODOT)" MODEL_ASSET_ID=monster_tiny_flyer_v0 MODEL_VIEWER_CHECK=1 godot --headless --path client --scene res://scenes/model_viewer.tscn
```

## Task 3 — Make targets

Files:

- Modify: `make/client.mk`

- [x] Step 3.1: Add `model-list` target that invokes the Python catalog list command through the
  project virtualenv.
- [x] Step 3.2: Add `model` target that requires `model=<asset_id>`, validates the asset ID through
  the catalog helper before launching Godot, and passes the selected ID to the viewer.
- [x] Step 3.3: Support check mode, for example `make model model=monster_tiny_flyer_v0 CHECK=1`,
  using Godot headless mode and the viewer PASS sentinel.
- [x] Step 3.4: Ensure `make help` includes both targets with concise usage text.

Verify:

```bash
make model-list
make model model=monster_tiny_flyer_v0 CHECK=1
```

Manual visual verification command to document:

```bash
make model model=monster_tiny_flyer_v0
```

## Task 4 — Client unit gate

Files:

- Create: `client/tests/test_model_viewer.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Step 4.1: Add a headless GDScript unit test that loads the viewer support code or scene for
  `character_paladin_v0` and `monster_tiny_flyer_v0`.
- [x] Step 4.2: Assert both selected assets resolve from manifest/presentation data and instantiate
  a visible preview root.
- [x] Step 4.3: Assert current presentation scenes expose an `AnimationPlayer` and at least one
  previewable animation clip.
- [x] Step 4.4: Add the test to `scripts/client_smoke.sh` with a stable sentinel so `make client-unit`
  covers the model viewer.

Verify:

```bash
make client-unit
```

## Task 5 — Docs and lifecycle

Files:

- Modify: `CLAUDE.md`
- Create during finish: `docs/as-built/v284_model-viewer-tool.md`
- Modify during finish: `PROGRESS.md`
- Existing: `docs/specs/v284_spec-model-viewer-tool.md`
- Existing: `docs/plans/v284_2026-06-19-model-viewer-tool.md`

- [x] Step 5.1: Add concise command docs near the existing command list:
  `make model-list`, `make model model=<asset_id>`, and optional `CHECK=1`.
- [x] Step 5.2: Record item/equipment preview as deferred in as-built notes, not as implementation
  scope.
- [x] Step 5.3: During `/finish`, update `PROGRESS.md` current status/lifecycle and create the
  as-built proof file with exact focused verification commands and manual visual command.

Verify:

```bash
rg -n "model-list|model model" CLAUDE.md docs/as-built/v284_model-viewer-tool.md PROGRESS.md
```

## Task 6 — Bot scenarios

No bot scenario is required. This slice does not touch gameplay, protocol, world presets, combat,
inventory, movement, transitions, or replay. Verification is Python tooling plus Godot headless and
manual visual preview.

## Task 7 — Final verification

- [x] `make maintainability`
- [x] `make validate-assets`
- [x] `.venv/bin/pytest tools/assets/test_model_catalog.py -q`
- [x] `make model-list`
- [x] `make model model=monster_tiny_flyer_v0 CHECK=1`
- [x] `make client-unit`
- [x] `make ci`

Deferred scope:

- Equipment/item preview, mounted weapons, paper-doll previews, and model import/export pipeline
  changes remain deferred to later slices.
- The post-v283 `$review` gate remains a process prerequisite unless explicitly waived before
  implementation.
