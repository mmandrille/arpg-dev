# Base Model Consolidation + Skeleton Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `base_humanoid.glb` with `barbarian.glb` as the single shared character model across all classes, and fix `hand_r`/`hand_l`/`foot_r`/`foot_l` bone positions so equipment sockets land at the visible hand/foot mesh.

**Architecture:** Two independent tasks. Task 1 is pure code/config changes — no GLB editing. Task 2 creates a stdlib-only Python script that patches the binary GLB by modifying bone node translations and recomputing inverse bind matrices (IBMs). Because all barbarian bones have identity rotation, IBM = inverse translation = negate the accumulated global translation from root. The binary patch is the authoritative fix; `barbarian.glb` is removed from `gen_glb.py`'s regeneration targets so `make gen-assets` never overwrites it.

**Tech Stack:** Python 3 stdlib only (`struct`, `json`), GDScript 4, glTF 2.0 binary format.

## Global Constraints

- `character_base_humanoid_v0` asset ID must be preserved — only `runtime_path` changes to `client/assets/characters/barbarian/barbarian.glb`
- `barbarian.glb` joints (17 bones, nodes 2–18): root, spine, chest, neck, head, arm_l, elbow_l, hand_l, arm_r, elbow_r, hand_r, leg_l, knee_l, foot_l, leg_r, knee_r, foot_r
- IBM accessor is accessor index 23, stored at `bufferView.byteOffset = 294732` in the binary chunk, MAT4 count=17
- All bones have identity rotation → IBM[i] = pure negated-translation matrix (no full matrix inverse needed)
- `gen_glb.py` must NOT list `barbarian.glb` in `_all_targets()` — it is a committed Blender export, not a generated stub
- `make validate-assets` must pass after Task 1; `make ci` must pass after Task 2
- No server, protocol, shared rules, or golden fixture changes

---

## File Map

| File | Action | Task |
|------|--------|------|
| `assets/manifests/assets.v0.json` | Modify | 1 |
| `client/scenes/character.tscn` | Modify | 1 |
| `client/scripts/class_presentations_loader.gd` | Modify | 1 |
| `client/tools/inspect_rig.gd` | Modify | 1 |
| `tools/assets/gen_glb.py` | Modify | 1 |
| `tools/assets/test_validate_assets.py` | Modify | 1 |
| `client/assets/characters/base_humanoid/base_humanoid.glb` | **Delete** | 1 |
| `client/assets/characters/base_humanoid/base_humanoid.glb.import` | **Delete** | 1 |
| `tools/assets/fix_skeleton.py` | **Create** | 2 |
| `client/assets/characters/barbarian/barbarian.glb` | Modify (binary) | 2 |

---

### Task 1: Redirect character_base_humanoid_v0 to barbarian.glb

**Files:**
- Modify: `assets/manifests/assets.v0.json`
- Modify: `client/scenes/character.tscn`
- Modify: `client/scripts/class_presentations_loader.gd`
- Modify: `client/tools/inspect_rig.gd`
- Modify: `tools/assets/gen_glb.py`
- Modify: `tools/assets/test_validate_assets.py`
- Delete: `client/assets/characters/base_humanoid/base_humanoid.glb`
- Delete: `client/assets/characters/base_humanoid/base_humanoid.glb.import`

**Interfaces:**
- Produces: `character_base_humanoid_v0` asset resolves to `client/assets/characters/barbarian/barbarian.glb` at runtime

- [ ] **Step 1: Update asset manifest**

In `assets/manifests/assets.v0.json`, replace the entire `character_base_humanoid_v0` value with:

