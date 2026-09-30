# v493 Town Floor Detail Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

Status: Draft
Spec: [`docs/specs/v493_spec-town-floor-detail.md`](../specs/v493_spec-town-floor-detail.md)

**Goal:** The live town gets a soft plaza edge (dirt band and stone rim), stone paths from the plaza to the vendor and mystery seller, and deterministic dirt patches and stone chunks on the grass.

**Architecture:** A new pure-planning module `client/scripts/town_ground_detail.gd` turns `town_presentation.v0.json` -> `dressing` data into tile layers (core / rim / edge) and scatter placements, then builds MultiMeshes from them. `TownDressing` delegates to it and stays small. All tuning is data; the planner takes explicit `frame` data, so it is unit-testable without a scene.

**Tech Stack:** GDScript (Godot 4, headless tests), Python (pytest, JSON-schema validators), KayKit CC0 glb assets, `make` targets.

## Global Constraints

Copied from the spec and repo policy. Every task implicitly includes these.

- Client presentation only: no change to `server/`, `shared/protocol/`, `shared/golden/`, `shared/rules/`, bot scenarios.
- Data-driven: every count, width, weight and radius lives in `shared/assets/town_presentation.v0.json` -> `dressing`; nothing tuning-related is hardcoded in GDScript.
- Test Locking Policy: expectations derive from the loaded catalog. No pinned coordinates, no exact tile counts.
- Determinism: placement uses `hash(Vector3i(...))` only (same scheme as `DungeonKitFloor.pick`); never `randf`/`randi`.
- Files stay at or below 600 lines; `town_dressing.gd` must end smaller than it starts (140 lines). No `helpers=globals()` extraction.
- Asset provenance: CC0 only; manifest entry needs `origin`, `source_url`, `license`, `sha256` (sha256 of the vendored glb, which equals the staged source file).
- Do **not** run `make ci-full`. Run targeted tests per task; `make ci` once, in Task 6.
- Godot import churn: after `--import`, `git status` and revert any *pre-existing* `.glb.import` file that Godot rewrote (memory: godot-import-churn). Only the new files' sidecars are kept.
- Commits: use the `commit-commands:commit` skill. Message style `feat: v493: <what>`. Append `Co-Authored-By: Claude Sonnet 5.5 <noreply@anthropic.com>`.

**Branch:** `v493-town-floor-detail` off `main` (`e8973608`).

Commands used throughout (run from the repo root):

```bash
GODOT=${GODOT:-godot}
gd() { "$GODOT" --headless --rendering-method gl_compatibility --path client --script "$1"; }
```

## File Structure

| File | Responsibility |
|------|----------------|
| `client/scripts/town_ground_detail.gd` (new) | Frame, region, layers, scatter planning (pure) + MultiMesh building |
| `client/scripts/town_dressing.gd` (modify) | Adds the ground-detail root; loses `build_plaza`; `in_plaza` reuses `capsule_contains` |
| `client/scripts/town_presentation_loader.gd` | No code change expected: `dressing()` already returns the whole subtree. Pinned by a test. |
| `shared/assets/town_presentation.v0.json` + `.schema.json` | New `dressing` keys: `anchors`, `service_paths`, `plaza.rim`, `edge`, `scatter` |
| `assets/manifests/assets.v0.json` + 9 glb | Nine new CC0 environment assets |
| `tools/assets/validate_assets.py` | Step [9] also resolves the new ids |
| `tools/test_town_dressing.py` | Anchors == world preset; targets known; ids in manifest; scatter clear of fence |
| `client/tests/test_town_ground_detail.gd` (new) | Planner and builder tests |
| `scripts/client_smoke.sh` | Register the new test |

Data shape added under `dressing` (values are the starting point; the visual gate in Task 5 tunes them):

```jsonc
"anchors": [ { "id": "town_vendor", "position": { "x": 20, "y": 12 } }, ... ],  // == world preset gameplay points
"service_paths": { "path_width_m": 3.0, "targets": ["town_vendor", "town_mystery_seller"] },
"plaza": { ..., "rim": { "width_m": 2.0, "tile_variants": [ {asset_id, weight}, ... ] } },
"edge":  { "enabled": true, "width_m": 2.0, "surface_y": 0.012, "tile_variants": [ {asset_id, weight}, ... ] },
"scatter": {
  "enabled": true, "radius_m": 13.0, "cell_m": 3.0, "occupancy_percent": 50,
  "patch_share_percent": 55, "min_clearance_m": 1.5, "fence_clearance_m": 1.5, "surface_y": 0.0,
  "patches": [ { "asset_id", "weight", "radius_m", "scale_min", "scale_max" } ],
  "rocks":   [ { "asset_id", "weight", "radius_m", "scale_min", "scale_max" } ]
}
```

`path_width_m` is 3.0, not 2.0, on purpose: a band with half-width >= tile/sqrt(2) (1.414 m for the 2 m tile) contains every grid cell a line crosses, so a diagonal path (mystery seller) stays 4-connected. Task 3 tests connectivity, so this is verified, not assumed.

---

### Task 1: Vendor the nine assets and register them in the manifest

**Files:**
- Create: `client/assets/environment/kaykit_dungeon/{floor_dirt_small_A,floor_dirt_small_B,floor_dirt_small_C,floor_dirt_small_D,floor_dirt_small_weeds,floor_dirt_large,floor_dirt_large_rocky}.glb` (+ Godot-generated `.import` and `_dungeon_texture.png(.import)` sidecars)
- Create: `client/assets/environment/kaykit_resource_bits/{stone_chunks_small,stone_chunks_large}.glb` (+ sidecars), `client/assets/environment/kaykit_resource_bits/LICENSE.txt`
- Modify: `assets/manifests/assets.v0.json`

**Interfaces:**
- Produces manifest ids used by every later task:
  `kaykit_dungeon_floor_dirt_small_a_v0`, `..._b_v0`, `..._c_v0`, `..._d_v0`, `kaykit_dungeon_floor_dirt_small_weeds_v0`, `kaykit_dungeon_floor_dirt_large_v0`, `kaykit_dungeon_floor_dirt_large_rocky_v0`, `kaykit_resource_bits_stone_chunks_small_v0`, `kaykit_resource_bits_stone_chunks_large_v0`.

- [ ] **Step 1: Create the branch and confirm the staged sources exist**

```bash
git switch -c v493-town-floor-detail
K=".artifacts/kaykit/extracted/KayKit-Dungeon-Remastered-1.0-b0ca9bd96a8072ab36a3a5464f00ed1e06a16d07/addons/kaykit_dungeon_remastered/Assets/gltf"
R=".artifacts/kaykit/itch/resource_bits_1.0/KayKit_ResourceBits_1.0_FREE"
for n in floor_dirt_small_A floor_dirt_small_B floor_dirt_small_C floor_dirt_small_D floor_dirt_small_weeds floor_dirt_large floor_dirt_large_rocky; do test -f "$K/$n.gltf.glb" && echo "ok $n"; done
test -f "$R/Assets/gltf/Stone_Chunks_Small.gltf" && test -f "$R/Assets/gltf/Stone_Chunks_Large.gltf" && test -f "$R/License.txt" && echo "ok resource bits"
```

Expected: seven `ok floor_dirt_*` lines and `ok resource bits`. If anything is missing, stop: the packs must be re-staged (`docs/researchs/kaykit-asset-inventory.md`).

- [ ] **Step 2: Copy the Dungeon pieces and pack the Resource Bits pieces**

```bash
D=client/assets/environment/kaykit_dungeon
for n in floor_dirt_small_A floor_dirt_small_B floor_dirt_small_C floor_dirt_small_D floor_dirt_small_weeds floor_dirt_large floor_dirt_large_rocky; do
  cp "$K/$n.gltf.glb" "$D/$n.glb"
done
B=client/assets/environment/kaykit_resource_bits
mkdir -p "$B"
python3 -m tools.assets.gltf_to_glb "$R/Assets/gltf/Stone_Chunks_Small.gltf" "$B/stone_chunks_small.glb"
python3 -m tools.assets.gltf_to_glb "$R/Assets/gltf/Stone_Chunks_Large.gltf" "$B/stone_chunks_large.glb"
cp "$R/License.txt" "$B/LICENSE.txt"
```

Expected: no error output; the two `stone_chunks_*.glb` files exist and are larger than 100 KB (the 1024 px texture is embedded).

- [ ] **Step 3: Register the manifest entries with a script (never hand-edit sha256)**

Save this one-shot script as `register_v493_town_assets.py` in the session scratchpad directory (it must not be committed) and run it from the repo root:

