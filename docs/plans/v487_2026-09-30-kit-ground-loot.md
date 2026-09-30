# v487 Plan — Kit ground loot

Status: Implemented
Spec: [`docs/specs/v487_spec-kit-ground-loot.md`](../specs/v487_spec-kit-ground-loot.md)
Goal: weapon/off-hand drops render the KayKit model the hero wields, posed from data and tinted
without losing the kit texture.
Architecture: `item_visuals.v0.json` becomes the only item → model link for hand items (ground and
equipped), read through a new static `ItemVisualsLoader`. `LootNodeFactory` resolves hand items
from it, falls back to the family `3d_model` for armor/jewelry, reads the ground pose from
`equipment_display.v0.json`, and tints through `ModelTint`. The validator forbids `3d_model` on
hand-item families.
Tech stack: Godot 4 GDScript client, shared presentation JSON + schema, Python validator.

## Baseline and shortcut decision

- Reuses v475 kit weapon GLBs and the `item_visuals.v0.json` mapping, v474 `ModelTint`, and the
  existing `EquipmentDisplayLoader` static loader (extended, not duplicated).
- Asset/plugin decision: **adopt** vendored KayKit weapon/shield/spellbook GLBs; **reject** new
  prop vendoring this slice; **keep** generated armor/jewelry fallbacks (ADR-0018 D10 explicit
  fallback). No plugin.