```json
"character_base_humanoid_v0": {
  "type": "character",
  "format": "glb",
  "scale_unit": "meters",
  "source_path": "assets/characters/barbarian/goliath_barbarian.glb",
  "runtime_path": "client/assets/characters/barbarian/barbarian.glb",
  "required_nodes": [
    "root", "spine", "chest", "neck", "head",
    "arm_l", "elbow_l", "hand_l",
    "arm_r", "elbow_r", "hand_r",
    "leg_l", "knee_l", "foot_l",
    "leg_r", "knee_r", "foot_r"
  ],
  "provenance": {
    "origin": "Redirected to barbarian.glb (shared base humanoid). See character_barbarian_v0 for full provenance.",
    "license": "CC0-1.0",
    "sha256": "80a6ef3d30702c97b51369649ecb6683df9ef49dab292bdb57454e2c2d078a89"
  }
},
```

Note: the sha256 `80a6ef3d30702c97b51369649ecb6683df9ef49dab292bdb57454e2c2d078a89` is the current barbarian.glb hash. Task 2 will update it after the skeleton is patched.

- [ ] **Step 2: Update character.tscn ext_resource**

In `client/scenes/character.tscn` line 8, change:
```
[ext_resource type="PackedScene" path="res://assets/characters/base_humanoid/base_humanoid.glb" id="1_glb"]
```
to:
```
[ext_resource type="PackedScene" path="res://assets/characters/barbarian/barbarian.glb" id="1_glb"]
```

- [ ] **Step 3: Update class_presentations_loader.gd fallback paths**

In `client/scripts/class_presentations_loader.gd`, change both occurrences of:
```
"client/assets/characters/base_humanoid/base_humanoid.glb"
```
to:
```
"client/assets/characters/barbarian/barbarian.glb"
```

(Lines ~39 and ~65 — exact lines may vary, search for the string.)

- [ ] **Step 4: Update inspect_rig.gd**

In `client/tools/inspect_rig.gd` line 4, replace the base_humanoid check entirely:

```gdscript
func _initialize() -> void:
	_check("res://assets/characters/barbarian/barbarian.glb", [
		"root", "spine", "chest", "neck", "head",
		"arm_l", "elbow_l", "hand_l",
		"arm_r", "elbow_r", "hand_r",
		"leg_l", "knee_l", "foot_l",
		"leg_r", "knee_r", "foot_r",
	])
	_check("res://assets/monsters/dummy/monster_dummy.glb", ["root", "pivot"])
	print("[rig-gate] PASS")
	quit(0)
```

- [ ] **Step 5: Remove base_humanoid from gen_glb.py _all_targets()**

In `tools/assets/gen_glb.py`, find the `_all_targets()` function and remove the `base_humanoid` line:

```python
def _all_targets() -> dict:
    from tools.assets import gen_glb_equipment

    return {
        # base_humanoid.glb is a committed Blender export — not generated here.
        # character_base_humanoid_v0 now points to barbarian.glb in assets.v0.json.
        "client/assets/characters/barbarian/barbarian.glb": barbarian_glb,
        "client/assets/characters/sorcerer/sorcerer.glb": sorcerer_glb,
        "client/assets/characters/paladin/paladin.glb": paladin_glb,
        "client/assets/characters/rogue/rogue.glb": rogue_glb,
        "client/assets/characters/ranger/ranger.glb": ranger_glb,
        **gen_glb_equipment.EQUIPMENT_TARGETS,
        "client/assets/monsters/dummy/monster_dummy.glb": monster_dummy_glb,
        "client/assets/monsters/skeleton/monster_skeleton.glb": monster_skeleton_glb,
    }
```

- [ ] **Step 6: Update test_validate_assets.py**

In `tools/assets/test_validate_assets.py`:

Change line 19:
```python
CHAR_GLB = "client/assets/characters/barbarian/barbarian.glb"
```

Find the fixture that defines `character_base_humanoid_v0` `required_nodes` (around line 118) and update to the 17-bone list:
```python
"required_nodes": ["root", "spine", "chest", "neck", "head",
                   "arm_l", "elbow_l", "hand_l",
                   "arm_r", "elbow_r", "hand_r",
                   "leg_l", "knee_l", "foot_l",
                   "leg_r", "knee_r", "foot_r"],
```

