# v215 — Client Pure Factory Extraction

**Status:** Draft  
**Date:** 2026-06-16  
**Codename:** client-pure-factories

---

## Purpose

Extract four self-contained domains out of `main.gd` into standalone helper files:

1. **`ClientConstants`** — all compile-time `const` declarations.
2. **`GroundWallFactory`** — procedural ground/wall texture and material generation.
3. **`WallRenderer`** — wall node construction and scene-tree management for dungeon walls.
4. **`LootNodeFactory`** — loot node and ground-equipment visual construction.

Each extracted file must be importable and unit-testable without importing `main.gd`. No
`helpers=globals()` or equivalent. `main.gd` becomes a consumer of the new helpers.

**Maintainability obligation:** `main.gd` is currently 6603 lines, baseline 6521 — already
82 lines over the allowed +25 drift. This slice must reduce it to ≤ 6521 and lower the
`.maintainability/file-size-baseline.tsv` entry accordingly.

Expected line savings: ~415 lines → `main.gd` ≈ 6188 after extraction.

---

## Non-goals

- No change to any game logic, rendering output, or observable behaviour.
- Tier 2 context-object extractions (BossVisualsController, TownNodeFactory) are deferred to v216.
- No new Godot autoloads. All new files use `class_name … extends RefCounted` (or Node where a
  scene tree is required).
- No changes to Go server, shared rules, or Python tools.

---

## Acceptance criteria

1. `make ci` green after extraction.
2. `main.gd` line count ≤ 6521; `.maintainability/file-size-baseline.tsv` entry updated to the
   new actual count (or removed if ≤ 600).
3. Each new file is ≤ 300 lines and **not** in the grandfathered baseline (i.e., they start below
   600 lines).
4. No `main.gd` symbol is accessible from any new file except through `ItemRulesLoader` (existing
   singleton) and constructor parameters.
5. A headless unit test can `preload` each new file directly and exercise its functions without
   `main.gd` in scope. (Existing `client-unit` or a new `test_factories.gd` can cover this.)
6. Rendering output in `make bot-visual` is visually unchanged (loot nodes, walls, ground).

---

## Scope and files touched

### New files

| File | `class_name` | Contents |
|------|-------------|----------|
| `client/scripts/client_constants.gd` | `ClientConstants` | All `const` from `main.gd` lines 59–119 and 282–295 (colors, ranges, scene keys, flow names, event clip maps, camera params, timing budgets) |
| `client/scripts/ground_wall_factory.gd` | `GroundWallFactory` | `make_ground_node`, `update_ground_material`, `ground_texture_id_for_level`, `ground_material_for_level`, `make_ground_texture`, `ground_texel`, `make_wall_texture`, `wall_texel`; holds `ground_textures`/`wall_textures` cache dicts internally |
| `client/scripts/wall_renderer.gd` | `WallRenderer` | `render_world_walls`, `render_wall_layout`, `clear_wall_nodes`, `normalized_wall_view`, `make_wall_node`; constructor takes `walls_root: Node3D` |
| `client/scripts/loot_node_factory.gd` | `LootNodeFactory` | `make_loot_node`, `add_loot_primitive`, `make_ground_equipment_model`, `ground_item_tint`, `add_loot_rarity_background`, `add_loot_label`, `add_loot_box`, `add_loot_cylinder`, `add_loot_mesh`, `loot_color`, `loot_label_color`, `loot_label_text`, `item_rarity_background`, `item_definition`, `generic_loot_name`, `res_path`; constructor takes `asset_manifest: Dictionary, resolver: EquipmentVisualResolver` |

### Modified files

| File | Change |
|------|--------|
| `client/scripts/main.gd` | Replace inline `const` with `ClientConstants.FOO`; delegate ground/wall texture calls to `_ground_factory`; delegate wall rendering to `_wall_renderer`; delegate loot node construction to `_loot_factory`; add `var _ground_factory: GroundWallFactory`, `var _wall_renderer: WallRenderer`, `var _loot_factory: LootNodeFactory` initialized in `_ready`; remove extracted functions |
| `.maintainability/file-size-baseline.tsv` | Lower `client/scripts/main.gd` entry to new actual line count |

### Unchanged

- All Go server files.
- All shared rules/protocol/golden files.
- All Python tools.
- All other client scripts (no public API changes; `class_name` declarations are automatically
  visible, so no import statements needed in callers).

---

## Implementation notes

### ClientConstants

All callers of bare const names in `main.gd` change to `ClientConstants.FOO`. Because
`class_name ClientConstants` is globally visible, no `preload` is needed in `main.gd`. The
existing `preload`-based `const` aliases at lines 5–54 (script handles, not game constants) stay
in `main.gd`.

### GroundWallFactory

Constructor signature: `GroundWallFactory.new()` — no params. Holds `ground_textures: Dictionary`
and `wall_textures: Dictionary` as instance vars (cache). Called via:

```gdscript
var _ground_factory := GroundWallFactory.new()
# ...
ground_node = _ground_factory.make_ground_node(dungeon_generation, current_level)
_ground_factory.update_ground_material(ground_node, dungeon_generation, current_level)
```

### WallRenderer

Constructor: `WallRenderer.new(walls_root: Node3D)`. `render_world_walls` and
`render_wall_layout` delegate to `_ground_factory` for wall textures — WallRenderer takes
`_ground_factory` as a second constructor param, or alternatively `render_world_walls` accepts
the `wall_textures` dict as a param. Plan decides.

### LootNodeFactory

Constructor: `LootNodeFactory.new(asset_manifest: Dictionary, resolver: EquipmentVisualResolver)`.

`res_path` (currently `_res_path` in main.gd at line 5510) is a two-line pure helper used only
by loot/entity visual code — move it here (or to a shared utility if v216 also needs it).

The `_loot_filter` var (line 170) stays in `main.gd` — it drives label visibility logic that
is not factory code.

`_set_pickable` (line 3166) is NOT called from factory functions — it is called from interaction
state code and stays in `main.gd`.

---

## Test and bot proof

- **Headless unit:** Add `client/tests/test_factories.gd` exercising `ClientConstants` key
  lookups, `GroundWallFactory.make_ground_texture` for at least two texture IDs, and
  `LootNodeFactory.loot_color` for a known item def. No scene tree required.
- **Smoke:** `make client-smoke` must pass (no regression in existing headless tests).
- **Bot:** `make bot` must pass. Loot pickup, equip, and movement scenarios exercise the loot
  node factory indirectly.
- **Visual replay:** `make bot-visual` must complete without visual regression (eyeball check:
  ground tiles, wall meshes, loot nodes unchanged).

---

## Open questions and risks

1. **WallRenderer → GroundWallFactory coupling.** `render_wall_layout` calls `make_wall_texture`
   for cached wall textures. Options: (a) WallRenderer takes `GroundWallFactory` as a ctor param,
   (b) WallRenderer receives a `wall_textures` dict from main.gd, (c) wall texture generation
   moves into WallRenderer. Plan must pick one.

2. **`_res_path` ownership.** It is currently used by loot factory code and potentially by entity
   visual code. If v216's TownNodeFactory also needs it, it should live in a shared utility rather
   than LootNodeFactory. Plan should check all call sites first.

3. **`PROJECTILE_LERP_SECONDS` ref in `_move_projectile_node`.** That function (line 5516) uses a
   constant currently in `main.gd`. After extraction, it references `ClientConstants.PROJECTILE_LERP_SECONDS`.
   Confirm `visual_replay_enabled` flag read in the same function doesn't prevent extraction (it
   stays in `main.gd`, so it's fine — the constant reference is the only moved piece).
