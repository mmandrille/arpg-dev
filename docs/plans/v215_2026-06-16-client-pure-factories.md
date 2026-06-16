# v215 Plan — Client Pure Factory Extraction

Status: Ready for implementation
Goal: Extract pure constants, ground/wall rendering, and loot-node construction out of `client/scripts/main.gd` without changing presentation output.
Architecture: This is a client-only maintainability slice. New Godot `class_name` helper scripts own pure factory code and are directly preloadable by tests. `main.gd` remains the scene coordinator and delegates to helpers through narrow constructor parameters.
Tech stack: Godot client, maintainability ratchet, headless Godot tests.

## Baseline and shortcut decision

Builds on v213 baseline plus the user-supplied v215 spec. Client assets/plugins decision: **adopt existing in-repo primitive mesh and manifest pipeline; reject external assets/plugins** because this slice only moves existing placeholder rendering code.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/client_constants.gd` | Compile-time constants formerly embedded in `main.gd`. |
| Add | `client/scripts/ground_wall_factory.gd` | Ground materials/textures and wall texture cache. |
| Add | `client/scripts/wall_renderer.gd` | Static/preset/generated wall normalization and scene-tree rendering. |
| Add | `client/scripts/loot_node_factory.gd` | Loot labels, primitive loot meshes, and ground item model loading. |
| Add/Modify | `client/tests/test_factories.gd` | Direct preload proof for new helpers without `main.gd`. |
| Modify | `client/scripts/main.gd` | Delegate extracted domains and remove inline helpers. |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `main.gd` baseline to actual post-extraction count. |
| Add | `docs/as-built/v215_client-pure-factories.md` | Record shipped proof and verification. |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Lifecycle consolidation after verification. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 — Extract pure client constants

Files:
- Create: `client/scripts/client_constants.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Move gameplay/presentation constants from `main.gd` into `ClientConstants`.
- [x] Step 1.2: Replace `main.gd` references with `ClientConstants.*` and keep script preloads local.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

## Task 2 — Extract ground and wall factories

Files:
- Create: `client/scripts/ground_wall_factory.gd`
- Create: `client/scripts/wall_renderer.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Move ground material/texture helpers into `GroundWallFactory`.
- [x] Step 2.2: Move wall normalization/rendering into `WallRenderer`, with a `GroundWallFactory` constructor dependency for wall texture generation.
- [x] Step 2.3: Keep `current_wall_layout` authoritative in `main.gd` by assigning the renderer return value after each render call.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

## Task 3 — Extract loot node factory

Files:
- Create: `client/scripts/loot_node_factory.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 3.1: Move loot node, primitive mesh, label, item definition, and runtime path helpers into `LootNodeFactory`.
- [x] Step 3.2: Pass only `asset_manifest`, `item_presentations`, and a model tint callable into the factory; do not pass `main.gd`.
- [x] Step 3.3: Keep loot label reveal/filter state in `main.gd`.
```bash
godot --headless --path client --script res://tests/test_factories.gd
```

## Task 4 — Client regression proof

Files:
- Modify: `client/tests/test_factories.gd`

- [x] Step 4.1: Add direct preload tests for each helper.
- [x] Step 4.2: Run existing client smoke proof.
```bash
make client-smoke
```

## Task 5 — Lifecycle docs and final verification

Files:
- Modify: `.maintainability/file-size-baseline.tsv`
- Add: `docs/as-built/v215_client-pure-factories.md`
- Modify: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`

- [x] Step 5.1: Lower the maintainability baseline for `client/scripts/main.gd`.
- [x] Step 5.2: Record as-built proof and lifecycle status.
- [x] Step 5.3: Run final standalone gates.
```bash
make maintainability
make client-unit
make client-smoke
make ci
```

## Final verification

- [x] `godot --headless --path client --script res://tests/test_factories.gd`
- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make client-smoke`
- [x] `make ci`