```python
import hashlib, json
from pathlib import Path

ROOT = Path.cwd()
MANIFEST = ROOT / "assets/manifests/assets.v0.json"
data = json.loads(MANIFEST.read_text(encoding="utf-8"))
assets = data["assets"]

DUNGEON = "client/assets/environment/kaykit_dungeon/"
BITS = "client/assets/environment/kaykit_resource_bits/"
entries = {
    "kaykit_dungeon_floor_dirt_small_a_v0": (DUNGEON, "floor_dirt_small_A"),
    "kaykit_dungeon_floor_dirt_small_b_v0": (DUNGEON, "floor_dirt_small_B"),
    "kaykit_dungeon_floor_dirt_small_c_v0": (DUNGEON, "floor_dirt_small_C"),
    "kaykit_dungeon_floor_dirt_small_d_v0": (DUNGEON, "floor_dirt_small_D"),
    "kaykit_dungeon_floor_dirt_small_weeds_v0": (DUNGEON, "floor_dirt_small_weeds"),
    "kaykit_dungeon_floor_dirt_large_v0": (DUNGEON, "floor_dirt_large"),
    "kaykit_dungeon_floor_dirt_large_rocky_v0": (DUNGEON, "floor_dirt_large_rocky"),
    "kaykit_resource_bits_stone_chunks_small_v0": (BITS, "stone_chunks_small"),
    "kaykit_resource_bits_stone_chunks_large_v0": (BITS, "stone_chunks_large"),
}
for asset_id, (folder, stem) in entries.items():
    path = folder + stem + ".glb"
    sha = hashlib.sha256((ROOT / path).read_bytes()).hexdigest()
    if folder == DUNGEON:
        origin = f"KayKit Dungeon Remastered 1.0 by Kay Lousberg (Assets/gltf/{stem}.gltf.glb @ b0ca9bd96a80)"
        url = "https://github.com/KayKit-Game-Assets/KayKit-Dungeon-Remastered-1.0"
    else:
        origin = ("KayKit Resource Bits 1.0 FREE by Kay Lousberg "
                  f"(Assets/gltf/{stem.title()}.gltf packed with tools.assets.gltf_to_glb; owner download, archive sha256 7056f131…)")
        url = "https://kaylousberg.itch.io/resource-bits"
    assets[asset_id] = {
        "type": "environment",
        "runtime_path": path,
        "format": "glb",
        "scale_unit": "meters",
        "provenance": {"origin": origin, "source_url": url, "license": "CC0-1.0", "sha256": sha},
    }
MANIFEST.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
print("registered", len(entries))
```

Run: `python3 <scratchpad>/register_v493_town_assets.py` (from the repo root)
Expected: `registered 9`, and `git diff --stat assets/manifests/assets.v0.json` shows only added lines (a JSON round trip of this file is byte-exact, verified at planning time).

- [ ] **Step 4: Run Godot import, then discard churn on old files**

```bash
"$GODOT" --headless --rendering-method gl_compatibility --path client --import >/dev/null 2>&1 || true
git status --short client/assets | grep -v '^??' 
```

Expected: the second command prints nothing. If it lists modified pre-existing `*.glb.import` files, run `git checkout -- <those files>`. New files appear only as `??`.

- [ ] **Step 5: Validate against the asset budgets and the orphan check**

Run: `make validate-assets`
Expected: PASS, including the nine new ids under `[8] asset budgets` (environment: <= 4000 tris, <= 1024 px) and `[7]` no orphans. If `[7]` reports `..._resource_bits_texture.png` files as orphans, the sidecar prefix rule in `manifest_client_asset_paths` did not match the extracted texture name. Fix the orphan rule minimally in `tools/assets/validate_assets.py` (it matches `stem + "_"`), do not rename assets.

- [ ] **Step 6: Verify each new piece resolves to one mesh**

```bash
cat > client/tests/_v493_probe.gd <<'EOF'
extends SceneTree
func _initialize() -> void:
	var ids := ["kaykit_dungeon_floor_dirt_small_a_v0","kaykit_dungeon_floor_dirt_small_b_v0","kaykit_dungeon_floor_dirt_small_c_v0","kaykit_dungeon_floor_dirt_small_d_v0","kaykit_dungeon_floor_dirt_small_weeds_v0","kaykit_dungeon_floor_dirt_large_v0","kaykit_dungeon_floor_dirt_large_rocky_v0","kaykit_resource_bits_stone_chunks_small_v0","kaykit_resource_bits_stone_chunks_large_v0"]
	for id in ids:
		var node := KitPieceLibrary.instantiate(id)
		var count := KitPieceLibrary.mesh_instances(node).size() if node != null else -1
		print("%s meshes=%d bounds=%s" % [id, count, KitPieceLibrary.bounds(id)])
	quit()
EOF
gd res://tests/_v493_probe.gd; rm client/tests/_v493_probe.gd
```

Expected: nine lines, each `meshes=1`. If a piece reports `meshes=0` or `>1`, `KitPieceLibrary.mesh()` (first mesh only) would drop geometry; record the id and either pick a different piece or extend the plan before continuing. The dirt small tiles must report a footprint of 2 x 2 m and the large ones 4 x 4 m (matches the planning-time gltf accessor check). The probe file is deleted right after running; do not commit it.

- [ ] **Step 7: Commit**

```bash
git add assets/manifests/assets.v0.json client/assets/environment/kaykit_dungeon client/assets/environment/kaykit_resource_bits
git status --short   # only the new files, nothing unrelated
```

Then invoke the `commit-commands:commit` skill. Message: `feat: v493: vendor dirt tiles and stone chunks for the town floor`.

---

### Task 2: Data, schema, loader pin and Python gates

**Files:**
- Modify: `shared/assets/town_presentation.v0.json`, `shared/assets/town_presentation.v0.schema.json`
- Modify: `tools/assets/validate_assets.py` (step [9], around lines 316-322)
- Modify: `tools/test_town_dressing.py`
- Create: `client/tests/test_town_ground_detail.gd` (first test only: loader pin)
- Modify: `scripts/client_smoke.sh` (register the test, next to line 170)

**Interfaces:**
- Consumes: manifest ids from Task 1.
- Produces: `TownPresentationLoader.dressing()` returns `anchors`, `service_paths`, `plaza.rim`, `edge`, `scatter` exactly as written in the data shape above. Task 3 reads them.

- [ ] **Step 1: Write the failing Python tests**

Append to `tools/test_town_dressing.py`:

```python
MANIFEST = json.loads((ROOT / "assets/manifests/assets.v0.json").read_text(encoding="utf-8"))["assets"]


def _all_ground_asset_ids() -> list[str]:
    ids = [v["asset_id"] for v in DRESSING["plaza"]["tile_variants"]]
    ids += [v["asset_id"] for v in DRESSING["plaza"]["rim"]["tile_variants"]]
    ids += [v["asset_id"] for v in DRESSING["edge"]["tile_variants"]]
    ids += [v["asset_id"] for v in DRESSING["scatter"]["patches"]]
    ids += [v["asset_id"] for v in DRESSING["scatter"]["rocks"]]
    return ids


def test_anchors_match_the_world_preset_exactly() -> None:
    anchors = {a["id"]: _xy(a["position"]) for a in DRESSING["anchors"]}
    assert anchors == _gameplay_points(), "dressing.anchors must list every world-preset gameplay point"


def test_service_path_targets_are_known_anchors() -> None:
    anchors = {a["id"] for a in DRESSING["anchors"]}
    targets = DRESSING["service_paths"]["targets"]
    assert targets, "at least one service path target"
    for target in targets:
        assert target in anchors, f"service path target {target} is not an anchor"


def test_ground_asset_ids_exist_in_the_manifest() -> None:
    for asset_id in _all_ground_asset_ids():
        assert asset_id in MANIFEST, f"{asset_id} missing from the asset manifest"
        assert MANIFEST[asset_id]["type"] == "environment"


def test_scatter_fence_clearance_is_positive() -> None:
    # The planner excludes the ring [fence - clearance, fence + clearance] (unit-tested in GDScript),
    # so the only thing to pin in data is that the clearance exists and the scatter radius is sane.
    scatter = DRESSING["scatter"]
    assert float(scatter["fence_clearance_m"]) > 0.0
    assert float(scatter["radius_m"]) > 0.0


def test_scatter_variants_are_well_formed() -> None:
    for group in ("patches", "rocks"):
        for v in DRESSING["scatter"][group]:
            assert v["weight"] >= 1 and v["radius_m"] > 0.0
            assert 0.0 < v["scale_min"] <= v["scale_max"], f"{v['asset_id']} scale range"
```

- [ ] **Step 2: Run to confirm they fail**

Run: `.venv/bin/pytest tools/test_town_dressing.py -v`
Expected: FAIL with `KeyError: 'anchors'` (or `'rim'`) on the new tests; the three v491 tests still pass.

- [ ] **Step 3: Generate the anchors block from the world preset (do not hand-type positions)**

