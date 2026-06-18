# v260 Plan - Biome Color Palettes

Status: Complete
Goal: Add shared depth-band biome palettes and apply them to client dungeon floor/wall textures.
Architecture: Store palette bands in `dungeon_generation.v0.json` with schema validation. Keep the
client renderer code-native: `GroundWallFactory` reads shared data and produces procedural textures
tinted by selected palette; `WallRenderer` supplies current level context. No server or protocol
changes.
Tech stack: Shared JSON/schema, Godot GDScript client, client unit tests, docs.

## Baseline and Shortcut Decision

Builds on v254-v255 dungeon readability and the existing `GroundWallFactory` procedural texture
system. This slice adds data-driven colors only; it does not add a new asset pipeline.

Asset/plugin decision: reject external assets, imported texture packs, shader plugins, and Godot
addons. Borrow existing code-native procedural ground/wall texture generation.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/dungeon_generation.v0.json` | Add biome palette bands |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate biome palette shape and hex colors |
| Modify | `client/scripts/ground_wall_factory.gd` | Load/select palettes and tint generated textures |
| Modify | `client/scripts/wall_renderer.gd` | Pass current level to wall material generation |
| Modify | `client/scripts/main.gd` | Keep wall renderer level context current |
| Modify | `client/tests/test_item_visuals.gd` | Prove palette selection and visual variation |
| Modify | `client/tests/test_factories.gd` | Prove wall renderer accepts level palette context |
| Modify | `docs/specs/v260_spec-biome-color-palettes.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v260 lifecycle row |
| Add | `docs/as-built/v260_biome-color-palettes.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and next selected autoloop item |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none planned
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Keep palette logic out of `main.gd`; add only a compact wall-renderer level-context sync.

Verification:
```bash
make maintainability
```

## Task 1 - Shared Palette Data

Files:
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`

- [x] Step 1.1: Add `biome_palettes` to required schema properties.
- [x] Step 1.2: Define at least two depth bands: shallow cave and deep vault.
- [x] Step 1.3: Validate stable ids, non-overlapping min/max shape, and hex color fields.

```bash
make validate-shared
```

## Task 2 - Client Palette Application

Files:
- Modify: `client/scripts/ground_wall_factory.gd`
- Modify: `client/scripts/wall_renderer.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_item_visuals.gd`
- Modify: `client/tests/test_factories.gd`

- [x] Step 2.1: Load `dungeon_generation.v0.json` in `GroundWallFactory`.
- [x] Step 2.2: Select palette by absolute dungeon depth, falling back safely when data is missing.
- [x] Step 2.3: Use palette colors in ground and wall procedural texels.
- [x] Step 2.4: Keep wall renderer level context synchronized.
- [x] Step 2.5: Extend client tests for shallow/deep palette differences and wall renderer context.

```bash
make client-unit
```

## Task 3 - Lifecycle Docs

Files:
- Modify: `docs/specs/v260_spec-biome-color-palettes.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v260_biome-color-palettes.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Mark the v260 spec complete.
- [x] Step 3.2: Add v260 lifecycle and as-built notes.
- [x] Step 3.3: Update `PROGRESS.md` current status and leave generated doors, doorway LOS, and
  quest marker work as remaining selected autoloop scope.

```bash
make maintainability
```

## Final Verification

- [x] `make validate-shared`
- [x] `make client-unit`
- [x] `make maintainability`
- [ ] Autoloop final batch gate: `make ci`

Manual visual proof, if desired:

```bash
make bot-visual scenario=14_dungeon_wall_rendering
```