- Spec review (2026-09-30): PASS. All 33 hand items/templates already have an `item_visuals`
  entry; 13 hand families (sword, greatsword, dagger, axe, war_hammer, hammer, mace, spear, halberd,
  staff, book, bow, shield) lose `3d_model`; `3d_model` is optional in the family schema; loot pick
  colliders are fixed-size per entity type (`main.gd` `_attach_pick_collider`), so model size
  cannot change pickup reach.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/item_visuals_loader.gd` | `class_name ItemVisualsLoader` static loader for `item_visuals.v0.json` (`ensure_loaded`, `invalidate`, `all`, `visual_for`, `hand_asset_id`) |
| Modify | `client/scripts/equipment_visuals.gd` | `_load_data` reads visuals via `ItemVisualsLoader` |
| Modify | `client/scripts/equipment_display_loader.gd` | parse `ground_pose` (`default`, `rig_native`, `assets`); `ground_pose_for(asset_id, rig_native)` |
| Modify | `shared/assets/equipment_display.v0.json` + `.schema.json` | `ground_pose` data (today's literals as `default`, kit pose as `rig_native`) |
| Modify | `client/scripts/loot_node_factory.gd` | hand items from `ItemVisualsLoader`; pose from data; `ModelTint`; rig-native tint strength |
| Modify | `shared/assets/item_presentations.v0.json` | drop `3d_model` from the 13 hand families |
| Modify | `tools/validate_item_presentations.py` | fail when a family whose items map to a hand `item_visuals` entry sets `3d_model` |
| Create | `tools/test_validate_item_presentations.py` | focused validator tests (pass + negative) |
| Create | `client/tests/test_item_visuals_loader.gd` | loader unit test; register in `scripts/client_smoke.sh` |
| Modify | `client/tests/test_item_visuals.gd`, `client/tests/test_loot_node_factory.gd` | catalog-derived ground model identity, armor/primitive fallback, tint, pose |
| Modify | `docs/CODEMAP.md`, ADR-0018 P3 row, `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | docs |
| Create | `docs/as-built/v487_kit-ground-loot.md` (+ `docs/as-built/assets/v487/*.png`) | as-built with captures |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` — not touched
- [x] `server/internal/game/game_test.go` — not touched
- [x] `tools/bot/run.py` — not touched
- [x] `tools/validate_shared.py` — not touched (the validator loads `item_visuals` itself)
- [x] Other over-limit file: none (`equipment_visuals.gd` 483, `loot_node_factory.gd` 396,
  `test_item_visuals.gd` 477 — all must stay ≤ 600)

Decision: the new loader is its own focused file; no extraction needed.

```bash
make maintainability
```

## Task 1 — Shared presentation data and validator

Files: `shared/assets/equipment_display.v0.json` + schema, `shared/assets/item_presentations.v0.json`,
`tools/validate_item_presentations.py`, `tools/test_validate_item_presentations.py`

- [x] 1.1 Add `ground_pose` to `equipment_display.v0.json`: `default` = today's code literals
  (`height 0.12`, `rotation_degrees {90, 35, 0}`, `scale 1.0`), `rig_native` = kit pose (tuned in
  Task 4), `assets` = per-asset overrides (initially empty). Extend the schema (all keys optional
  numbers, `additionalProperties: false`).
- [x] 1.2 Remove `3d_model` from the 13 hand families in `item_presentations.v0.json`.
- [x] 1.3 Validator: load `item_visuals.v0.json`; for each family setting `3d_model`, fail if any of
  its items maps to an `item_visuals` entry with slot `main_hand`/`off_hand`.
- [x] 1.4 Focused pytest: committed data passes; a hand family with `3d_model` fails; an armor
  family with `3d_model` passes.

```bash
make validate-shared
.venv/bin/pytest -q tools/test_validate_item_presentations.py
```

## Task 2 — Client loaders

Files: `client/scripts/item_visuals_loader.gd`, `client/scripts/equipment_visuals.gd`,
`client/scripts/equipment_display_loader.gd`, `client/tests/test_item_visuals_loader.gd`,
`scripts/client_smoke.sh`

- [x] 2.1 `ItemVisualsLoader` (static, `ensure_loaded`/`invalidate`), `hand_asset_id(item_def_id)`
  returns the asset only for `main_hand`/`off_hand` entries, plus `is_rig_native(item_def_id)`.
- [x] 2.2 `EquipmentVisuals._load_data` uses the loader (`reload_data_only` invalidates first).
- [x] 2.3 `EquipmentDisplayLoader.ground_pose_for(asset_id, rig_native)` merges
  `default` ← `rig_native` (if kit) ← `assets[asset_id]`, returning height, rotation (Vector3) and
  scale.
- [x] 2.4 Loader unit test (catalog-derived) and gate registration.

```bash
godot --headless --path client --script res://tests/test_item_visuals_loader.gd
godot --headless --path client --script res://tests/test_item_visuals.gd
```

## Task 3 — Ground model resolution, pose and tint

Files: `client/scripts/loot_node_factory.gd`, `client/tests/test_loot_node_factory.gd`,
`client/tests/test_item_visuals.gd`

- [x] 3.1 `ground_model_asset_id(item_def_id)`: `ItemVisualsLoader.hand_asset_id` first, else the
  presentation `3d_model`. Drop the `fallback_equipment_off_hand_v0 → null` special case.
- [x] 3.2 Pose from `EquipmentDisplayLoader.ground_pose_for`; scale keeps
  `GROUND_EQUIPMENT_MODEL_SCALE × presentation ground.scale × ground_multiplier × pose.scale`.
- [x] 3.3 `apply_model_tint` uses `ModelTint.tinted_material`; rig-native models tint with
  `WHITE.lerp(rarity_tint, rig_native_tint_strength())`.
- [x] 3.4 Tests: every hand `item_visuals` entry → `GroundModel_<asset_id>`; helm-family item →
  its fallback; gold/potion → primitive; kit tint keeps `albedo_texture` and the lerped colour;
  node transform equals the loaded pose.

```bash
godot --headless --path client --script res://tests/test_loot_node_factory.gd
godot --headless --path client --script res://tests/test_item_visuals.gd
make client-unit
```

## Task 4 — Visual gate and pose tuning

- [x] 4.1 `make regen-screenshots SUITE="floor-item"` (before on `HEAD`, after on the slice).
- [x] 4.2 Tune `ground_pose.rig_native` (and any per-asset override, e.g. bow/shield/spellbook) until
  kit weapons lie flat, readable, above the floor. Re-capture.
- [x] 4.3 Copy representative before/after PNGs to `docs/as-built/assets/v487/`.

## Task 5 — Bot scenarios

No new scenario (spec §Test and bot proof: headless `gl_compatibility` cannot judge the kit look;
ADR-0018 D9). Regression: an existing loot pickup client scenario still passes.

```bash
make bot-client SCENARIO=click_to_kill HEADLESS=1
```

## Task 6 — Lifecycle docs and CI

- [x] ADR-0018 P3 row: ground loot weapons/off-hands on kit models (armor/jewelry fallbacks remain).
- [x] `docs/CODEMAP.md`: `item_visuals_loader.gd`.
- [x] `docs/as-built/v487_kit-ground-loot.md`, lifecycle row, `PROGRESS.md` current status.

```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `make client-unit`
- [x] `make regen-screenshots SUITE="floor-item"` reviewed
- [x] `make ci`

## Execution notes (2026-09-30)

- Task 4 tuning added two optional pose keys beyond the plan: `max_extent` (cap the longest side;
  kit polearms are up to 3.1 m native) and `rest_on_floor` (centre posed bounds on the loot root and
  sit them on the floor; kit weapons rotate around a grip origin). Per-asset overrides: bow
  (`x: 0`), closed spellbook (`z: 90`), kit shields (`scale: 0.75`).
- `docs/CODEMAP.md` was updated during Task 2 (the `validate_codemap` gate requires new scripts to
  be listed).