Any test that checks for `["root", "spine", "arm_l", "hand_l", "arm_r", "hand_r", "leg_l", "leg_r"]` (the old 8-bone list) must be updated to the 17-bone list above.

- [ ] **Step 7: Delete base_humanoid runtime GLB files**

```bash
git rm client/assets/characters/base_humanoid/base_humanoid.glb \
       client/assets/characters/base_humanoid/base_humanoid.glb.import
```

- [ ] **Step 8: Run validate-assets to confirm the redirect works**

```bash
make validate-assets 2>&1 | tail -10
```

Expected: passes with no errors.

- [ ] **Step 9: Run Python unit tests**

```bash
.venv/bin/pytest tools/assets/test_validate_assets.py -v 2>&1 | tail -20
```

Expected: all tests pass.

- [ ] **Step 10: Run rig gate**

```bash
godot --headless --path client --script res://tools/inspect_rig.gd 2>&1 | grep -E "PASS|FAIL|ERROR"
```

Expected: `[rig-gate] PASS`

- [ ] **Step 11: Render skeleton viewer to confirm barbarian shows as default**

```bash
python3 skills/showme/scripts/render_focus.py --focus skeleton --output .artifacts/showme/skeleton-base.png 2>&1 | grep saved
```

Expected: screenshot saved. Open and confirm the barbarian model is visible.

- [ ] **Step 12: Commit**

```bash
git add assets/manifests/assets.v0.json \
        client/scenes/character.tscn \
        client/scripts/class_presentations_loader.gd \
        client/tools/inspect_rig.gd \
        tools/assets/gen_glb.py \
        tools/assets/test_validate_assets.py
git commit -m "feat: redirect character_base_humanoid_v0 to barbarian.glb as single shared base model"
```

---

### Task 2: Create fix_skeleton.py and apply first bone position fix

**Files:**
- Create: `tools/assets/fix_skeleton.py`
- Modify (binary): `client/assets/characters/barbarian/barbarian.glb`

**Interfaces:**
- Consumes: `client/assets/characters/barbarian/barbarian.glb` (17-bone Blender export)
- Produces: same GLB with `hand_r`, `hand_l`, `foot_r`, `foot_l` local translations modified and IBMs recomputed

- [ ] **Step 1: Create tools/assets/fix_skeleton.py**

Create `tools/assets/fix_skeleton.py` with the complete content below:

```python
#!/usr/bin/env python3
"""Patch barbarian.glb bone local translations and recompute inverse bind matrices.

Usage:
  python3 tools/assets/fix_skeleton.py           # apply DEFAULT_DELTAS
  python3 tools/assets/fix_skeleton.py --dry-run  # print changes without writing
  python3 tools/assets/fix_skeleton.py --delta hand_r=0,-0.1,0 --delta foot_r=0,-0.05,0

All barbarian bones have identity rotation, so the IBM for each joint is simply:
  [ 1  0  0  -tx ]
  [ 0  1  0  -ty ]
  [ 0  0  1  -tz ]
  [ 0  0  0   1  ]
where (tx, ty, tz) is the joint's accumulated global translation from scene root.
"""
from __future__ import annotations
import argparse, json, struct, sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
GLB_PATH = ROOT / "client/assets/characters/barbarian/barbarian.glb"

# First-iteration deltas: (dx, dy, dz) added to current local_translation.
# hand_r/l: extend further down the arm (more negative Y).
# foot_r/l: extend further below the ankle (more negative Y).
DEFAULT_DELTAS: dict[str, tuple[float, float, float]] = {
    "hand_r": (0.0, -0.195, 0.021),
    "hand_l": (0.0, -0.195, 0.021),
    "foot_r": (0.0, -0.10,  0.0),
    "foot_l": (0.0, -0.10,  0.0),
}


def _read_glb(path: Path) -> tuple[dict, bytearray, int]:
    data = path.read_bytes()
    _magic, _ver, _total = struct.unpack_from("<III", data, 0)
    json_len, _json_type = struct.unpack_from("<II", data, 12)
    gltf = json.loads(data[20 : 20 + json_len])
    bin_offset = 20 + json_len
    bin_len, _bin_type = struct.unpack_from("<II", data, bin_offset)
    bin_data = bytearray(data[bin_offset + 8 : bin_offset + 8 + bin_len])
    return gltf, bin_data, json_len


def _write_glb(path: Path, gltf: dict, bin_data: bytearray) -> None:
    json_bytes = json.dumps(gltf, separators=(",", ":"), sort_keys=True).encode()
    while len(json_bytes) % 4:
        json_bytes += b" "
    while len(bin_data) % 4:
        bin_data += b"\x00"
    json_chunk = struct.pack("<II", len(json_bytes), 0x4E4F534A) + json_bytes
    bin_chunk  = struct.pack("<II", len(bin_data),  0x004E4942) + bytes(bin_data)
    total = 12 + len(json_chunk) + len(bin_chunk)
    header = struct.pack("<III", 0x46546C67, 2, total)
    path.write_bytes(header + json_chunk + bin_chunk)


def _global_translations(nodes: list[dict]) -> dict[int, tuple[float,float,float]]:
    """Return accumulated (tx, ty, tz) for every node by traversing parent chain."""
    parent: dict[int, int] = {}
    for i, n in enumerate(nodes):
        for c in n.get("children", []):
            parent[c] = i

    def _accum(idx: int) -> tuple[float,float,float]:
        t = nodes[idx].get("translation", [0.0, 0.0, 0.0])
        if idx not in parent:
            return (t[0], t[1], t[2])
        px, py, pz = _accum(parent[idx])
        return (px + t[0], py + t[1], pz + t[2])

    return {i: _accum(i) for i in range(len(nodes))}


def _make_ibm(tx: float, ty: float, tz: float) -> list[float]:
    """Column-major 4×4 inverse-translation matrix for a translation-only transform."""
    return [
        1.0, 0.0, 0.0, 0.0,   # col 0
        0.0, 1.0, 0.0, 0.0,   # col 1
        0.0, 0.0, 1.0, 0.0,   # col 2
        -tx, -ty, -tz, 1.0,   # col 3
    ]


def fix_bones(
    glb_path: Path,
    deltas: dict[str, tuple[float, float, float]],
    dry_run: bool = False,
) -> None:
    gltf, bin_data, _json_len = _read_glb(glb_path)
    nodes = gltf["nodes"]
    skin = gltf["skins"][0]
    joints: list[int] = skin["joints"]

    name_to_node = {nodes[j].get("name", ""): j for j in joints}

    changed = False
    for bone_name, (dx, dy, dz) in deltas.items():
        node_idx = name_to_node.get(bone_name)
        if node_idx is None:
            print(f"WARNING: bone '{bone_name}' not found — skipping")
            continue
        t = nodes[node_idx].get("translation", [0.0, 0.0, 0.0])
        new_t = [t[0] + dx, t[1] + dy, t[2] + dz]
        print(f"  {bone_name}: {[round(v, 4) for v in t]} → {[round(v, 4) for v in new_t]}")
        nodes[node_idx]["translation"] = new_t
        changed = True

    if not changed:
        print("No bones modified.")
        return

    # Recompute IBMs for all 17 skin joints using updated translations.
    global_t = _global_translations(nodes)

    ibm_acc_idx = skin["inverseBindMatrices"]
    ibm_bv = gltf["bufferViews"][gltf["accessors"][ibm_acc_idx]["bufferView"]]
    ibm_byte_offset = ibm_bv.get("byteOffset", 0)

    for slot, joint_node in enumerate(joints):
        tx, ty, tz = global_t[joint_node]
        ibm = _make_ibm(tx, ty, tz)
        struct.pack_into("<16f", bin_data, ibm_byte_offset + slot * 64, *ibm)

    if dry_run:
        print("[dry-run] would write updated barbarian.glb")
        return

    gltf["nodes"] = nodes
    _write_glb(glb_path, gltf, bin_data)
    print(f"Written: {glb_path}")


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument(
        "--delta", action="append", default=[], metavar="BONE=dx,dy,dz",
        help="Override delta for a named bone (e.g. hand_r=0,-0.1,0). "
             "Repeat for multiple bones. If none given, DEFAULT_DELTAS is used.",
    )
    args = ap.parse_args()

    if args.delta:
        deltas: dict[str, tuple[float, float, float]] = {}
        for item in args.delta:
            name, xyz = item.split("=", 1)
            vals = tuple(float(v) for v in xyz.split(","))
            if len(vals) != 3:
                print(f"ERROR: --delta {item}: expected 3 floats", file=sys.stderr)
                return 1
            deltas[name] = vals  # type: ignore[assignment]
    else:
        deltas = DEFAULT_DELTAS

    fix_bones(GLB_PATH, deltas, dry_run=args.dry_run)
    return 0


if __name__ == "__main__":
    sys.exit(main())
```