```bash
python3 - <<'EOF'
import json
w = json.load(open("shared/rules/worlds.v0.json"))["worlds"]["dungeon_levels"]
pts = {"player_spawn": w["player"]["position"]}
for e in w["entities"]:
    if e["type"] in ("interactable", "monster"):
        pts[e.get("interactable_def_id") or e.get("monster_def_id")] = e["position"]
print(json.dumps([{"id": k, "position": {"x": v["x"], "y": v["y"]}} for k, v in pts.items()], indent=2))
EOF
```

Expected: fourteen entries (player_spawn, town_exit_gate, stairs_down, teleporter, town_blacksmith, town_stash, town_bishop, town_quest_giver, town_vendor, town_mystery_seller, town_market_board, town_mercenary_board, town_unique_chest, town_training_doll). Paste the printed array as `dressing.anchors`.

- [ ] **Step 4: Add the remaining data to `town_presentation.v0.json` -> `dressing`**

Add inside `dressing` (leave existing keys untouched; `plaza.rim` goes inside `plaza`):

```json
"anchors": [ ...the array from Step 3... ],
"service_paths": {
  "path_width_m": 3.0,
  "targets": ["town_vendor", "town_mystery_seller"]
},
"edge": {
  "enabled": true,
  "width_m": 2.0,
  "surface_y": 0.012,
  "tile_variants": [
    { "asset_id": "kaykit_dungeon_floor_dirt_small_a_v0", "weight": 3 },
    { "asset_id": "kaykit_dungeon_floor_dirt_small_b_v0", "weight": 3 },
    { "asset_id": "kaykit_dungeon_floor_dirt_small_c_v0", "weight": 3 },
    { "asset_id": "kaykit_dungeon_floor_dirt_small_d_v0", "weight": 3 },
    { "asset_id": "kaykit_dungeon_floor_dirt_small_weeds_v0", "weight": 1 }
  ]
},
"scatter": {
  "enabled": true,
  "radius_m": 13.0,
  "cell_m": 3.0,
  "occupancy_percent": 50,
  "patch_share_percent": 55,
  "min_clearance_m": 1.5,
  "fence_clearance_m": 1.5,
  "surface_y": 0.0,
  "patches": [
    { "asset_id": "kaykit_dungeon_floor_dirt_large_v0", "weight": 2, "radius_m": 2.0, "scale_min": 0.7, "scale_max": 1.0 },
    { "asset_id": "kaykit_dungeon_floor_dirt_large_rocky_v0", "weight": 1, "radius_m": 2.0, "scale_min": 0.7, "scale_max": 1.0 },
    { "asset_id": "kaykit_dungeon_floor_dirt_small_weeds_v0", "weight": 2, "radius_m": 1.0, "scale_min": 0.8, "scale_max": 1.2 }
  ],
  "rocks": [
    { "asset_id": "kaykit_resource_bits_stone_chunks_small_v0", "weight": 3, "radius_m": 0.6, "scale_min": 0.5, "scale_max": 0.9 },
    { "asset_id": "kaykit_resource_bits_stone_chunks_large_v0", "weight": 1, "radius_m": 0.8, "scale_min": 0.35, "scale_max": 0.6 }
  ]
}
```

And inside `dressing.plaza`:

```json
"rim": {
  "width_m": 2.0,
  "tile_variants": [
    { "asset_id": "kaykit_dungeon_floor_tile_small_v0", "weight": 5 },
    { "asset_id": "kaykit_dungeon_floor_tile_small_decorated_v0", "weight": 1 },
    { "asset_id": "kaykit_dungeon_floor_tile_small_weeds_a_v0", "weight": 1 }
  ]
}
```

- [ ] **Step 5: Extend the JSON schema**

`town_presentation.v0.schema.json` has no shared `$defs` for tile variants, so the fragments below are fully expanded (paste as-is). Two edits:

**(a)** Inside `dressing.properties.plaza.properties`, next to `tile_variants`, add `rim`:

```json
"rim": {
  "type": "object",
  "additionalProperties": false,
  "required": ["width_m", "tile_variants"],
  "properties": {
    "width_m": { "type": "number", "minimum": 0 },
    "tile_variants": {
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["asset_id", "weight"],
        "properties": {
          "asset_id": { "type": "string", "minLength": 1 },
          "weight": { "type": "integer", "minimum": 1 }
        }
      }
    }
  }
}
```

**(b)** Inside `dressing.properties` (siblings of `plaza`, `min_clearance_m`, `props`), add the four new keys. They are optional (not added to `dressing.required`), so data without them still validates:

```json
"anchors": {
  "type": "array",
  "items": {
    "type": "object",
    "additionalProperties": false,
    "required": ["id", "position"],
    "properties": {
      "id": { "type": "string", "minLength": 1 },
      "position": {
        "type": "object",
        "additionalProperties": false,
        "required": ["x", "y"],
        "properties": { "x": { "type": "number" }, "y": { "type": "number" } }
      }
    }
  }
},
"service_paths": {
  "type": "object",
  "additionalProperties": false,
  "required": ["path_width_m", "targets"],
  "properties": {
    "path_width_m": { "type": "number", "exclusiveMinimum": 0 },
    "targets": { "type": "array", "items": { "type": "string", "minLength": 1 } }
  }
},
"edge": {
  "type": "object",
  "additionalProperties": false,
  "required": ["enabled", "width_m", "surface_y", "tile_variants"],
  "properties": {
    "enabled": { "type": "boolean" },
    "width_m": { "type": "number", "minimum": 0 },
    "surface_y": { "type": "number", "minimum": 0, "maximum": 0.2 },
    "tile_variants": {
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["asset_id", "weight"],
        "properties": {
          "asset_id": { "type": "string", "minLength": 1 },
          "weight": { "type": "integer", "minimum": 1 }
        }
      }
    }
  }
},
"scatter": {
  "type": "object",
  "additionalProperties": false,
  "required": ["enabled", "radius_m", "cell_m", "occupancy_percent", "patch_share_percent", "min_clearance_m", "fence_clearance_m", "surface_y", "patches", "rocks"],
  "properties": {
    "enabled": { "type": "boolean" },
    "radius_m": { "type": "number", "exclusiveMinimum": 0 },
    "cell_m": { "type": "number", "exclusiveMinimum": 0 },
    "occupancy_percent": { "type": "integer", "minimum": 0, "maximum": 100 },
    "patch_share_percent": { "type": "integer", "minimum": 0, "maximum": 100 },
    "min_clearance_m": { "type": "number", "minimum": 0 },
    "fence_clearance_m": { "type": "number", "minimum": 0 },
    "surface_y": { "type": "number", "minimum": 0, "maximum": 0.2 },
    "patches": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["asset_id", "weight", "radius_m", "scale_min", "scale_max"],
        "properties": {
          "asset_id": { "type": "string", "minLength": 1 },
          "weight": { "type": "integer", "minimum": 1 },
          "radius_m": { "type": "number", "exclusiveMinimum": 0 },
          "scale_min": { "type": "number", "exclusiveMinimum": 0 },
          "scale_max": { "type": "number", "exclusiveMinimum": 0 }
        }
      }
    },
    "rocks": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["asset_id", "weight", "radius_m", "scale_min", "scale_max"],
        "properties": {
          "asset_id": { "type": "string", "minLength": 1 },
          "weight": { "type": "integer", "minimum": 1 },
          "radius_m": { "type": "number", "exclusiveMinimum": 0 },
          "scale_min": { "type": "number", "exclusiveMinimum": 0 },
          "scale_max": { "type": "number", "exclusiveMinimum": 0 }
        }
      }
    }
  }
}
```

Then run `make validate-shared` to confirm both the schema and the Task 2 Step 4 data are accepted.

- [ ] **Step 6: Extend `validate_assets.py` step [9]**

In the `if town_path.is_file():` block (`tools/assets/validate_assets.py` ~line 318), after the existing `plaza.tile_variants` line, add:

```python
            plaza = dressing.get("plaza", {})
            kit_ids += [v["asset_id"] for v in plaza.get("rim", {}).get("tile_variants", [])]  # v493
            kit_ids += [v["asset_id"] for v in dressing.get("edge", {}).get("tile_variants", [])]
            scatter = dressing.get("scatter", {})
            kit_ids += [v["asset_id"] for v in scatter.get("patches", []) + scatter.get("rocks", [])]
```

- [ ] **Step 7: Write the loader-pin GDScript test and register it**

Create `client/tests/test_town_ground_detail.gd`:

```gdscript
extends SceneTree

## v493 town ground detail: edge/rim/paths/scatter planning derived from
## town_presentation.v0.json -> dressing. No pinned coordinates or counts.

var _pass_count := 0
var _fail_count := 0


func _initialize() -> void:
	_test_loader_keeps_ground_detail_keys()
	_finish()


func _test_loader_keeps_ground_detail_keys() -> void:
	var dressing := TownPresentationLoader.dressing()
	for key in ["anchors", "service_paths", "edge", "scatter"]:
		_assert_true("dressing keeps %s (v491 merge-drop regression)" % key, dressing.has(key))
	_assert_true("plaza keeps rim", (dressing.get("plaza", {}) as Dictionary).has("rim"))


func _assert_true(label: String, value: bool) -> void:
	if value:
		_pass_count += 1
		return
	_fail_count += 1
	printerr("[gdtest] FAIL %s" % label)


func _finish() -> void:
	if _fail_count > 0:
		printerr("[gdtest] FAIL: test_town_ground_detail (%d passed, %d failed)" % [_pass_count, _fail_count])
		quit(1)
		return
	print("[gdtest] PASS: test_town_ground_detail (%d passed, %d failed)" % [_pass_count, _fail_count])
	quit()
```

In `scripts/client_smoke.sh`, directly after the `town dressing test` gate (line ~170) add:

```bash
run_gate "GDScript town ground detail test" "[gdtest] PASS: test_town_ground_detail" res://tests/test_town_ground_detail.gd
```

- [ ] **Step 8: Run everything for this task**

```bash
.venv/bin/pytest tools/test_town_dressing.py -v
make validate-shared
make validate-assets
gd res://tests/test_town_ground_detail.gd
```

Expected: pytest all PASS (eight tests: the three v491 ones plus five new); `validate-shared` accepts the schema and data; `validate-assets` step [9] lists the new ids as resolving; the Godot test prints `[gdtest] PASS: test_town_ground_detail (5 passed, 0 failed)`.

- [ ] **Step 9: Commit** (`commit-commands:commit`): `feat: v493: town ground detail data, schema and gates`.

---

### Task 3: Pure planner (frame, region, layers, paths, scatter)

**Files:**
- Create: `client/scripts/town_ground_detail.gd`
- Modify: `client/tests/test_town_ground_detail.gd`

**Interfaces:**
- Consumes: `dressing` dictionary (Task 2 shape), `TownPresentationLoader.center()/gate_position()/radius_m()`, `KitPieceLibrary.bounds()`, `DungeonKitFloor.pick()`.
- Produces (all `static`, on `class_name TownGroundDetail`):
  - `frame(dressing: Dictionary) -> Dictionary` with keys `center: Vector2`, `gate: Vector2`, `fence_radius: float`, `tile: float`.
  - `capsule_contains(p: Vector2, a: Vector2, b: Vector2, half: float) -> bool`
  - `anchors(dressing) -> Dictionary` (id -> Vector2)
  - `path_segments(dressing, center: Vector2) -> Array` of `{id, a, b, half}`
  - `region_capsules(dressing, frame) -> Array` of `{a, b, half}`; `region_contains(caps: Array, p: Vector2) -> bool`
  - `cell_index(p: Vector2, frame) -> Vector2i`, `cell_center(idx: Vector2i, frame) -> Vector2`
  - `layers(dressing, frame) -> Dictionary` with `"core"`, `"rim"`, `"edge"`, each an `Array` of `Vector2` cell centres
  - `scatter(dressing, frame) -> Array` of `{kind: "patch"|"rock", asset_id: String, position: Vector2, yaw_quarter: int, yaw_degrees: float, scale: float, radius_m: float}`
  - `rock_transform(pos: Vector2, yaw_degrees: float, scale: float, box: AABB, surface_y: float) -> Transform3D`
  - constants `LAYER_CORE = "core"`, `LAYER_RIM = "rim"`, `LAYER_EDGE = "edge"`, `SCATTER_SALT = 4931`

- [ ] **Step 1: Write the failing planner tests**

Replace `_initialize()` and add the helpers/tests in `client/tests/test_town_ground_detail.gd`:

```gdscript
func _initialize() -> void:
	var dressing := TownPresentationLoader.dressing()
	var frame := TownGroundDetail.frame(dressing)
	_test_loader_keeps_ground_detail_keys()
	_test_frame(frame)
	_test_layers_partition_and_definitions(dressing, frame)
	_test_paths_connect_plaza_to_targets(dressing, frame)
	_test_scatter_rules(dressing, frame)
	_test_scatter_is_deterministic_and_data_driven(dressing, frame)
	_test_rock_transform_seats_on_the_surface()
	_finish()


func _test_frame(frame: Dictionary) -> void:
	_assert_true("tile size is positive", float(frame["tile"]) > 0.0)
	_assert_true("fence radius comes from the town catalog", is_equal_approx(float(frame["fence_radius"]), TownPresentationLoader.radius_m()))


func _key(p: Vector2, frame: Dictionary) -> Vector2i:
	return TownGroundDetail.cell_index(p, frame)


func _set_of(cells: Array, frame: Dictionary) -> Dictionary:
	var out := {}
	for c in cells:
		out[_key(c, frame)] = true
	return out


func _chebyshev_steps(width_m: float, tile: float) -> int:
	return floori(width_m / tile + 0.001)


func _near(set: Dictionary, idx: Vector2i, steps: int) -> bool:
	for dy in range(-steps, steps + 1):
		for dx in range(-steps, steps + 1):
			if (dx != 0 or dy != 0) and set.has(idx + Vector2i(dx, dy)):
				return true
	return false


func _test_layers_partition_and_definitions(dressing: Dictionary, frame: Dictionary) -> void:
	var layers := TownGroundDetail.layers(dressing, frame)
	var caps := TownGroundDetail.region_capsules(dressing, frame)
	var tile := float(frame["tile"])
	var core := _set_of(layers["core"], frame)
	var rim := _set_of(layers["rim"], frame)
	var edge := _set_of(layers["edge"], frame)
	_assert_true("core is not empty", core.size() > 0)
	_assert_true("rim is not empty", rim.size() > 0)
	_assert_true("edge is not empty", edge.size() > 0)
	var overlap := 0
	for k in core:
		if rim.has(k) or edge.has(k):
			overlap += 1
	for k in rim:
		if edge.has(k):
			overlap += 1
	_assert_true("layers are disjoint", overlap == 0)
	var paved := {}
	for k in core:
		paved[k] = true
	for k in rim:
		paved[k] = true
	var bad_paved := 0
	for c in layers["core"] + layers["rim"]:
		if not TownGroundDetail.region_contains(caps, c):
			bad_paved += 1
	_assert_true("core and rim cells are inside the paved region", bad_paved == 0)
	var edge_steps := _chebyshev_steps(float((dressing["edge"] as Dictionary)["width_m"]), tile)
	var rim_steps := _chebyshev_steps(float(((dressing["plaza"] as Dictionary)["rim"] as Dictionary)["width_m"]), tile)
	var bad_edge := 0
	for c in layers["edge"]:
		var idx := _key(c, frame)
		if TownGroundDetail.region_contains(caps, c) or paved.has(idx) or not _near(paved, idx, edge_steps):
			bad_edge += 1
	_assert_true("edge cells are unpaved and within the edge width of paved", bad_edge == 0)
	var not_paved := {}
	var bad_rim := 0
	for c in layers["rim"]:
		var idx := _key(c, frame)
		for dy in range(-rim_steps, rim_steps + 1):
			for dx in range(-rim_steps, rim_steps + 1):
				var n := idx + Vector2i(dx, dy)
				if (dx != 0 or dy != 0) and not paved.has(n):
					not_paved[idx] = true
		if not not_paved.has(idx):
			bad_rim += 1
	_assert_true("rim cells touch unpaved ground within the rim width", bad_rim == 0)
	var bad_core := 0
	for c in layers["core"]:
		var idx := _key(c, frame)
		for dy in range(-rim_steps, rim_steps + 1):
			for dx in range(-rim_steps, rim_steps + 1):
				if (dx != 0 or dy != 0) and not paved.has(idx + Vector2i(dx, dy)):
					bad_core += 1
	_assert_true("core cells are deeper than the rim width", bad_core == 0)
	_assert_true("the plaza centre is core or rim", paved.has(_key(frame["center"], frame)))
	var reaches_gate := false
	for c in layers["core"] + layers["rim"]:
		if (c as Vector2).distance_to(frame["gate"]) <= float(frame["tile"]):
			reaches_gate = true
	_assert_true("the road still reaches the gate", reaches_gate)


func _test_paths_connect_plaza_to_targets(dressing: Dictionary, frame: Dictionary) -> void:
	var layers := TownGroundDetail.layers(dressing, frame)
	var paved := _set_of(layers["core"] + layers["rim"], frame)
	var anchors := TownGroundDetail.anchors(dressing)
	var targets: Array = (dressing["service_paths"] as Dictionary)["targets"]
	_assert_true("there are service path targets", targets.size() > 0)
	for id in targets:
		var target: Vector2 = anchors[str(id)]
		# 4-connected flood fill from the centre cell over paved cells.
		var seen := {}
		var stack: Array = [_key(frame["center"], frame)]
		while not stack.is_empty():
			var cur: Vector2i = stack.pop_back()
			if seen.has(cur) or not paved.has(cur):
				continue
			seen[cur] = true
			for d in [Vector2i(1, 0), Vector2i(-1, 0), Vector2i(0, 1), Vector2i(0, -1)]:
				stack.append(cur + d)
		var reached := false
		var half := float((dressing["service_paths"] as Dictionary)["path_width_m"]) * 0.5
		for idx in seen:
			if TownGroundDetail.cell_center(idx, frame).distance_to(target) <= half + float(frame["tile"]):
				reached = true
		_assert_true("path tiles connect the plaza to %s" % id, reached)


func _test_scatter_rules(dressing: Dictionary, frame: Dictionary) -> void:
	var cfg: Dictionary = dressing["scatter"]
	var placements := TownGroundDetail.scatter(dressing, frame)
	var layers := TownGroundDetail.layers(dressing, frame)
	var occupied := _set_of(layers["core"] + layers["rim"] + layers["edge"], frame)
	var center: Vector2 = frame["center"]
	var tile := float(frame["tile"])
	var steps := ceili(float(cfg["radius_m"]) / float(cfg["cell_m"]))
	var max_cells := (2 * steps + 1) * (2 * steps + 1)
	_assert_true("scatter produces placements", placements.size() > 0)
	_assert_true("scatter count is bounded by the grid it walks", placements.size() <= max_cells)
	var avoid: Array = []
	for pos in TownGroundDetail.anchors(dressing).values():
		avoid.append(pos)
	for prop in dressing.get("props", []):
		avoid.append(Vector2(float(prop["position"]["x"]), float(prop["position"]["y"])))
	var bad_radius := 0
	var bad_fence := 0
	var bad_avoid := 0
	var bad_tiles := 0
	for p in placements:
		var pos: Vector2 = p["position"]
		var r := float(p["radius_m"])
		if pos.distance_to(center) + r > float(cfg["radius_m"]) + 0.001:
			bad_radius += 1
		if absf(pos.distance_to(center) - float(frame["fence_radius"])) < float(cfg["fence_clearance_m"]) + r - 0.001:
			bad_fence += 1
		for a in avoid:
			if pos.distance_to(a) < float(cfg["min_clearance_m"]) + r - 0.001:
				bad_avoid += 1
		var idx := _key(pos, frame)
		var reach := ceili(r / tile) + 1
		for dy in range(-reach, reach + 1):
			for dx in range(-reach, reach + 1):
				var n := idx + Vector2i(dx, dy)
				if not occupied.has(n):
					continue
				var c := TownGroundDetail.cell_center(n, frame)
				var ddx := maxf(absf(pos.x - c.x) - tile * 0.5, 0.0)
				var ddy := maxf(absf(pos.y - c.y) - tile * 0.5, 0.0)
				if Vector2(ddx, ddy).length() < r - 0.001:
					bad_tiles += 1
	_assert_true("placements stay inside the scatter radius", bad_radius == 0)
	_assert_true("placements keep clear of the fence ring", bad_fence == 0)
	_assert_true("placements keep clear of anchors and props", bad_avoid == 0)
	_assert_true("placements do not touch paved or edge tiles", bad_tiles == 0)


func _test_scatter_is_deterministic_and_data_driven(dressing: Dictionary, frame: Dictionary) -> void:
	var a := TownGroundDetail.scatter(dressing, frame)
	var b := TownGroundDetail.scatter(dressing, frame)
	_assert_true("scatter is deterministic", a == b)
	var denser := dressing.duplicate(true)
	var denser_scatter: Dictionary = denser["scatter"]
	denser_scatter["occupancy_percent"] = 100
	var c := TownGroundDetail.scatter(denser, frame)
	_assert_true("occupancy_percent drives the count", c.size() > a.size())
	var off := dressing.duplicate(true)
	var off_scatter: Dictionary = off["scatter"]
	off_scatter["enabled"] = false
	_assert_true("scatter.enabled=false yields nothing", TownGroundDetail.scatter(off, frame).is_empty())
	var no_edge := dressing.duplicate(true)
	var no_edge_cfg: Dictionary = no_edge["edge"]
	no_edge_cfg["enabled"] = false
	_assert_true("edge.enabled=false yields no edge cells", (TownGroundDetail.layers(no_edge, frame)["edge"] as Array).is_empty())


func _test_rock_transform_seats_on_the_surface() -> void:
	var box := AABB(Vector3(-0.5, -0.1, -0.4), Vector3(1.0, 0.6, 0.8))
	var t := TownGroundDetail.rock_transform(Vector2(3.0, 4.0), 37.0, 0.5, box, 0.02)
	var bottom_y := t.origin.y + box.position.y * 0.5
	_assert_true("rock bottom sits on the surface", is_equal_approx(bottom_y, 0.02))
	var centre := t * box.get_center()
	_assert_true("rock centre lands on the requested xz", is_equal_approx(centre.x, 3.0) and is_equal_approx(centre.z, 4.0))
```

