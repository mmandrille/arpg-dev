# v284 As Built: Model Viewer Tool

Date: 2026-06-19
Spec: [`docs/specs/v284_spec-model-viewer-tool.md`](../specs/v284_spec-model-viewer-tool.md)
Plan: [`docs/plans/v284_2026-06-19-model-viewer-tool.md`](../plans/v284_2026-06-19-model-viewer-tool.md)

## What shipped

- Added `make model-list`, backed by `tools/assets/model_catalog.py`, to list previewable character
  and monster model asset IDs dynamically from the asset manifest plus class/monster presentation
  metadata.
- Added `make model model=<asset_id>` to launch an isolated Godot model viewer for a single
  character or monster asset.
- Added `CHECK=1` support for headless verification, for example:

```bash
make model model=monster_tiny_flyer_v0 CHECK=1
```

- Added `client/scenes/model_viewer.tscn` and `client/scripts/model_viewer.gd`, reusing existing
  character/monster presentation scenes and their `AnimationPlayer` libraries.
- Added Python and Godot tests so both the dynamic catalog and preview scene are covered by local
  gates.

## Proof

Focused verification:

```bash
make maintainability
make validate-assets
.venv/bin/pytest tools/assets/test_model_catalog.py -q
make model-list
make model model=monster_tiny_flyer_v0 CHECK=1
make client-unit
```

Full verification:

```bash
make ci
```

Result: green on rerun, 2026-06-19. The first CI attempt hit a transient
`mercenary_recovery_ui` client bot timeout; the focused rerun passed, and the subsequent full
`make ci` completed successfully.

## Manual visual command

```bash
make model model=monster_tiny_flyer_v0
```

## Deferred

- Equipment/item previews remain deferred.
- Mounted weapon and paper-doll model previews remain deferred.
- New assets, external asset plugins, and Blender/export-pipeline changes remain deferred.