- [ ] **Step 2: Dry-run to verify the script parses correctly**

```bash
python3 tools/assets/fix_skeleton.py --dry-run
```

Expected output (exact translations will show current → adjusted):
```
  hand_r: [0.0, -0.3349, 0.0489] → [0.0, -0.5299, 0.0699]
  hand_l: [0.0, -0.3349, 0.0489] → [0.0, -0.5299, 0.0699]
  foot_r: [0.0, -0.394, 0.0] → [0.0, -0.494, 0.0]
  foot_l: [0.0, -0.394, 0.0] → [0.0, -0.494, 0.0]
[dry-run] would write updated barbarian.glb
```

- [ ] **Step 3: Apply the fix**

```bash
python3 tools/assets/fix_skeleton.py
```

Expected: same output as dry-run but ending with `Written: .../barbarian.glb`

- [ ] **Step 4: Render skeleton viewer for all classes and inspect**

```bash
for class in barbarian sorcerer paladin rogue ranger; do
  python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id $class \
    --output .artifacts/showme/skeleton-$class.png
  echo "saved skeleton-$class.png"
done
```

Open each PNG and verify: `right_hand` / `off_hand` socket markers (blue spheres) have moved lower down the arm (closer to the visible ghost-mesh hand position).

- [ ] **Step 5: Update manifest sha256 for both entries pointing to barbarian.glb**

After patching, the GLB binary has changed and the sha256 is stale. Recompute and update:

```bash
python3 -c "
import hashlib, json
from pathlib import Path

glb = Path('client/assets/characters/barbarian/barbarian.glb')
new_hash = hashlib.sha256(glb.read_bytes()).hexdigest()
print('New sha256:', new_hash)

mf_path = Path('assets/manifests/assets.v0.json')
m = json.loads(mf_path.read_text())
m['assets']['character_barbarian_v0']['provenance']['sha256'] = new_hash
m['assets']['character_base_humanoid_v0']['provenance']['sha256'] = new_hash
mf_path.write_text(json.dumps(m, indent=2, ensure_ascii=False) + '\n')
print('Manifest updated.')
"
```

- [ ] **Step 6: Run validate-assets to confirm sha256 now matches**

```bash
make validate-assets 2>&1 | tail -10
```

Expected: passes with no sha256 mismatch errors.

- [ ] **Step 7: Run CI**

```bash
make ci 2>&1 | tail -20
```

Expected: green.

- [ ] **Step 8: Commit**

```bash
git add tools/assets/fix_skeleton.py \
        client/assets/characters/barbarian/barbarian.glb \
        assets/manifests/assets.v0.json
git commit -m "feat: fix_skeleton.py — extend hand/foot bones to mesh-aligned positions (first iteration)"
```