- [ ] **Step 2: Run to confirm the test fails**

Run: `gd res://tests/test_town_ground_detail.gd`
Expected: script error `Identifier "TownGroundDetail" not declared` (the class does not exist yet). If Godot says the class is unknown even after Step 3, run `"$GODOT" --headless --path client --import` to refresh the global class cache.

- [ ] **Step 3: Implement the planner**

Create `client/scripts/town_ground_detail.gd`:

```gdscript
## Town ground detail (ADR-0018, v493): stone rim, dirt edge band, paths to services and a
## deterministic scatter of dirt patches and stone chunks, all planned from
## town_presentation.v0.json -> dressing. Presentation only: the server never sees it.
##
## Planning functions are pure (data in, cells/placements out) so they are unit-testable without a
## scene; the builders at the bottom turn a plan into MultiMeshes. Placement uses hash() only (same
## scheme as DungeonKitFloor.pick) so every client renders the same town.
class_name TownGroundDetail
extends RefCounted

const LoaderScript := preload("res://scripts/town_presentation_loader.gd")
const LibraryScript := preload("res://scripts/kit_piece_library.gd")
const KitFloorScript := preload("res://scripts/dungeon_kit_floor.gd")

const PLAZA_NAME := "TownPlaza"
const RIM_NAME := "TownPlazaRim"
const EDGE_NAME := "TownPlazaEdge"
const SCATTER_NAME := "TownScatter"
const ROOT_NAME := "TownGround"

const LAYER_CORE := "core"
const LAYER_RIM := "rim"
const LAYER_EDGE := "edge"

## Tile-variant hash salts per layer (core keeps 0 so the v491 plaza look is unchanged).
const SALT_CORE := 0
const SALT_RIM := 1
const SALT_EDGE := 2
const SCATTER_SALT := 4931


## Everything the planner needs from outside the dressing dictionary.
static func frame(dressing: Dictionary) -> Dictionary:
	return {
		"center": LoaderScript.center(),
		"gate": LoaderScript.gate_position(),
		"fence_radius": LoaderScript.radius_m(),
		"tile": tile_size(dressing),
	}


## Grid pitch: the first plaza variant's footprint times plaza.tile_scale (same rule as v491).
static func tile_size(dressing: Dictionary) -> float:
	var plaza: Dictionary = dressing.get("plaza", {})
	var variants: Array = plaza.get("tile_variants", [])
	if variants.is_empty():
		return 0.0
	var box := LibraryScript.bounds(str((variants[0] as Dictionary).get("asset_id", "")))
	return maxf(box.size.x, box.size.z) * float(plaza.get("tile_scale", 1.0))


static func capsule_contains(p: Vector2, a: Vector2, b: Vector2, half: float) -> bool:
	var seg := b - a
	var t := clampf((p - a).dot(seg) / maxf(seg.length_squared(), 0.0001), 0.0, 1.0)
	return p.distance_to(a + seg * t) <= half


static func anchors(dressing: Dictionary) -> Dictionary:
	var out := {}
	for raw in dressing.get("anchors", []):
		if typeof(raw) != TYPE_DICTIONARY:
			continue
		var entry := raw as Dictionary
		var pos: Dictionary = entry.get("position", {})
		out[str(entry.get("id", ""))] = Vector2(float(pos.get("x", 0.0)), float(pos.get("y", 0.0)))
	return out


## One capsule from the town centre to each configured service anchor.
static func path_segments(dressing: Dictionary, center: Vector2) -> Array:
	var cfg: Dictionary = dressing.get("service_paths", {})
	var half := float(cfg.get("path_width_m", 0.0)) * 0.5
	var known := anchors(dressing)
	var out: Array = []
	if half <= 0.0:
		return out
	for id in cfg.get("targets", []):
		if known.has(str(id)):
			out.append({"id": str(id), "a": center, "b": known[str(id)], "half": half})
	return out


## Paved region as capsules: the plaza disc (a == b), the gate road, and the service paths.
static func region_capsules(dressing: Dictionary, frame_data: Dictionary) -> Array:
	var plaza: Dictionary = dressing.get("plaza", {})
	var center: Vector2 = frame_data["center"]
	var caps: Array = [
		{"a": center, "b": center, "half": float(plaza.get("radius_m", 0.0))},
		{"a": center, "b": frame_data["gate"], "half": float(plaza.get("path_width_m", 0.0)) * 0.5},
	]
	caps.append_array(path_segments(dressing, center))
	return caps


static func region_contains(caps: Array, p: Vector2) -> bool:
	for cap in caps:
		if capsule_contains(p, cap["a"], cap["b"], float(cap["half"])):
			return true
	return false


static func cell_index(p: Vector2, frame_data: Dictionary) -> Vector2i:
	var rel: Vector2 = (p - (frame_data["center"] as Vector2)) / float(frame_data["tile"])
	return Vector2i(roundi(rel.x), roundi(rel.y))


static func cell_center(idx: Vector2i, frame_data: Dictionary) -> Vector2:
	return (frame_data["center"] as Vector2) + Vector2(idx) * float(frame_data["tile"])


## True when any other cell within `steps` (Chebyshev) is in `set` == `want`.
static func _has_neighbour(paved: Dictionary, idx: Vector2i, steps: int, want_paved: bool) -> bool:
	for dy in range(-steps, steps + 1):
		for dx in range(-steps, steps + 1):
			if (dx != 0 or dy != 0) and paved.has(idx + Vector2i(dx, dy)) == want_paved:
				return true
	return false


## Grid layers. core: paved. rim: paved cells within plaza.rim.width_m of unpaved ground. edge:
## unpaved cells within edge.width_m of paved ground. Distances are Chebyshev between cell centres,
## so width_m == tile size selects the 8-neighbour ring.
static func layers(dressing: Dictionary, frame_data: Dictionary) -> Dictionary:
	var out := {LAYER_CORE: [], LAYER_RIM: [], LAYER_EDGE: []}
	var tile := float(frame_data["tile"])
	if tile <= 0.0:
		return out
	var caps := region_capsules(dressing, frame_data)
	var center: Vector2 = frame_data["center"]
	var rim_cfg: Dictionary = (dressing.get("plaza", {}) as Dictionary).get("rim", {})
	var edge_cfg: Dictionary = dressing.get("edge", {})
	var rim_steps := floori(float(rim_cfg.get("width_m", 0.0)) / tile + 0.001)
	var edge_steps := 0
	if bool(edge_cfg.get("enabled", false)):
		edge_steps = floori(float(edge_cfg.get("width_m", 0.0)) / tile + 0.001)
	var reach := 0.0
	for cap in caps:
		reach = maxf(reach, maxf(center.distance_to(cap["a"]), center.distance_to(cap["b"])) + float(cap["half"]))
	var span := ceili(reach / tile) + edge_steps + 1
	var paved := {}
	for iy in range(-span, span + 1):
		for ix in range(-span, span + 1):
			var idx := Vector2i(ix, iy)
			if region_contains(caps, cell_center(idx, frame_data)):
				paved[idx] = true
	for idx in paved:
		var layer := LAYER_RIM if rim_steps > 0 and _has_neighbour(paved, idx, rim_steps, false) else LAYER_CORE
		(out[layer] as Array).append(cell_center(idx, frame_data))
	if edge_steps > 0:
		for iy in range(-span, span + 1):
			for ix in range(-span, span + 1):
				var idx := Vector2i(ix, iy)
				if not paved.has(idx) and _has_neighbour(paved, idx, edge_steps, true):
					(out[LAYER_EDGE] as Array).append(cell_center(idx, frame_data))
	return out


static func _weighted(variants: Array, roll: int) -> int:
	var total := 0
	for v in variants:
		total += int((v as Dictionary).get("weight", 1))
	if total <= 0:
		return 0
	roll = roll % total
	for i in variants.size():
		roll -= int((variants[i] as Dictionary).get("weight", 1))
		if roll < 0:
			return i
	return 0


## Scatter plan: a jittered grid of scatter.cell_m cells inside scatter.radius_m. Each cell rolls
## occupancy, kind (patch/rock), variant, scale and yaw from hash(); a placement is dropped when its
## footprint would touch the fence ring, an anchor or v491 prop, or a paved/edge tile.
static func scatter(dressing: Dictionary, frame_data: Dictionary) -> Array:
	var cfg: Dictionary = dressing.get("scatter", {})
	var out: Array = []
	var tile := float(frame_data["tile"])
	var radius := float(cfg.get("radius_m", 0.0))
	var cell_m := float(cfg.get("cell_m", 0.0))
	if not bool(cfg.get("enabled", false)) or tile <= 0.0 or radius <= 0.0 or cell_m <= 0.0:
		return out
	var patches: Array = cfg.get("patches", [])
	var rocks: Array = cfg.get("rocks", [])
	if patches.is_empty() and rocks.is_empty():
		return out
	var center: Vector2 = frame_data["center"]
	var fence_radius := float(frame_data["fence_radius"])
	var clearance := float(cfg.get("min_clearance_m", 0.0))
	var fence_clear := float(cfg.get("fence_clearance_m", 0.0))
	var occupancy := int(cfg.get("occupancy_percent", 0))
	var patch_share := int(cfg.get("patch_share_percent", 0))
	var avoid: Array = anchors(dressing).values()
	for prop in dressing.get("props", []):
		var pp: Dictionary = (prop as Dictionary).get("position", {})
		avoid.append(Vector2(float(pp.get("x", 0.0)), float(pp.get("y", 0.0))))
	var lay := layers(dressing, frame_data)
	var occupied := {}
	for name in [LAYER_CORE, LAYER_RIM, LAYER_EDGE]:
		for c in lay[name]:
			occupied[cell_index(c, frame_data)] = true
	var steps := ceili(radius / cell_m)
	for iy in range(-steps, steps + 1):
		for ix in range(-steps, steps + 1):
			var h := absi(hash(Vector3i(ix, iy, SCATTER_SALT)))
			if h % 100 >= occupancy:
				continue
			var h2 := absi(hash(Vector3i(ix, iy, SCATTER_SALT + 1)))
			var jitter := Vector2(float((h / 100) % 1000) / 1000.0 - 0.5, float((h / 100000) % 1000) / 1000.0 - 0.5)
			var pos := center + Vector2(ix, iy) * cell_m + jitter * cell_m * 0.8
			var is_patch := (h2 % 100) < patch_share
			if patches.is_empty():
				is_patch = false
			elif rocks.is_empty():
				is_patch = true
			var group: Array = patches if is_patch else rocks
			var variant := group[_weighted(group, (h2 / 100) % 100000)] as Dictionary
			var t := float((h2 / 10000) % 1000) / 999.0
			var scale := lerpf(float(variant.get("scale_min", 1.0)), float(variant.get("scale_max", 1.0)), t)
			var r := float(variant.get("radius_m", 0.0)) * scale
			if not _placeable(pos, r, center, radius, fence_radius, fence_clear, clearance, avoid, occupied, frame_data):
				continue
			out.append({
				"kind": "patch" if is_patch else "rock",
				"asset_id": str(variant.get("asset_id", "")),
				"position": pos,
				"yaw_quarter": (h2 / 10000000) % 4,
				"yaw_degrees": float((h2 / 7) % 360),
				"scale": scale,
				"radius_m": r,
			})
	return out


static func _placeable(pos: Vector2, r: float, center: Vector2, radius: float, fence_radius: float, fence_clear: float, clearance: float, avoid: Array, occupied: Dictionary, frame_data: Dictionary) -> bool:
	var d := pos.distance_to(center)
	if d + r > radius:
		return false
	if absf(d - fence_radius) < fence_clear + r:
		return false
	for a in avoid:
		if pos.distance_to(a) < clearance + r:
			return false
	var tile := float(frame_data["tile"])
	var idx := cell_index(pos, frame_data)
	var reach := ceili(r / tile) + 1
	for dy in range(-reach, reach + 1):
		for dx in range(-reach, reach + 1):
			var n := idx + Vector2i(dx, dy)
			if not occupied.has(n):
				continue
			var c := cell_center(n, frame_data)
			var gap := Vector2(maxf(absf(pos.x - c.x) - tile * 0.5, 0.0), maxf(absf(pos.y - c.y) - tile * 0.5, 0.0))
			if gap.length() < r:
				return false
	return true


## Free-yaw placement that rests the piece's bottom on surface_y and centres it on `pos` (x, z).
static func rock_transform(pos: Vector2, yaw_degrees: float, scale: float, box: AABB, surface_y: float) -> Transform3D:
	var basis := Basis(Vector3.UP, deg_to_rad(yaw_degrees)).scaled(Vector3.ONE * scale)
	var c := box.get_center()
	var origin := Vector3(pos.x, surface_y - box.position.y * scale, pos.y) - basis * Vector3(c.x, 0.0, c.z)
	return Transform3D(basis, origin)
```

- [ ] **Step 4: Run the test and fix until green**

Run: `gd res://tests/test_town_ground_detail.gd`
Expected: `[gdtest] PASS: test_town_ground_detail (N passed, 0 failed)`.
If `path tiles connect the plaza to town_mystery_seller` fails, do not loosen the test: widen `service_paths.path_width_m` (data), since the connectivity argument requires half-width >= tile/sqrt(2). If `scatter produces placements` fails, the exclusions are eating everything: print `placements.size()` before the `_placeable` filter and check `scatter.radius_m` against the fence (15) and anchors, then tune the data.

- [ ] **Step 5: Commit** (`commit-commands:commit`): `feat: v493: town ground detail planner`.

---

### Task 4: Builders and TownDressing integration

**Files:**
- Modify: `client/scripts/town_ground_detail.gd` (append builders)
- Modify: `client/scripts/town_dressing.gd`
- Modify: `client/tests/test_town_ground_detail.gd`, `client/tests/test_town_dressing.gd`

**Interfaces:**
- Consumes: Task 3 planner API.
- Produces: `TownGroundDetail.build(dressing: Dictionary) -> Node3D`, a node named `TownGround` with children `TownPlaza` (core), `TownPlazaRim`, `TownPlazaEdge`, `TownScatter` (each present only when it has cells/placements), each holding one `MultiMeshInstance3D` per asset named `<layer>_<asset_id>`. `TownDressing.PLAZA_NAME` stays `"TownPlaza"` (`test_item_visuals.gd:373` finds it by name, recursive).

- [ ] **Step 1: Write the failing builder tests**

Add to `client/tests/test_town_ground_detail.gd` (and call `_test_build_creates_layer_nodes(dressing)` from `_initialize` before `_finish()`):

```gdscript
func _instance_total(layer: Node) -> int:
	var total := 0
	for child in layer.get_children():
		var mmi := child as MultiMeshInstance3D
		if mmi != null:
			total += mmi.multimesh.instance_count
	return total


func _test_build_creates_layer_nodes(dressing: Dictionary) -> void:
	var frame := TownGroundDetail.frame(dressing)
	var lay := TownGroundDetail.layers(dressing, frame)
	var root := TownGroundDetail.build(dressing)
	_assert_true("build returns the ground root", root != null and root.name == TownGroundDetail.ROOT_NAME)
	for pair in [[TownGroundDetail.PLAZA_NAME, "core"], [TownGroundDetail.RIM_NAME, "rim"], [TownGroundDetail.EDGE_NAME, "edge"]]:
		var layer := root.find_child(pair[0], false, false)
		_assert_true("%s layer exists" % pair[0], layer != null)
		if layer != null:
			_assert_true("%s holds one instance per planned cell" % pair[0], _instance_total(layer) == (lay[pair[1]] as Array).size())
	var scat := root.find_child(TownGroundDetail.SCATTER_NAME, false, false)
	_assert_true("scatter layer exists", scat != null)
	if scat != null:
		_assert_true("scatter holds one instance per placement", _instance_total(scat) == TownGroundDetail.scatter(dressing, frame).size())
	var disabled := dressing.duplicate(true)
	var disabled_scatter: Dictionary = disabled["scatter"]
	var disabled_edge: Dictionary = disabled["edge"]
	disabled_scatter["enabled"] = false
	disabled_edge["enabled"] = false
	var quiet := TownGroundDetail.build(disabled)
	_assert_true("disabled layers are not built", quiet.find_child(TownGroundDetail.SCATTER_NAME, false, false) == null and quiet.find_child(TownGroundDetail.EDGE_NAME, false, false) == null)
	quiet.free()
	root.free()
```

Extend `client/tests/test_town_dressing.gd` inside `_test_sync_attaches_world_aligned_in_town_only` after the `root != null` assertion:

```gdscript
		_assert_true("dressing root holds the ground detail", root.find_child(TownGroundDetail.ROOT_NAME, false, false) != null)
```

- [ ] **Step 2: Run to confirm failure**

Run: `gd res://tests/test_town_ground_detail.gd`
Expected: error `Invalid call. Nonexistent function 'build' in base 'GDScript'` (or similar).

- [ ] **Step 3: Append the builders to `town_ground_detail.gd`**

```gdscript
## The whole ground detail under one node, in town world space (x, z = town x, y).
static func build(dressing: Dictionary) -> Node3D:
	var root := Node3D.new()
	root.name = ROOT_NAME
	var frame_data := frame(dressing)
	if float(frame_data["tile"]) <= 0.0:
		return root
	var plaza: Dictionary = dressing.get("plaza", {})
	var scale := float(plaza.get("tile_scale", 1.0))
	var base_id := str(((plaza.get("tile_variants", []) as Array)[0] as Dictionary).get("asset_id", ""))
	var base_box := LibraryScript.bounds(base_id)
	var lay := layers(dressing, frame_data)
	var rim_cfg: Dictionary = plaza.get("rim", {})
	var edge_cfg: Dictionary = dressing.get("edge", {})
	var plaza_y := float(plaza.get("surface_y", 0.0))
	_add(root, _tile_layer(PLAZA_NAME, lay[LAYER_CORE], plaza.get("tile_variants", []), base_box, scale, plaza_y, SALT_CORE))
	_add(root, _tile_layer(RIM_NAME, lay[LAYER_RIM], rim_cfg.get("tile_variants", []), base_box, scale, plaza_y, SALT_RIM))
	_add(root, _tile_layer(EDGE_NAME, lay[LAYER_EDGE], edge_cfg.get("tile_variants", []), base_box, scale, float(edge_cfg.get("surface_y", plaza_y)), SALT_EDGE))
	_add(root, _scatter_layer(dressing, frame_data, base_box))
	return root


static func _add(root: Node3D, layer: Node3D) -> void:
	if layer != null:
		root.add_child(layer)


static func _multimesh_instance(node_name: String, mesh: Mesh, transforms: Array) -> MultiMeshInstance3D:
	var mm := MultiMesh.new()
	mm.transform_format = MultiMesh.TRANSFORM_3D
	mm.mesh = mesh
	mm.instance_count = transforms.size()
	for i in transforms.size():
		mm.set_instance_transform(i, transforms[i])
	var instance := MultiMeshInstance3D.new()
	instance.name = node_name
	instance.multimesh = mm
	return instance


## One MultiMesh per variant; the variant and quarter-turn come from DungeonKitFloor.pick(cell, salt).
static func _tile_layer(layer_name: String, cells: Array, variants: Array, base_box: AABB, scale: float, surface_y: float, salt: int) -> Node3D:
	if cells.is_empty() or variants.is_empty():
		return null
	var ids: Array = []
	var weights: Array = []
	var per_variant: Array = []
	for v in variants:
		ids.append(str((v as Dictionary).get("asset_id", "")))
		weights.append(int((v as Dictionary).get("weight", 1)))
		per_variant.append([])
	for cell in cells:
		var choice := KitFloorScript.pick(cell, salt, weights)
		(per_variant[choice.x] as Array).append(Vector3(cell.x, float(choice.y), cell.y))
	var layer := Node3D.new()
	layer.name = layer_name
	for i in ids.size():
		var mesh := LibraryScript.mesh(str(ids[i]))
		var placements: Array = per_variant[i]
		if mesh == null or placements.is_empty():
			continue
		var box := mesh.get_aabb()
		var transforms: Array = []
		for cell3 in placements:
			transforms.append(KitFloorScript.tile_transform(cell3, box, base_box, surface_y, scale))
		layer.add_child(_multimesh_instance("%s_%s" % [layer_name, str(ids[i])], mesh, transforms))
	return layer


## Patches are seated like tiles (quarter turns, top on the plain tile's slab); rocks rest on the surface.
static func _scatter_layer(dressing: Dictionary, frame_data: Dictionary, base_box: AABB) -> Node3D:
	var placements := scatter(dressing, frame_data)
	if placements.is_empty():
		return null
	var surface_y := float((dressing.get("scatter", {}) as Dictionary).get("surface_y", 0.0))
	var by_asset := {}
	for p in placements:
		var id := str(p["asset_id"])
		if not by_asset.has(id):
			by_asset[id] = []
		(by_asset[id] as Array).append(p)
	var layer := Node3D.new()
	layer.name = SCATTER_NAME
	for id in by_asset:
		var mesh := LibraryScript.mesh(str(id))
		if mesh == null:
			continue
		var box := mesh.get_aabb()
		var transforms: Array = []
		for p in by_asset[id]:
			var pos: Vector2 = p["position"]
			if str(p["kind"]) == "patch":
				transforms.append(KitFloorScript.tile_transform(Vector3(pos.x, float(p["yaw_quarter"]), pos.y), box, base_box, surface_y, float(p["scale"])))
			else:
				transforms.append(rock_transform(pos, float(p["yaw_degrees"]), float(p["scale"]), box, surface_y))
		layer.add_child(_multimesh_instance("%s_%s" % [SCATTER_NAME, str(id)], mesh, transforms))
	return layer
```

Note: `TownGroundDetail.build` deliberately reads the first plaza variant for `base_box`, matching `tile_size`. If `plaza.tile_variants` is empty, `tile_size` is 0 and `build` returns the empty root before touching the array.

- [ ] **Step 4: Integrate into `TownDressing`**

In `client/scripts/town_dressing.gd`:
1. Add `const GroundDetailScript := preload("res://scripts/town_ground_detail.gd")`.
2. Change `const PLAZA_NAME := "TownPlaza"` to `const PLAZA_NAME := GroundDetailScript.PLAZA_NAME`.
3. In `build()`, replace the `var plaza := build_plaza(...)` block with `root.add_child(GroundDetailScript.build(cfg))`.
4. Delete `build_plaza` entirely (its logic now lives in `TownGroundDetail._tile_layer`).
5. Replace `in_plaza`'s body with `return p.distance_to(center) <= radius or GroundDetailScript.capsule_contains(p, center, gate, half_width)`.
6. Remove `const KitFloorScript` and `const LibraryScript` only if nothing else in the file uses them: `_make_prop` still uses `LibraryScript.instantiate`, so keep `LibraryScript`; drop `KitFloorScript`.

Expected size: well under the current 140 lines (`wc -l client/scripts/town_dressing.gd`).

- [ ] **Step 5: Run the town tests**

```bash
gd res://tests/test_town_ground_detail.gd
gd res://tests/test_town_dressing.gd
gd res://tests/test_item_visuals.gd
```

Expected: three `[gdtest] PASS` lines. `test_item_visuals` proves the preview still finds `TownPlaza` and the correct prop count.

- [ ] **Step 6: Commit** (`commit-commands:commit`): `feat: v493: build town ground detail into the town dressing`.

---

### Task 5: Visual gate and tuning

**Files:**
- Modify (tuning only): `shared/assets/town_presentation.v0.json`
- Create: `docs/as-built/assets/v493/town-before.png`, `town-after.png` (+ close-ups)

**Interfaces:** consumes the built town; produces final tuned data and the captures the as-built links.

- [ ] **Step 1: Capture the "before"** from a checkout of `main` (`git stash` is not needed; use a worktree) or reuse `docs/as-built/assets/v491/town-after.png`. Copy it to `docs/as-built/assets/v493/town-before.png`.

- [ ] **Step 2: Capture the "after"**

Read `skills/showme/SKILL.md` for the current flags, then run:

```bash
python3 skills/showme/scripts/render_focus.py --focus town
```

Expected: prints a screenshot path under `.artifacts/showme/`. Open the image (Read tool) and inspect: plaza edge, the vendor path (east, y=12), the mystery-seller path (south-east diagonal), the grass.

- [ ] **Step 3: Apply the spec's decision rules**

Judge against the risks the spec named, and tune data only (no code changes):
- Edge looks like pasted brown squares: reduce `edge.width_m` weight of plain dirt, raise `weeds` weight, or widen the rim's decorated/weeds share. If it still fails, say so in the as-built and keep the edge only if it is better than v491, not to force a pass.
- Patches read as pasted rectangles: lower `patch_share_percent`, shrink `scale_max`, or drop the `dirt_large` patch entry and keep the small weeds.
- Rocks clash with the dungeon palette: lower their weight, or empty `rocks` (data only) and record it. The spec allows dropping them.
- Anything overlapping the palisade, a service stand or the vendor/mystery seller: raise `min_clearance_m` / `fence_clearance_m`.
- Camera coverage: `scatter.radius_m` must not exceed what the play camera reveals past the fence; measure from a live capture (`make play` is not needed; the showme town capture uses the play camera framing) and set the radius accordingly.

After each change: `.venv/bin/pytest tools/test_town_dressing.py && gd res://tests/test_town_ground_detail.gd`, then re-capture. Stop after at most four tuning rounds; whatever remains is recorded honestly in the as-built.

- [ ] **Step 4: Save captures**

```bash
mkdir -p docs/as-built/assets/v493
cp <path printed by render_focus> docs/as-built/assets/v493/town-after.png
```

Also save one close-up of the plaza edge and one of a service path (re-run `render_focus` with the focus/zoom flags its SKILL.md documents; if no zoom flag exists, crop with any image tool into `town-edge-closeup.png` and `town-path-closeup.png`).

- [ ] **Step 5: Commit** (`commit-commands:commit`): `feat: v493: tune town ground detail from captures`.

---

### Task 6: Regression, docs and CI

**Files:**
- Create: `docs/as-built/v493_town-floor-detail.md`
- Modify: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, `docs/progress/slice-codename-index.md`, `docs/CODEMAP.md` (Town services row: add `client/scripts/town_ground_detail.gd`, `client/tests/test_town_ground_detail.gd`, `tools/test_town_dressing.py`), `docs/specs/v493_spec-town-floor-detail.md` (Status: Implemented), this plan (Status: Implemented + execution notes)

- [ ] **Step 1: Targeted regression**

```bash
.venv/bin/pytest tools/test_town_dressing.py tools/assets -q
make validate-shared
make validate-assets
make client-unit
make bot-client SCENARIO=<town vendor scenario id> HEADLESS=1
```

For the last command use the client scenarios named in `docs/progress/scenario-catalog.md` for the vendor panel (`15_town_vendor_shop_panel`), the stash panel (`23_account_stash_panel`) and the town teleporter (`07_town_teleporter_auto_approach`); run each once. Expected: all green. These prove the extra ground nodes did not break picking/approach in town.

- [ ] **Step 2: Maintainability ratchet**

Run: `make maintainability`
Expected: pass; `client/scripts/town_dressing.gd` is smaller than before and `town_ground_detail.gd` is under 600 lines (`wc -l`). No baseline changes needed.

- [ ] **Step 3: Measure the draw-call/instance delta** for the as-built: in the Godot town capture, note the count of `MultiMeshInstance3D` nodes under `TownGround` (one per distinct asset) and total instance counts (`_instance_total`), and record both plus the town-entry build time (a one-off `Time.get_ticks_usec()` around `TownGroundDetail.build` in a scratch script, not committed).

- [ ] **Step 4: Write the as-built and update the indexes**

`docs/as-built/v493_town-floor-detail.md`: what shipped, the final tuned data, the captures, the measured instance/draw-call numbers, what the visual gate decided about the edge/rocks (honestly, including anything dropped), and known limits (dirt tiles are not a real terrain blend; the v494 dungeon props follow). Update PROGRESS "Latest completed slice"/"Next slice" (v494 = dungeon props), the lifecycle table row, and the codename index. Mark the spec and this plan `Implemented`.

- [ ] **Step 5: Pre-PR gate**

Run: `make ci`
Expected: green (~6-15 min). The known `paladin_class_foundation` flake is recorded in PROGRESS; re-run that one scenario in isolation (`make bot scenario=<its id>`) before calling it a flake, and say so if it recurs. Do not run `ci-full`.

- [ ] **Step 6: Commit docs** (`commit-commands:commit`): `feat: v493: town floor detail as-built`. Then follow the branch-finishing flow (`superpowers:finishing-a-development-branch`) for the PR.

---

## Self-review against the spec

- Soft plaza edge (dirt band + stone rim): Tasks 2 (data), 3 (layers), 4 (build), 5 (visual).
- Paths to services: Tasks 2 (targets), 3 (`path_segments`, connectivity test), 4.
- Grass variation: Tasks 2, 3 (`scatter` + exclusions), 4.
- Reuse of the v491 seating fix (`tile_transform`) and the `TownDressing` root/ground-offset: Task 4 (`_tile_layer`, patches).
- Removal on non-town levels: unchanged; existing test in `test_town_dressing.gd` still covers it, and the ground root lives under the same `TownDressing` root.
- Loader regression pin: Task 2 Step 7.
- Python gates (anchors == world preset, targets, manifest ids, fence clearance): Task 2.
- Nine manifest entries within D8 budgets: Task 1.
- Maintainability (`town_dressing.gd` shrinks; new file < 600): Tasks 4 and 6.
- `make ci` once, no `ci-full`: Task 6.
- Known deviations from the spec text, on purpose: (1) `width_m` distances are Chebyshev between cell centres (documented in code and tests); (2) `path_width_m` is 3.0, not the spec's implied 2.0, to guarantee connected diagonal paths; (3) `anchors` (14 world-preset gameplay points) is a new data key the spec did not list, needed so the client can avoid gameplay positions and resolve path targets without reading the server's world preset.
