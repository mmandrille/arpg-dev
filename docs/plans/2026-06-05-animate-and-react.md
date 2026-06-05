# Animate and React (Slice v3) Implementation Plan

Status: Ready for implementation (2026-06-05) — gaps vs spec/codebase closed in this revision.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prove the `rigged GLB → Skeleton3D → state-driven AnimationPlayer + BoneAttachment3D` pipeline on the player (idle/walk/attack, client-derived) and the monster (hit/death, authoritative-event-driven), with zero server/protocol change.

**Architecture:** Deterministic stdlib generator emits *skinned* GLBs (real `skins`, inverse-bind matrices, `JOINTS_0`/`WEIGHTS_0`); a committed headless GDScript builder emits `AnimationLibrary` `.tres` clips that target skeleton bone poses; one injected `AnimationController` (RefCounted) drives clip state per entity; the equipped weapon rides the `hand_r` bone via a `BoneAttachment3D` named `right_hand_socket`, so the existing `EquipmentVisualResolver` mounts under it unchanged. The client begins reading the already-transmitted `state_delta.events` array; nothing on the server changes.

**Tech Stack:** Python 3 (stdlib `struct`/`json` + jsonschema + pytest), Godot 4.6.3 / GDScript, glTF/GLB, JSON Schema. Go server untouched.

**Spec:** `docs/specs/spec-animate-and-react.md`
**Baseline:** slice v2 `equip-and-see-it` (complete — `make ci` green)
**Branch:** `feature/animate-and-react` (already created; the spec is committed there)

---

## File Structure

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `tools/assets/gen_glb.py` | Add a skinned-mesh builder; emit rigged `base_humanoid.glb` + new `monster_dummy.glb`; sword stays static |
| Modify | `assets/manifests/assets.v0.schema.json` | Add `"monster"` to the asset `type` enum |
| Modify | `assets/manifests/assets.v0.json` | Add `monster_dummy_v0`; character `required_nodes` → rig joints incl. `hand_r`; regenerate `sha256` |
| Modify | `tools/assets/validate_assets.py` | Check [5]: mount-bone `hand_r` in character `required_nodes`; check [6]: skin-joint hard-check |
| Modify | `tools/assets/test_validate_assets.py` | Update v2 fixtures for rig joints + skin GLBs; cover monster + mount-bone + skin-joint pass/fail |
| Create | `client/tools/inspect_rig.gd` | Rig gate: assert both GLBs import as skinned `Skeleton3D` (wired into `client_smoke.sh`) |
| Create | `assets/monsters/dummy/README.md` | Source notes for the generated monster dummy (mirrors character/sword README stubs) |
| Create | `client/tools/build_animations.gd` | Headless builder: emit `AnimationLibrary` `.tres` clips targeting skeleton bone poses |
| Create | `client/animations/character_anims.tres` | Library: `idle`, `walk`, `attack` (generated artifact) |
| Create | `client/animations/monster_anims.tres` | Library: `idle`, `hit`, `death` (generated artifact) |
| Create | `client/scripts/character_visual.gd` | Attaches a `BoneAttachment3D` named `right_hand_socket` to bone `hand_r` |
| Modify | `client/scenes/character.tscn` | Instance rigged GLB; attach `character_visual.gd`; add `AnimationPlayer` with the library |
| Create | `client/scenes/monster_dummy.tscn` | Dummy GLB + `AnimationPlayer` with the monster library |
| Create | `client/scripts/animation_controller.gd` | Injected `AnimationController` state machine (one script, per-entity instances) |
| Create | `client/tests/test_animation.gd` | Headless: rig/scene assertions + controller state-machine logic |
| Modify | `client/scripts/equipment_visuals.gd` | Delete `play_attack_swing()` + tween machinery; mount logic unchanged |
| Modify | `client/scripts/main.gd` | Controllers; locomotion from input; attack one-shot; read `events`; `entities[id] = {node,controller,type}` (incl. `_apply_snapshot` clear); monster uses scene |
| Modify | `client/scripts/smoke.gd` | Player attack-clip assert; monster hit→death event path; resume death pose |
| Modify | `scripts/client_smoke.sh` | Run `inspect_rig.gd` + `test_animation.gd` before the slice smoke |
| Modify | `shared/assets/item_visuals.v0.json` | Re-tune `rusty_sword` `local_transform` if the bone-mounted grip looks wrong (data-only) |
| Modify | `make/shared.mk` | Add `gen-anims` target (regenerate clip libraries) |
| Create | `docs/adr/0007-animation-state-model.md` | Durable decision: client-derived vs event-driven, discrete clips, no protocol change |
| Modify | `docs/adr/0006-asset-pipeline.md` | As-built: `Node3D` socket → `BoneAttachment3D` on `hand_r` |

**Sequencing:** Task 1 (rig) is the highest risk — it gates everything and is verified by import before any animation work builds on it. Tasks 1–2 are Python/asset only. Tasks 3–5 produce Godot assets. Task 6 is the pure controller. Tasks 7–8 wire the client. Tasks 9–10 test. Task 11 is docs + final gate.

---

## Task 1: Skinned rig generator

Rewrites `gen_glb.py` to emit **skinned** humanoid + dummy GLBs (the sword stays a static mesh). Uses *rigid* skinning — every vertex of a body part is weighted 100% to one joint — which is the simplest valid glTF skin and is byte-deterministic.

**Files:**
- Modify: `tools/assets/gen_glb.py`
- Modify (regenerated bytes): `client/assets/characters/base_humanoid/base_humanoid.glb`
- Create (regenerated bytes): `client/assets/monsters/dummy/monster_dummy.glb`

- [ ] **Step 1.1: Add the skinned-GLB builder to `gen_glb.py`**

Insert this function after `_build_glb` (it reuses `_cube_geometry`, `_pad`, and the `_FACES` data already in the file). A "joint" is `(name, parent_index, local_translation)`; a "part" is a cube `(joint_index, translation, scale)` whose 24 vertices are baked into mesh space and weighted to that joint.

```python
def _mat4_inverse_translation(t):
    # Joints here are translation-only, so the inverse bind matrix is just a
    # translation by -t. glTF stores matrices column-major, 16 floats.
    x, y, z = t
    return [
        1.0, 0.0, 0.0, 0.0,
        0.0, 1.0, 0.0, 0.0,
        0.0, 0.0, 1.0, 0.0,
        -x, -y, -z, 1.0,
    ]


def _joint_globals(joints):
    # joints: [(name, parent_idx, (lx,ly,lz))]; parent_idx == -1 for root.
    globals_ = []
    for _name, parent, local in joints:
        if parent < 0:
            globals_.append(tuple(local))
        else:
            pg = globals_[parent]
            globals_.append((pg[0] + local[0], pg[1] + local[1], pg[2] + local[2]))
    return globals_


def _build_skinned_glb(color, joints, parts):
    """Build a .glb with a single skinned mesh + a skeleton.

    joints: [(name, parent_idx, (lx,ly,lz))]  -- translation-only bind pose
    parts:  [(joint_idx, (tx,ty,tz), (sx,sy,sz))] -- a cube baked into mesh space,
            fully weighted (1.0) to joint_idx.
    """
    cube_pos, cube_nrm, cube_idx = _cube_geometry()
    globals_ = _joint_globals(joints)

    positions, normals, indices, joints0, weights0 = [], [], [], [], []
    for joint_idx, (tx, ty, tz), (sx, sy, sz) in parts:
        base = len(positions)
        for (px, py, pz), n in zip(cube_pos, cube_nrm):
            positions.append((px * sx + tx, py * sy + ty, pz * sz + tz))
            normals.append(n)
            joints0.append((joint_idx, 0, 0, 0))
            weights0.append((1.0, 0.0, 0.0, 0.0))
        for i in cube_idx:
            indices.append(base + i)

    bin_buf = bytearray()
    pos_off = len(bin_buf)
    for p in positions:
        bin_buf += struct.pack("<fff", *p)
    nrm_off = len(bin_buf)
    for n in normals:
        bin_buf += struct.pack("<fff", *n)
    j_off = len(bin_buf)
    for j in joints0:
        bin_buf += struct.pack("<HHHH", *j)
    w_off = len(bin_buf)
    for w in weights0:
        bin_buf += struct.pack("<ffff", *w)
    idx_off = len(bin_buf)
    for i in indices:
        bin_buf += struct.pack("<H", i)
    ibm_off = len(bin_buf)
    for g in globals_:
        bin_buf += struct.pack("<16f", *_mat4_inverse_translation(g))
    _pad(bin_buf)

    pmin = [min(c[i] for c in positions) for i in range(3)]
    pmax = [max(c[i] for c in positions) for i in range(3)]

    # Joint nodes first (indices 0..n-1) so skin.joints == range(n); then mesh
    # node; then a scene root that parents the root joint and the mesh.
    nodes = []
    for idx, (name, parent, local) in enumerate(joints):
        node = {"name": name, "translation": list(local)}
        children = [k for k, jt in enumerate(joints) if jt[1] == idx]
        if children:
            node["children"] = children
        nodes.append(node)
    mesh_node_idx = len(nodes)
    nodes.append({"name": "Mesh", "mesh": 0, "skin": 0})
    root_joint_idx = next(k for k, jt in enumerate(joints) if jt[1] == -1)

    gltf = {
        "asset": {"version": "2.0", "generator": "arpg-dev/tools/assets/gen_glb.py"},
        "scene": 0,
        "scenes": [{"nodes": [root_joint_idx, mesh_node_idx]}],
        "nodes": nodes,
        "meshes": [{
            "primitives": [{
                "attributes": {"POSITION": 0, "NORMAL": 1, "JOINTS_0": 3, "WEIGHTS_0": 4},
                "indices": 2,
                "material": 0,
                "mode": 4,
            }],
        }],
        "skins": [{"joints": list(range(len(joints))), "inverseBindMatrices": 5}],
        "materials": [{
            "pbrMetallicRoughness": {
                "baseColorFactor": list(color),
                "metallicFactor": 0.0,
                "roughnessFactor": 0.9,
            },
        }],
        "accessors": [
            {"bufferView": 0, "componentType": 5126, "count": len(positions), "type": "VEC3", "min": pmin, "max": pmax},
            {"bufferView": 1, "componentType": 5126, "count": len(normals), "type": "VEC3"},
            {"bufferView": 2, "componentType": 5123, "count": len(indices), "type": "SCALAR"},
            {"bufferView": 3, "componentType": 5123, "count": len(joints0), "type": "VEC4"},
            {"bufferView": 4, "componentType": 5126, "count": len(weights0), "type": "VEC4"},
            {"bufferView": 5, "componentType": 5126, "count": len(joints), "type": "MAT4"},
        ],
        "bufferViews": [
            {"buffer": 0, "byteOffset": pos_off, "byteLength": nrm_off - pos_off, "target": 34962},
            {"buffer": 0, "byteOffset": nrm_off, "byteLength": j_off - nrm_off, "target": 34962},
            {"buffer": 0, "byteOffset": idx_off, "byteLength": ibm_off - idx_off, "target": 34963},
            {"buffer": 0, "byteOffset": j_off, "byteLength": w_off - j_off, "target": 34962},
            {"buffer": 0, "byteOffset": w_off, "byteLength": idx_off - w_off, "target": 34962},
            {"buffer": 0, "byteOffset": ibm_off, "byteLength": len(joints) * 64},
        ],
        "buffers": [{"byteLength": len(bin_buf)}],
    }

    json_bytes = bytearray(json.dumps(gltf, sort_keys=True, separators=(",", ":")).encode("utf-8"))
    while len(json_bytes) % 4 != 0:
        json_bytes.append(0x20)
    json_chunk = struct.pack("<II", len(json_bytes), 0x4E4F534A) + bytes(json_bytes)
    bin_chunk = struct.pack("<II", len(bin_buf), 0x004E4942) + bytes(bin_buf)
    total = 12 + len(json_chunk) + len(bin_chunk)
    header = b"glTF" + struct.pack("<II", 2, total)
    return header + json_chunk + bin_chunk
```

> Note: bufferView byte offsets above are derived from the write order
> (`pos, nrm, joints, weights, indices, ibm`); the `byteOffset`/`byteLength`
> pairs reference the correct slices even though the JSON lists them by accessor
> order. Keep the write order exactly as shown.

- [ ] **Step 1.2: Replace `base_humanoid_glb()` with a rigged version**

Replace the existing `base_humanoid_glb` function body with:

```python
def base_humanoid_glb() -> bytes:
    """Low-poly blue-grey humanoid (~1.9 m) as a SKINNED rig.

    Joints (translation-only bind): arm_r pivots at the shoulder so the attack
    clip swings the arm; hand_r is a child of arm_r (the weapon mount) so a
    BoneAttachment3D on hand_r rides the swing. leg_l/leg_r drive the walk clip.
    """
    color = (0.55, 0.62, 0.72, 1.0)
    joints = [
        ("root", -1, (0.0, 0.0, 0.0)),     # 0
        ("spine", 0, (0.0, 1.15, 0.0)),    # 1  global (0,1.15,0)
        ("arm_r", 1, (0.42, 0.35, 0.0)),   # 2  global (0.42,1.5,0) = shoulder
        ("hand_r", 2, (0.0, -0.68, 0.12)), # 3  global (0.42,0.82,0.12) = hand
        ("leg_l", 0, (-0.16, 0.9, 0.0)),   # 4
        ("leg_r", 0, (0.16, 0.9, 0.0)),    # 5
    ]
    parts = [
        (1, (0.0, 1.15, 0.0), (0.5, 0.8, 0.3)),     # torso -> spine
        (1, (0.0, 1.78, 0.0), (0.34, 0.34, 0.34)),  # head  -> spine
        (1, (-0.42, 1.15, 0.0), (0.16, 0.72, 0.16)),# left arm -> spine
        (2, (0.42, 1.15, 0.0), (0.16, 0.72, 0.16)), # right arm -> arm_r
        (4, (-0.16, 0.45, 0.0), (0.2, 0.9, 0.2)),   # left leg -> leg_l
        (5, (0.16, 0.45, 0.0), (0.2, 0.9, 0.2)),    # right leg -> leg_r
    ]
    return _build_skinned_glb(color, joints, parts)


def monster_dummy_glb() -> bytes:
    """Training dummy: a post on a base, skinned so hit/death clips rotate it.

    pivot sits at the base; the post is weighted to pivot so a rotation about
    the base topples the dummy (death) or wobbles it (hit).
    """
    color = (0.62, 0.34, 0.34, 1.0)
    joints = [
        ("root", -1, (0.0, 0.0, 0.0)),   # 0  base
        ("pivot", 0, (0.0, 0.1, 0.0)),   # 1  just above the base
    ]
    parts = [
        (0, (0.0, 0.1, 0.0), (0.9, 0.2, 0.9)),  # base slab -> root
        (1, (0.0, 0.95, 0.0), (0.35, 1.5, 0.35)),  # post -> pivot
    ]
    return _build_skinned_glb(color, joints, parts)
```

- [ ] **Step 1.3: Register the new target + monster source stub**

Update the `TARGETS` dict:

```python
TARGETS = {
    "client/assets/characters/base_humanoid/base_humanoid.glb": base_humanoid_glb,
    "client/assets/equipment/weapons/rusty_sword/rusty_sword.glb": rusty_sword_glb,
    "client/assets/monsters/dummy/monster_dummy.glb": monster_dummy_glb,
}
```

Create `assets/monsters/dummy/README.md` (mirrors the character/sword source stubs):
runtime bytes live under `client/assets/monsters/dummy/`; authorship is
`tools/assets/gen_glb.py`. Update the `gen_glb.py` module docstring to describe
skinned rigs (v3) instead of empty socket nodes (v2).

- [ ] **Step 1.4: Regenerate the GLBs**

Run: `make gen-assets`
Expected: prints `wrote client/assets/characters/base_humanoid/base_humanoid.glb (...)`, the sword, and `wrote client/assets/monsters/dummy/monster_dummy.glb (...)`.

- [ ] **Step 1.5: RIG GATE — import in Godot and introspect the skeleton (fail-fast)**

This is the highest-risk gate: bad skin math imports as a collapsed/empty mesh.

Create `client/tools/inspect_rig.gd`:

```gdscript
extends SceneTree
# Rig gate (spec §10): confirm both rigged GLBs import as real skinned scenes.
func _initialize() -> void:
    _check("res://assets/characters/base_humanoid/base_humanoid.glb", ["root", "spine", "arm_r", "hand_r", "leg_l", "leg_r"])
    _check("res://assets/monsters/dummy/monster_dummy.glb", ["root", "pivot"])
    print("[rig-gate] PASS")
    quit(0)

func _check(path: String, expected_bones: Array) -> void:
    var packed = load(path)
    if packed == null:
        _fail("cannot load %s" % path)
    var scene = (packed as PackedScene).instantiate()
    var skel := scene.find_child("Skeleton3D", true, false) as Skeleton3D
    if skel == null:
        _fail("no Skeleton3D in %s (skin import failed)" % path)
    var names := []
    for i in range(skel.get_bone_count()):
        names.append(skel.get_bone_name(i))
    for b in expected_bones:
        if not names.has(b):
            _fail("%s missing bone %s (have %s)" % [path, b, names])
    var mesh := scene.find_child("*", true, false)  # ensure tree is non-trivial
    print("[rig-gate] %s bones=%s skeleton_path=%s" % [path, names, scene.get_path_to(skel)])

func _fail(msg: String) -> void:
    printerr("[rig-gate] FAIL: ", msg)
    quit(1)
```

Run:
```bash
godot --headless --path client --import
godot --headless --path client --script res://tools/inspect_rig.gd
```
Expected: `[rig-gate] PASS`, and the printed `skeleton_path=` lines (note them — Tasks 3/4 use that path; the conventional value is `Skeleton3D`). If FAIL with "no Skeleton3D", the skin data is wrong — debug the generator before continuing. This script is **not** throwaway — Task 9 wires it into `client_smoke.sh` as the CI rig gate (spec §10).

- [ ] **Step 1.6: Commit**

```bash
git add tools/assets/gen_glb.py assets/monsters/dummy/README.md \
        client/assets/characters/base_humanoid/base_humanoid.glb \
        client/assets/monsters/dummy/monster_dummy.glb client/tools/inspect_rig.gd
git add client/assets/monsters/dummy/monster_dummy.glb.import client/assets/characters/base_humanoid/base_humanoid.glb.import 2>/dev/null || true
git commit -m "feat(assets): skinned humanoid + dummy rigs (gen_glb skin support)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Manifest schema, data, and validator skin-joint check

**Files:**
- Modify: `assets/manifests/assets.v0.schema.json`
- Modify: `assets/manifests/assets.v0.json`
- Modify: `tools/assets/validate_assets.py`
- Test: `tools/assets/test_validate_assets.py`

- [ ] **Step 2.1: Add `"monster"` to the asset `type` enum**

In `assets/manifests/assets.v0.schema.json`, change:

```json
          "type": { "type": "string", "enum": ["character", "equipment"] },
```
to:
```json
          "type": { "type": "string", "enum": ["character", "equipment", "monster"] },
```

- [ ] **Step 2.2: Update the manifest data**

In `assets/manifests/assets.v0.json`, set the character's `required_nodes` to the rig joints and add the monster entry. Use placeholder all-zero shas for now; Step 2.7 fixes them.

Change `character_base_humanoid_v0`'s `required_nodes` to:
```json
      "required_nodes": ["root", "spine", "arm_r", "hand_r", "leg_l", "leg_r"],
```

Add this entry inside `"assets"` (after `weapon_rusty_sword_v0`):
```json
    "monster_dummy_v0": {
      "type": "monster",
      "source_path": "assets/monsters/dummy/monster_dummy.blend",
      "runtime_path": "client/assets/monsters/dummy/monster_dummy.glb",
      "format": "glb",
      "scale_unit": "meters",
      "required_nodes": ["root", "pivot"],
      "provenance": {
        "origin": "generated by tools/assets/gen_glb.py (stdlib, byte-deterministic)",
        "license": "CC0-1.0",
        "sha256": "0000000000000000000000000000000000000000000000000000000000000000"
      }
    }
```

- [ ] **Step 2.3: Teach the validator to read GLB skin joints**

In `tools/assets/validate_assets.py`, add this function next to `parse_glb_node_names`:

```python
def parse_glb_skin_joint_names(path: Path) -> set[str] | None:
    """Return the set of node names referenced by any skin's `joints`, or None.

    A required bone must be an actual skin joint, not merely a named node — this
    is what proves the GLB is skinned (spec §6), not a v2 socket placeholder.
    """
    try:
        data = path.read_bytes()
        if len(data) < 20 or data[0:4] != b"glTF":
            return None
        chunk_len, chunk_type = struct.unpack_from("<II", data, 12)
        if chunk_type != 0x4E4F534A:  # 'JSON'
            return None
        gltf = json.loads(data[20 : 20 + chunk_len].decode("utf-8"))
        nodes = gltf.get("nodes", [])
        joint_idx: set[int] = set()
        for skin in gltf.get("skins", []):
            joint_idx.update(skin.get("joints", []))
        return {nodes[i]["name"] for i in joint_idx if i < len(nodes) and "name" in nodes[i]}
    except Exception:  # noqa: BLE001
        return None
```

- [ ] **Step 2.4: Replace check [5] mount-socket logic with mount-bone coverage**

v2 check [5] required `required_nodes` to include every `item_visuals.mount_socket` name (e.g. `right_hand_socket`). v3 moves the socket out of the GLB — `character_visual.gd` creates a runtime `BoneAttachment3D` named `right_hand_socket` on bone `hand_r` (spec §4.3, §5.5). If check [5] is left unchanged, `make validate-assets` fails immediately after Step 2.2.

In `validate_assets.py`, replace the entire `# [5] character required_nodes cover every referenced mount socket ...` block with:

```python
    # [5] character mount-bone coverage (spec §4.3): item_visuals still names the
    #     runtime socket right_hand_socket, but the manifest required_nodes list
    #     rig joints. The weapon mount contract is satisfied when hand_r is declared.
    print("[5] character mount-bone coverage")
    WEAPON_MOUNT_BONE = "hand_r"
    for asset_id, entry in sorted(characters.items()):
        declared = set(entry.get("required_nodes", []))
        if WEAPON_MOUNT_BONE not in declared:
            report.fail(
                "mount bone",
                f"{asset_id}: required_nodes missing weapon mount bone {WEAPON_MOUNT_BONE}",
            )
        else:
            report.ok(f"{asset_id} declares weapon mount bone {WEAPON_MOUNT_BONE}")
```

Also update the module docstring bullet for check [4]/[5] to describe mount-bone coverage instead of socket-name coverage.

- [ ] **Step 2.5: Replace check [6] with a skin-joint hard-check**

In `validate_assets.py`, replace the entire `# [6] best-effort GLB node-name inspection ...` block with:

```python
    # [6] GLB skin-joint inspection: required_nodes must be SKIN JOINTS, proving
    #     the GLB is actually rigged (spec §6, §10). Characters/monsters are
    #     skinned; equipment (the sword) is static and declares no required_nodes.
    print("[6] GLB skin-joint inspection")
    for asset_id, entry in sorted(assets.items()):
        required = entry.get("required_nodes", [])
        if not required:
            continue
        rt = root / entry["runtime_path"]
        if not rt.is_file():
            continue  # already failed in [2]
        joints = parse_glb_skin_joint_names(rt)
        if joints is None:
            report.fail("glb skin", f"{asset_id}: could not parse GLB skin joints")
            continue
        absent = [n for n in required if n not in joints]
        if absent:
            report.fail("glb joint", f"{asset_id}: required_nodes not skin joints: {absent}")
        else:
            report.ok(f"{asset_id} GLB skin includes joints {required}")
```

- [ ] **Step 2.6: Update existing pytest fixtures + add new cases**

The test file uses `build_root(tmp_path)` — **not** a `tmp_repo` fixture. Several v2 tests break once checks [5]/[6] change:

1. **Update `default_manifest()`** — character `required_nodes` becomes rig joints:
   ```python
   "required_nodes": ["root", "spine", "arm_r", "hand_r", "leg_l", "leg_r"],
   ```
2. **Add `MONSTER_GLB`** constant and a `make_skinned_glb(joint_names: list[str])` helper that emits a minimal valid glTF skin (non-empty `skins`, `JOINTS_0`/`WEIGHTS_0`, inverse bind matrices) so headless tests do not depend on Godot import. Use it in `build_root()` for the character GLB and optionally a monster stub.
3. **Update `test_socket_coverage_failure`** — remove `hand_r` from `required_nodes` and assert a `mount bone` failure (not the old socket-name message).
4. **Rename/update `test_glb_missing_required_node`** — assert a `glb joint` / `not skin joints` failure when a declared joint is absent from the skin.
5. **Append new tests:**

```python
MONSTER_GLB = "client/assets/monsters/dummy/monster_dummy.glb"


def test_monster_entry_passes(tmp_path):
    manifest = default_manifest()
    manifest["assets"]["monster_dummy_v0"] = {
        "type": "monster",
        "runtime_path": MONSTER_GLB,
        "format": "glb",
        "required_nodes": ["root", "pivot"],
    }
    root = build_root(tmp_path, manifest=manifest)
    write(root / MONSTER_GLB, make_skinned_glb(["root", "pivot"]))
    report = run(root)
    assert report.failures == []


def test_required_node_not_a_skin_joint_fails(tmp_path):
    manifest = default_manifest()
    manifest["assets"]["character_base_humanoid_v0"]["required_nodes"] = ["not_a_joint"]
    root = build_root(tmp_path, manifest=manifest, char_nodes=["root", "hand_r"])
    report = run(root)
    assert any("not skin joints" in f or "not_a_joint" in f for f in report.failures)
```

- [ ] **Step 2.7: Fix the manifest sha + run everything green**

```bash
make gen-assets   # ensure bytes are current
# print the real shas:
python3 -c "import hashlib,pathlib; [print(p, hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest()) for p in ['client/assets/characters/base_humanoid/base_humanoid.glb','client/assets/monsters/dummy/monster_dummy.glb']]"
```
Paste each printed sha into the matching `provenance.sha256` in `assets.v0.json` (character + monster).

Run:
```bash
make validate-assets
.venv/bin/python -m pytest -q tools/assets
```
Expected: `ASSET VALIDATION OK` and all pytest green.

- [ ] **Step 2.8: Commit**

```bash
git add assets/manifests/ tools/assets/validate_assets.py tools/assets/test_validate_assets.py
git commit -m "feat(assets): manifest monster entry + validator skin-joint hard-check

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Animation clip builder + libraries

A committed headless GDScript builder emits `AnimationLibrary` `.tres` files whose tracks target skeleton bone poses (rotation_3d). The builder is the source-of-truth (analogous to `gen_glb.py`); the `.tres` are committed artifacts. It discovers the skeleton path at build time so no node path is hard-coded.

**Files:**
- Create: `client/tools/build_animations.gd`
- Create: `client/animations/character_anims.tres`
- Create: `client/animations/monster_anims.tres`
- Modify: `make/shared.mk`

- [ ] **Step 3.1: Write the builder**

Create `client/tools/build_animations.gd`:

```gdscript
extends SceneTree
# Source-of-truth for the committed AnimationLibrary .tres clips (spec §4.4).
# Builds rotation_3d bone-pose tracks against the imported skeleton and saves a
# library next to client/animations/. Run via `make gen-anims`. Deterministic
# for a pinned Godot. Clip motion is crude on purpose (art is a non-goal).
const DEG := PI / 180.0

func _initialize() -> void:
    _build(
        "res://assets/characters/base_humanoid/base_humanoid.glb",
        "res://animations/character_anims.tres",
        _character_clips())
    _build(
        "res://assets/monsters/dummy/monster_dummy.glb",
        "res://animations/monster_anims.tres",
        _monster_clips())
    print("[build-anims] PASS")
    quit(0)

func _build(glb_path: String, out_path: String, clips: Dictionary) -> void:
    var scene = (load(glb_path) as PackedScene).instantiate()
    var skel := scene.find_child("Skeleton3D", true, false) as Skeleton3D
    if skel == null:
        _fail("no Skeleton3D in %s" % glb_path)
    var skel_path := str(scene.get_path_to(skel))  # e.g. "Skeleton3D"
    var lib := AnimationLibrary.new()
    for clip_name in clips:
        lib.add_animation(clip_name, _make_anim(skel_path, clips[clip_name]))
    var err := ResourceSaver.save(lib, out_path)
    if err != OK:
        _fail("save %s failed: %d" % [out_path, err])
    print("[build-anims] wrote %s (skeleton=%s)" % [out_path, skel_path])

func _make_anim(skel_path: String, spec: Dictionary) -> Animation:
    # spec: { "length": float, "loop": bool, "bones": { bone: [[t, x,y,z(deg)], ...] } }
    var a := Animation.new()
    a.length = spec["length"]
    a.loop_mode = Animation.LOOP_LINEAR if spec.get("loop", false) else Animation.LOOP_NONE
    for bone in spec["bones"]:
        var ti := a.add_track(Animation.TYPE_ROTATION_3D)
        a.track_set_path(ti, NodePath("%s:%s" % [skel_path, bone]))
        for key in spec["bones"][bone]:
            var t: float = key[0]
            var q := Quaternion.from_euler(Vector3(key[1] * DEG, key[2] * DEG, key[3] * DEG))
            a.rotation_track_insert_key(ti, t, q)
    return a

func _character_clips() -> Dictionary:
    return {
        # idle: a near-still pose (one identity key so the track exists).
        "idle": {"length": 1.0, "loop": true, "bones": {"spine": [[0.0, 0, 0, 0]]}},
        # walk: alternate the legs back/forth.
        "walk": {"length": 0.8, "loop": true, "bones": {
            "leg_l": [[0.0, 25, 0, 0], [0.4, -25, 0, 0], [0.8, 25, 0, 0]],
            "leg_r": [[0.0, -25, 0, 0], [0.4, 25, 0, 0], [0.8, -25, 0, 0]],
        }},
        # attack: swing the right arm down and back (the weapon rides hand_r).
        "attack": {"length": 0.35, "loop": false, "bones": {
            "arm_r": [[0.0, 0, 0, 0], [0.12, -110, 0, 0], [0.35, 0, 0, 0]],
        }},
    }

func _monster_clips() -> Dictionary:
    return {
        "idle": {"length": 1.0, "loop": true, "bones": {"pivot": [[0.0, 0, 0, 0]]}},
        # hit: a quick wobble about the base.
        "hit": {"length": 0.3, "loop": false, "bones": {
            "pivot": [[0.0, 0, 0, 0], [0.1, 0, 0, 18], [0.3, 0, 0, 0]],
        }},
        # death: topple over and hold (terminal pose).
        "death": {"length": 0.6, "loop": false, "bones": {
            "pivot": [[0.0, 0, 0, 0], [0.6, 0, 0, 88]],
        }},
    }

func _fail(msg: String) -> void:
    printerr("[build-anims] FAIL: ", msg)
    quit(1)
```

- [ ] **Step 3.2: Add the `gen-anims` make target**

In `make/shared.mk`, after the `gen-assets` target, add:

```makefile
.PHONY: gen-anims
gen-anims: ## Regenerate committed AnimationLibrary .tres clips (requires Godot)
	$(GODOT) --headless --path client --import >/dev/null 2>&1 || true
	$(GODOT) --headless --path client --script res://tools/build_animations.gd
```

Also add `gen-anims` to the `.PHONY` line at the top of the file alongside `gen-assets`.

- [ ] **Step 3.3: Generate the libraries**

```bash
mkdir -p client/animations
make gen-anims
```
Expected: `[build-anims] wrote res://animations/character_anims.tres (skeleton=Skeleton3D)`, same for monster, then `[build-anims] PASS`. Two `.tres` files now exist under `client/animations/`.

- [ ] **Step 3.4: Commit**

```bash
git add client/tools/build_animations.gd client/animations/ make/shared.mk
git commit -m "feat(client): animation clip builder + character/monster libraries

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Character scene — bone socket + AnimationPlayer

**Files:**
- Create: `client/scripts/character_visual.gd`
- Modify: `client/scenes/character.tscn`

- [ ] **Step 4.1: Write the socket-attach script**

Create `client/scripts/character_visual.gd`. It attaches the `BoneAttachment3D` named `right_hand_socket` to `hand_r` at runtime by finding the skeleton by type — no hard-coded node path, mirroring the resolver's name-based decoupling.

```gdscript
extends Node3D
# Attaches the weapon mount socket to the rig's hand bone (spec §5.5).
# Named "right_hand_socket" so EquipmentVisualResolver.find_child() resolves it
# unchanged. Created in code (not the .tscn) to stay robust against the exact
# imported skeleton node path.

const MOUNT_BONE := "hand_r"
const SOCKET_NAME := "right_hand_socket"


func _ready() -> void:
    var skel := find_child("Skeleton3D", true, false) as Skeleton3D
    if skel == null:
        push_warning("[character] no Skeleton3D; cannot attach %s" % SOCKET_NAME)
        return
    if skel.find_child(SOCKET_NAME, false, false) != null:
        return  # already attached (e.g. duplicated instance)
    var att := BoneAttachment3D.new()
    att.name = SOCKET_NAME
    skel.add_child(att)
    att.bone_name = MOUNT_BONE
    if att.bone_idx < 0:
        push_warning("[character] bone %s not found on skeleton" % MOUNT_BONE)
```

- [ ] **Step 4.2: Rewrite `character.tscn`**

Replace the contents of `client/scenes/character.tscn` with (the GLB now carries a skeleton; the `AnimationPlayer` uses the generated library; `root_node` points at the GLB instance so library track paths resolve):

```
[gd_scene load_steps=4 format=3]

; Slice v3 player character (spec animate-and-react §5.1/§5.5). The rigged GLB is
; instanced as "ModelRoot" (carries Skeleton3D with bone hand_r). character_visual.gd
; attaches a BoneAttachment3D named "right_hand_socket" to hand_r so the equipped
; weapon rides the arm. AnimationPlayer plays idle/walk/attack from the committed
; library; root_node targets ModelRoot so bone-pose tracks resolve.
[ext_resource type="PackedScene" path="res://assets/characters/base_humanoid/base_humanoid.glb" id="1_glb"]
[ext_resource type="Script" path="res://scripts/character_visual.gd" id="2_script"]
[ext_resource type="AnimationLibrary" path="res://animations/character_anims.tres" id="3_anims"]

[node name="CharacterVisual" type="Node3D"]
script = ExtResource("2_script")

[node name="ModelRoot" parent="." instance=ExtResource("1_glb")]

[node name="AnimationPlayer" type="AnimationPlayer" parent="."]
root_node = NodePath("../ModelRoot")
libraries = {
"": ExtResource("3_anims")
}
```

> **Do not hard-code a `uid://` on the GLB ext_resource** — regenerating
> `base_humanoid.glb` in Task 1 changes the import UID. Omit it (Godot rewrites
> on save) or let the editor assign one after `--import`. The v2 scene used
> `uid://c47pdy6tnkrny`; that value will be stale after the rig rewrite.

> If Step 1.5 reported a `skeleton_path` other than `Skeleton3D`, the script
> still works (it searches by type). The `root_node` only needs to be the GLB
> instance, which it is.

- [ ] **Step 4.3: Verify the scene wires up headless**

Create a temporary check and run it (delete after):

```bash
cat > client/tools/_check_char.gd <<'EOF'
extends SceneTree
func _initialize():
    var s = (load("res://scenes/character.tscn") as PackedScene).instantiate()
    get_root().add_child(s)  # triggers _ready -> socket attach
    var sock = s.find_child("right_hand_socket", true, false)
    var ap = s.find_child("AnimationPlayer", true, false) as AnimationPlayer
    assert(sock is BoneAttachment3D, "socket missing/not BoneAttachment3D")
    assert(sock.bone_name == "hand_r", "socket not bound to hand_r")
    assert(ap.has_animation("idle") and ap.has_animation("walk") and ap.has_animation("attack"), "missing clips")
    print("[check-char] PASS"); quit(0)
EOF
godot --headless --path client --import >/dev/null 2>&1 || true
godot --headless --path client --script res://tools/_check_char.gd
rm client/tools/_check_char.gd
```
Expected: `[check-char] PASS`. (This logic becomes permanent in Task 9.)

- [ ] **Step 4.4: Commit**

```bash
git add client/scripts/character_visual.gd client/scenes/character.tscn
git commit -m "feat(client): rigged character scene with bone-attached weapon socket

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Monster dummy scene

**Files:**
- Create: `client/scenes/monster_dummy.tscn`

- [ ] **Step 5.1: Create the scene**

Create `client/scenes/monster_dummy.tscn`:

```
[gd_scene load_steps=3 format=3]

; Slice v3 monster (spec animate-and-react §5.3). Rigged dummy GLB + an
; AnimationPlayer playing idle/hit/death from the committed monster library.
; Instanced at runtime by main.gd::_make_entity_node("monster").
[ext_resource type="PackedScene" path="res://assets/monsters/dummy/monster_dummy.glb" id="1_glb"]
[ext_resource type="AnimationLibrary" path="res://animations/monster_anims.tres" id="2_anims"]

[node name="MonsterDummy" type="Node3D"]

[node name="ModelRoot" parent="." instance=ExtResource("1_glb")]

[node name="AnimationPlayer" type="AnimationPlayer" parent="."]
root_node = NodePath("../ModelRoot")
libraries = {
"": ExtResource("2_anims")
}
```

- [ ] **Step 5.2: Verify headless**

```bash
cat > client/tools/_check_mon.gd <<'EOF'
extends SceneTree
func _initialize():
    var s = (load("res://scenes/monster_dummy.tscn") as PackedScene).instantiate()
    var ap = s.find_child("AnimationPlayer", true, false) as AnimationPlayer
    assert(ap.has_animation("idle") and ap.has_animation("hit") and ap.has_animation("death"), "missing clips")
    print("[check-mon] PASS"); quit(0)
EOF
godot --headless --path client --import >/dev/null 2>&1 || true
godot --headless --path client --script res://tools/_check_mon.gd
rm client/tools/_check_mon.gd
```
Expected: `[check-mon] PASS`.

- [ ] **Step 5.3: Commit**

```bash
git add client/scenes/monster_dummy.tscn
git commit -m "feat(client): rigged monster dummy scene (idle/hit/death)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: AnimationController (TDD)

**Files:**
- Create: `client/scripts/animation_controller.gd`
- Test: covered by `client/tests/test_animation.gd` (created here, extended in Task 9)

- [ ] **Step 6.1: Write the failing controller test**

Create `client/tests/test_animation.gd` (controller section first; rig/scene asserts come in Task 9):

```gdscript
extends SceneTree
# Headless tests for the v3 animation layer (spec §10). Server-independent.
# Run: godot --headless --path client --script res://tests/test_animation.gd
const ControllerScript := preload("res://scripts/animation_controller.gd")


func _initialize() -> void:
    _test_controller_locomotion()
    _test_controller_one_shot_returns()
    _test_controller_terminal_latches()
    _test_controller_hit_ignored_after_death()
    print("[gdtest] PASS: animation controller")
    quit(0)


func _make_player(clips: Array) -> AnimationPlayer:
    var ap := AnimationPlayer.new()
    var lib := AnimationLibrary.new()
    for c in clips:
        var a := Animation.new()
        a.length = (0.5 if c != "idle" and c != "walk" else 1.0)
        lib.add_animation(c, a)
    ap.add_animation_library("", lib)
    get_root().add_child(ap)  # animations need the player in-tree
    return ap


func _test_controller_locomotion() -> void:
    var ap := _make_player(["idle", "walk", "attack"])
    var c = ControllerScript.new(ap)
    c.set_locomotion(false)
    _assert(c.current_clip() == "idle", "idle when not moving, got %s" % c.current_clip())
    c.set_locomotion(true)
    _assert(c.current_clip() == "walk", "walk when moving, got %s" % c.current_clip())
    ap.queue_free()


func _test_controller_one_shot_returns() -> void:
    var ap := _make_player(["idle", "walk", "attack"])
    var c = ControllerScript.new(ap)
    c.set_locomotion(true)
    c.play_one_shot("attack")
    _assert(c.current_clip() == "attack", "attack active, got %s" % c.current_clip())
    # Simulate the clip finishing.
    ap.emit_signal("animation_finished", "attack")
    _assert(c.current_clip() == "walk", "returns to locomotion (walk) after one-shot, got %s" % c.current_clip())
    ap.queue_free()


func _test_controller_terminal_latches() -> void:
    var ap := _make_player(["idle", "hit", "death"])
    var c = ControllerScript.new(ap)
    c.enter_terminal("death")
    _assert(c.current_clip() == "death", "death active, got %s" % c.current_clip())
    c.play_one_shot("hit")        # ignored
    c.set_locomotion(true)        # ignored
    _assert(c.current_clip() == "death", "terminal latched, got %s" % c.current_clip())
    _assert(c.get_debug_state()["terminal"] == true, "terminal flag set")
    ap.queue_free()


func _test_controller_hit_ignored_after_death() -> void:
    var ap := _make_player(["idle", "hit", "death"])
    var c = ControllerScript.new(ap)
    c.enter_terminal("death")
    c.play_one_shot("hit")
    _assert(c.current_clip() == "death", "hit ignored after terminal death, got %s" % c.current_clip())
    ap.queue_free()


func _assert(cond: bool, msg: String) -> void:
    if not cond:
        printerr("[gdtest] FAIL: ", msg)
        quit(1)
```

- [ ] **Step 6.2: Run it to verify it fails**

Run: `godot --headless --path client --script res://tests/test_animation.gd`
Expected: FAIL — `animation_controller.gd` does not exist / class methods undefined.

- [ ] **Step 6.3: Implement the controller**

Create `client/scripts/animation_controller.gd`:

```gdscript
extends RefCounted
class_name AnimationController
# Per-entity animation state machine (spec §4.5). Injected with its
# AnimationPlayer (no absolute scene-path lookups), so main.gd and smoke.gd
# share one code path. It does NOT parse protocol events or know entity types.
#
# State priority (highest wins): terminal (death) > one-shot (attack/hit) >
# locomotion (idle/walk).

var _player: AnimationPlayer
var _moving: bool = false
var _one_shot: String = ""
var _terminal: bool = false
var _terminal_clip: String = ""
var _warnings: Array = []

const IDLE := "idle"
const WALK := "walk"


func _init(player: AnimationPlayer) -> void:
    _player = player
    if _player != null:
        _player.animation_finished.connect(_on_finished)
    _play(IDLE)


func set_locomotion(is_moving: bool) -> void:
    _moving = is_moving
    if _terminal or _one_shot != "":
        return
    _play(WALK if is_moving else IDLE)


func play_one_shot(name: String) -> void:
    if _terminal:
        return
    _one_shot = name
    _play(name)


func enter_terminal(name: String) -> void:
    _terminal = true
    _terminal_clip = name
    _one_shot = ""
    _play(name)


func current_clip() -> String:
    if _player == null:
        return ""
    return str(_player.current_animation)


func get_debug_state() -> Dictionary:
    return {
        "current_clip": current_clip(),
        "terminal": _terminal,
        "is_moving": _moving,
        "warnings": _warnings,
    }


func _on_finished(name: String) -> void:
    if _terminal:
        return
    if name == _one_shot:
        _one_shot = ""
        _play(WALK if _moving else IDLE)


func _play(name: String) -> void:
    if _player == null:
        return
    if not _player.has_animation(name):
        _warn({"code": "unknown_clip", "clip": name})
        return
    _player.play(name)


func _warn(entry: Dictionary) -> void:
    push_warning("[anim] %s" % JSON.stringify(entry))
    _warnings.append(entry)
```

- [ ] **Step 6.4: Run the test to verify it passes**

Run: `godot --headless --path client --script res://tests/test_animation.gd`
Expected: `[gdtest] PASS: animation controller`.

- [ ] **Step 6.5: Commit**

```bash
git add client/scripts/animation_controller.gd client/tests/test_animation.gd
git commit -m "feat(client): AnimationController state machine + tests (TDD)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Resolver simplification

The skeletal `attack` clip now moves the arm; the weapon rides `hand_r`. Delete the fake tween from the resolver.

**Files:**
- Modify: `client/scripts/equipment_visuals.gd`

- [ ] **Step 7.1: Delete the swing fields and constants**

In `equipment_visuals.gd`, remove these member vars:
```gdscript
var _rest_rotation_degrees := Vector3.ZERO
var _swing_tween: Tween = null
```
and remove these constants:
```gdscript
const ATTACK_SWING_DOWN_DEG := 90.0
const ATTACK_SWING_DOWN_SEC := 0.08
const ATTACK_SWING_RETURN_SEC := 0.12
```

- [ ] **Step 7.2: Delete `play_attack_swing()`**

Remove the entire `func play_attack_swing() -> void: ...` function.

- [ ] **Step 7.3: Remove tween handling from `_clear_mounted` and `_refresh_weapon`**

In `_clear_mounted`, delete:
```gdscript
    if _swing_tween != null and _swing_tween.is_valid():
        _swing_tween.kill()
    _swing_tween = null
```
In `_refresh_weapon`, delete the line:
```gdscript
    _rest_rotation_degrees = inst.rotation_degrees
```

- [ ] **Step 7.4: Verify the resolver still loads and the visual test passes**

Run: `godot --headless --path client --script res://tests/test_item_visuals.gd`
Expected: `[gdtest] PASS: item visual resolution ...` (resolver still parses; no reference to removed symbols).

```bash
grep -n "play_attack_swing\|_swing_tween\|_rest_rotation_degrees\|ATTACK_SWING" client/scripts/equipment_visuals.gd || echo "clean"
```
Expected: `clean`.

- [ ] **Step 7.5: Re-tune weapon grip offset if needed (spec §5.5)**

The sword no longer hangs from a static empty node — it rides `hand_r` through the
bone attachment. If the mounted weapon looks misaligned in manual play, adjust
`shared/assets/item_visuals.v0.json` → `rusty_sword.local_transform` only (no
code change). Skip if the default `(0,0,0)` grip still looks acceptable on the
crude rig.

- [ ] **Step 7.6: Commit**

```bash
git add client/scripts/equipment_visuals.gd
git commit -m "refactor(client): drop fake weapon-tilt tween; attack is now skeletal

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Wire `main.gd`

**Files:**
- Modify: `client/scripts/main.gd`

- [ ] **Step 8.1: Add the controller preload, the player controller var, and the event→clip map**

Near the top, after `const EquipmentResolverScript := preload(...)`, add:
```gdscript
const AnimationControllerScript := preload("res://scripts/animation_controller.gd")
const MonsterDummyScene := preload("res://scenes/monster_dummy.tscn")
const MONSTER_EVENT_CLIPS := {
	"monster_damaged": "hit",
	"monster_killed": "death",
}
```
After `var resolver: EquipmentVisualResolver`, add:
```gdscript
var player_anim: AnimationController
```

- [ ] **Step 8.2: Change the `entities` type comment and build the player controller**

Change the declaration comment:
```gdscript
var entities: Dictionary = {}        # id (String) -> {node:Node3D, controller:AnimationController|null, type:String}
```
In `_ready`, after `resolver = EquipmentResolverScript.new(character_visual)`, add:
```gdscript
	var ap := character_visual.find_child("AnimationPlayer", true, false) as AnimationPlayer
	if ap != null:
		player_anim = AnimationControllerScript.new(ap)
```

- [ ] **Step 8.3: Drive locomotion + attack one-shot**

In `_handle_input`, locate the movement block. Right after the `if input != Vector2.ZERO and _send_cooldown <= 0.0:` block, the player either moved or not. Replace the locomotion drive by adding, at the END of `_handle_input` (before its final line), nothing — instead drive locomotion in `_process`. In `_process`, after `_handle_input(delta)`, add:
```gdscript
	if player_anim != null:
		var moving := client.ready_state() == WebSocketPeer.STATE_OPEN \
			and (Input.is_key_pressed(KEY_W) or Input.is_key_pressed(KEY_A) \
			or Input.is_key_pressed(KEY_S) or Input.is_key_pressed(KEY_D))
		player_anim.set_locomotion(moving)
```
In `_try_attack_toward_mouse`, replace:
```gdscript
	if resolver != null:
		resolver.play_attack_swing()
```
with:
```gdscript
	if player_anim != null:
		player_anim.play_one_shot("attack")
```

- [ ] **Step 8.4: Update entity storage + accessors (`{node,controller,type}`)**

Rewrite `_make_entity_node` to return a `Node3D` and instance the monster scene:
```gdscript
func _make_entity_node(kind: String) -> Node3D:
	# Monster adopts the rigged dummy scene (spec §5.3); loot stays a primitive.
	if kind == "monster":
		var packed := MonsterDummyScene
		if packed != null:
			return packed.instantiate()
		# Fallback: red primitive so positioning/targeting still works.
		var fallback := MeshInstance3D.new()
		var fm := StandardMaterial3D.new()
		fm.albedo_color = Color(1.0, 0.3, 0.3)
		fallback.mesh = BoxMesh.new()
		fallback.material_override = fm
		return fallback
	var node := MeshInstance3D.new()  # loot
	var box := BoxMesh.new()
	box.size = Vector3(0.5, 0.5, 0.5)
	node.mesh = box
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(1.0, 0.85, 0.2)
	node.material_override = mat
	return node
```
Rewrite `_upsert_entity`'s non-player branch to store the record + build a controller for monsters. Replace the section from `var node: MeshInstance3D` through the end of the function with:
```gdscript
	var rec: Dictionary
	if entities.has(id):
		rec = entities[id]
	else:
		var node := _make_entity_node(e["type"])
		entities_root.add_child(node)
		var controller: AnimationController = null
		if e["type"] == "monster":
			var ap := node.find_child("AnimationPlayer", true, false) as AnimationPlayer
			if ap != null:
				controller = AnimationControllerScript.new(ap)
			else:
				push_warning("[main] monster %s has no AnimationPlayer" % id)
		rec = {"node": node, "controller": controller, "type": str(e["type"])}
		entities[id] = rec
		if e["type"] == "loot" and not loot_ids.has(id):
			loot_ids.append(id)
		if e["type"] == "monster" and not monster_ids.has(id):
			monster_ids.append(id)
	(rec["node"] as Node3D).position = server_pos
	# Resume/snapshot consistency: a monster already dead in the snapshot enters
	# the terminal death pose without waiting for an event (spec §5.4).
	if rec["type"] == "monster" and rec["controller"] != null:
		var hp = e.get("hp", null)
		if hp != null and int(hp) <= 0:
			rec["controller"].enter_terminal("death")
```
Update `_remove_entity` to read through the record:
```gdscript
func _remove_entity(id: String) -> void:
	if entities.has(id):
		(entities[id]["node"] as Node3D).queue_free()
		entities.erase(id)
	loot_ids.erase(id)
	monster_ids.erase(id)
```
Update `_best_monster_in_direction`: change
```gdscript
		var entity_node: MeshInstance3D = entities[id]
		var to_monster: Vector3 = entity_node.position - predicted_pos
```
to
```gdscript
		var entity_node: Node3D = entities[id]["node"]
		var to_monster: Vector3 = entity_node.position - predicted_pos
```

Update `_apply_snapshot`'s entity-clear loop (currently `entities[id].queue_free()`) to read through the record — same shape as `_remove_entity`:
```gdscript
	for id in entities.keys():
		(entities[id]["node"] as Node3D).queue_free()
	entities.clear()
```
Without this change, the first snapshot after wiring Task 8.4 crashes because `entities[id]` is no longer a `Node3D`.

- [ ] **Step 8.5: Read the `events` array (changes first, then events)**

In `_apply_delta`, the loop currently processes only `changes`. After the existing `for c in p.get("changes", []):` match block completes (i.e., right before the final `_reconcile_player()` call), add:
```gdscript
	for ev in p.get("events", []):
		var clip = MONSTER_EVENT_CLIPS.get(str(ev.get("event_type", "")), null)
		if clip == null:
			continue
		var eid := str(ev.get("entity_id", ""))
		if not entities.has(eid):
			continue
		var ctrl = entities[eid]["controller"]
		if ctrl == null:
			continue
		if clip == "death":
			ctrl.enter_terminal("death")
		else:
			ctrl.play_one_shot(clip)
```

- [ ] **Step 8.6: Fix the debug entity count (optional safety)**

`_update_debug` uses `entities.size()` which is unaffected. No change needed. Verify no remaining `: MeshInstance3D = entities[` casts:
```bash
grep -n "MeshInstance3D = entities\|entities\[id\]\.position\|resolver.play_attack_swing" client/scripts/main.gd || echo "clean"
```
Expected: `clean`.

- [ ] **Step 8.7: Syntax-check by importing**

Run: `godot --headless --path client --import`
Expected: no GDScript parse errors printed for `main.gd`.

- [ ] **Step 8.8: Commit**

```bash
git add client/scripts/main.gd
git commit -m "feat(client): wire animation controllers, locomotion, attack, monster events

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 9: Rig/scene assertions in `test_animation.gd` + smoke wiring

**Files:**
- Modify: `client/tests/test_animation.gd`
- Modify: `scripts/client_smoke.sh`

- [ ] **Step 9.1: Add rig/scene assertions to the test**

In `client/tests/test_animation.gd`, add three calls inside `_initialize()` before the final print:
```gdscript
	_test_character_scene()
	_test_monster_scene()
```
And add these functions:
```gdscript
func _test_character_scene() -> void:
	var s = (load("res://scenes/character.tscn") as PackedScene).instantiate()
	get_root().add_child(s)  # _ready attaches the socket
	var sock = s.find_child("right_hand_socket", true, false)
	_assert(sock is BoneAttachment3D, "right_hand_socket must be a BoneAttachment3D")
	_assert(sock.bone_name == "hand_r", "socket bound to hand_r, got %s" % sock.bone_name)
	var ap := s.find_child("AnimationPlayer", true, false) as AnimationPlayer
	_assert(ap != null, "character AnimationPlayer missing")
	for clip in ["idle", "walk", "attack"]:
		_assert(ap.has_animation(clip), "character missing clip %s" % clip)
	s.queue_free()


func _test_monster_scene() -> void:
	var s = (load("res://scenes/monster_dummy.tscn") as PackedScene).instantiate()
	get_root().add_child(s)
	var ap := s.find_child("AnimationPlayer", true, false) as AnimationPlayer
	_assert(ap != null, "monster AnimationPlayer missing")
	for clip in ["idle", "hit", "death"]:
		_assert(ap.has_animation(clip), "monster missing clip %s" % clip)
	s.queue_free()
```
Update the final print to `print("[gdtest] PASS: animation controller + scenes")`.

- [ ] **Step 9.2: Run the full animation test**

```bash
godot --headless --path client --import >/dev/null 2>&1 || true
godot --headless --path client --script res://tests/test_animation.gd
```
Expected: `[gdtest] PASS: animation controller + scenes`.

- [ ] **Step 9.3: Wire rig gate + animation test into the smoke script**

In `scripts/client_smoke.sh`, after the item-visual test block (step "2."), insert:
```bash
# 2b. Rig gate: both GLBs import as skinned Skeleton3D (spec §10 fail-fast).
echo "[client-smoke] running GDScript rig gate"
"$GODOT" --headless --path "$CLIENT_DIR" --script res://tools/inspect_rig.gd

# 2c. Animation controller + rigged scene test (server-independent).
echo "[client-smoke] running GDScript animation test"
"$GODOT" --headless --path "$CLIENT_DIR" --script res://tests/test_animation.gd
```

- [ ] **Step 9.4: Commit**

```bash
git add client/tests/test_animation.gd scripts/client_smoke.sh
git commit -m "test(client): rig/scene animation assertions; wire into client smoke

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 10: Extend the slice smoke (monster hit/death + resume death pose)

**Files:**
- Modify: `client/scripts/smoke.gd`

The smoke already kills monster `1002` (it sets `killed=true` on the `monster_killed` event). Extend it to (a) instance a real `monster_dummy` + controller and drive it through the event path, and (b) assert the player attack clip and the resume death pose. Keep it deterministic — clip-name/state asserts only.

- [ ] **Step 10.1: Add a monster controller harness to the smoke**

Add near the other preloads:
```gdscript
const AnimControllerScript := preload("res://scripts/animation_controller.gd")
const MonsterScene := preload("res://scenes/monster_dummy.tscn")
```
Add members:
```gdscript
var monster_anim: AnimationController
var monster_saw_hit: bool = false
```
In `_initialize()`, after `resolver = _make_resolver()`, add:
```gdscript
	var mon := MonsterScene.instantiate()
	get_root().add_child(mon)
	var mon_ap := mon.find_child("AnimationPlayer", true, false) as AnimationPlayer
	monster_anim = AnimControllerScript.new(mon_ap)
```

- [ ] **Step 10.2: Drive the monster controller from the same event stream**

In `_handle`, inside the `for ev in p.get("events", []):` loop, replace:
```gdscript
		if ev.get("event_type", "") == "monster_killed":
			killed = true
```
with:
```gdscript
		var et := str(ev.get("event_type", ""))
		if et == "monster_damaged" and monster_anim != null:
			monster_anim.play_one_shot("hit")
			monster_saw_hit = true
		if et == "monster_killed":
			killed = true
			if monster_anim != null:
				monster_anim.enter_terminal("death")
```

- [ ] **Step 10.3: Assert the monster reached the death pose at verify time**

In `_verify_equip()`, after the existing `if server_ok and visual_ok:` success branch prints, before `return true`, add a monster assertion:
```gdscript
	if server_ok and visual_ok:
		if monster_anim != null and monster_anim.get_debug_state()["terminal"] != true:
			_fail("monster did not reach terminal death pose after kill: %s" % monster_anim.get_debug_state())
			return false
		print("[smoke] equip verified + monster death pose terminal (saw_hit=%s)" % monster_saw_hit)
		return true
```
(Remove the old `print("[smoke] equip verified: ...")` line so there is one success print.)

- [ ] **Step 10.4: Assert resume death pose from snapshot (spec §5.4 / acceptance #8)**

The primary smoke already resumes the session for the weapon visual. Extend the
resume phase to assert a dead monster enters terminal `death` from snapshot `hp`
alone — no delta replay.

In `_step_resume`, after `resolver_resume.apply_snapshot(env["payload"])`, also
upsert monsters from the snapshot with the same controller wiring as `main.gd`
(or instantiate one `MonsterScene` + controller and call `enter_terminal("death")`
when any snapshot entity has `type == "monster"` and `hp <= 0`). Assert
`monster_anim.get_debug_state()["terminal"] == true` before the existing weapon
resume check. This covers acceptance #8 for monsters without depending on
`recent_events`.

- [ ] **Step 10.5: Assert the player attack clip during the moving phase**

In `_make_resolver` the smoke uses a bare Node3D mount (no AnimationPlayer), so the smoke has no *player* AnimationController — the player attack-clip assertion lives in `test_animation.gd` (Task 6/9) which already proves `play_one_shot("attack")` → `current_clip()=="attack"` → returns. No change needed here; this step is a no-op confirmation that player attack-clip coverage exists in `test_animation.gd`.

- [ ] **Step 10.6: Run the full smoke against a running server**

```bash
make db-up
make server &
SERVER_PID=$!
sleep 2
make client-smoke
kill $SERVER_PID
```
Expected (if Godot installed): `[smoke] PASS: ...` and `[client-smoke] PASS`. If Godot is absent, `client-smoke` SKIPs with exit 0 — acceptable for CI parity.

- [ ] **Step 10.7: Commit**

```bash
git add client/scripts/smoke.gd
git commit -m "test(client): smoke drives monster hit->death via event path

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 11: ADRs + final verification

**Files:**
- Modify: `docs/adr/0006-asset-pipeline.md`
- Create: `docs/adr/0007-animation-state-model.md`
- Modify: `docs/specs/spec-animate-and-react.md` (As-Built section)

- [ ] **Step 11.1: Append an as-built note to ADR-0006**

Add a short section at the end of `docs/adr/0006-asset-pipeline.md`:
```markdown
## As-built update (slice v3, 2026-06-05)

The predicted upgrade from a placeholder `Node3D` mount socket to a
`BoneAttachment3D` is now executed. `right_hand_socket` keeps its name but is
attached (in `character_visual.gd`) to the rig bone `hand_r`, so the equipped
weapon rides the arm during the skeletal `attack` clip. The asset validator was
upgraded from a node-name check to a **skin-joint** check: a required mount bone
must be an actual glTF skin joint, proving the GLB is genuinely rigged.
```

- [ ] **Step 11.2: Write ADR-0007**

Create `docs/adr/0007-animation-state-model.md`:
```markdown
# ADR-0007: Client animation state model

Status: Accepted (2026-06-05)
Context: ADR-0001 (tech stack), ADR-0006 (asset pipeline), slice v3
`animate-and-react`.

## Decision

Animation is **client-side presentation state**, never authored on the wire.

- The local player's `idle/walk/attack` states are derived from signals already
  present in the client: movement input/prediction and the local attack input.
- The monster's `hit/death` states are driven by the **authoritative
  `monster_damaged` / `monster_killed` events** that the server already emits in
  `state_delta.events`. The client begins reading the `events` array; no new
  message type, schema, or sim change is introduced.
- States are **discrete clips** managed by a small priority state machine
  (`terminal > one-shot > locomotion`) in an injected `AnimationController`. No
  `AnimationTree`/blend spaces in this slice.
- The event→clip mapping is a client-only constant (`main.gd`), deliberately not
  in `shared/`, which is reserved for cross-language server/client contracts.

## Consequences

- Adding entity reactions later (e.g. player damage) requires the server to
  emit the authoritative trigger first; the client mapping then extends trivially.
- Because animation never crosses the wire, server tests and the protocol remain
  untouched (acceptance: empty `server/` diff).
```

- [ ] **Step 11.3: Fill the spec As-Built section**

Replace the placeholder in `docs/specs/spec-animate-and-react.md` §11 with a short summary of what shipped and any deviations (e.g. the socket is attached in `character_visual.gd` rather than declared in the `.tscn`, for robustness against the imported skeleton path; clips are built by `client/tools/build_animations.gd` and committed as `.tres`). Set the Status line to `Implemented (2026-06-05)`.

- [ ] **Step 11.4: Full local gate**

```bash
make validate-shared
make validate-assets
.venv/bin/python -m pytest -q tools
( cd server && go test ./... )
git diff --stat origin/main -- server   # must show NO server changes
make ci
```
Expected: all green; the `server` diff is empty (acceptance #11). `make ci` green including the extended Godot smoke when Godot is installed.

- [ ] **Step 11.5: Commit**

```bash
git add docs/adr/0006-asset-pipeline.md docs/adr/0007-animation-state-model.md docs/specs/spec-animate-and-react.md
git commit -m "docs: ADR-0006 as-built + ADR-0007 animation state model; spec as-built

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Self-Review

**Gaps closed in this revision (would have blocked implementation):**
- Check [5] still required `right_hand_socket` in `required_nodes` while the manifest
  switched to rig joints → added Step 2.4 (mount-bone `hand_r` coverage).
- `_apply_snapshot` still called `entities[id].queue_free()` on raw nodes → added
  Step 8.4 snapshot-clear fix alongside `_remove_entity`.
- Pytest step referenced nonexistent `tmp_repo` / `_ok_labels` fixtures → Step 2.6
  now matches `build_root(tmp_path)` and adds `make_skinned_glb`.
- Hard-coded GLB `uid://` in Task 4.2 would stale after rig regen → omit UID.
- Rig gate (`inspect_rig.gd`) was committed but never wired into CI → Step 9.3.
- Spec §10 resume death pose + hit-after-death controller case were missing →
  Steps 10.4 and 6.1 respectively.
- Spec §5.5 grip re-tune after bone mount → Step 7.5.

**Spec coverage:**
- §3 file map — every Create/Modify path appears in a task. ✓ (`.glb.import` files are committed in Task 1.6 / regenerated by import.)
- §4.1/§4.2 rigs (joints incl. `hand_r`, `pivot`) — Task 1. ✓
- §4.3 manifest (`type:"monster"`, `required_nodes`, sha, mount-bone check [5]) — Task 2. ✓
- §4.4 clips (idle/walk/attack, idle/hit/death) — Task 3. ✓
- §4.5 `AnimationController` (priority, API, `get_debug_state`) — Task 6. ✓
- §4.5 event→clip map in `main.gd` — Task 8.1/8.5. ✓
- §4.6 no protocol change; read `events` — Task 8.5 + acceptance check Task 11.4. ✓
- §5.2 locomotion from input (not reconciliation) — Task 8.3. ✓
- §5.3 monster scene + controller + changes-before-events ordering — Task 8.4/8.5. ✓
- §5.4 resume `hp==0` death pose — Task 8.4 + smoke Step 10.4. ✓
- §5.5 `BoneAttachment3D` + resolver simplification — Tasks 4 + 7. ✓
- §5.6 `_make_entity_node`→`Node3D`, accessors via `["node"]` — Task 8.4. ✓
- §6 skin data required (non-empty skins, JOINTS_0/WEIGHTS_0) — Task 1 + validator Task 2.5. ✓
- §7 failure behavior (unknown clip warn; monster scene load fallback; missing AnimationPlayer warn) — Task 6.3 (`_warn`), 8.4 (fallback). ✓
- §8 acceptance — covered across Tasks 2/4/5/6/9/10/11. ✓
- §10 testing plan (rig gate in smoke, pytest skin GLBs, test_animation.gd hit-after-death, smoke resume death) — Tasks 1.5/2/6/9/10. ✓

**Placeholder scan:** No "TBD"/"handle edge cases"; the only literal placeholder is the all-zero sha in Task 2.2, explicitly replaced in Task 2.7. ✓

**Type consistency:** `AnimationController` API (`set_locomotion`, `play_one_shot`, `enter_terminal`, `current_clip`, `get_debug_state`) is identical across Tasks 6, 8, 9, 10. `entities[id]` is `{node, controller, type}` consistently in Task 8.4/8.5. Event field is `event_type` and entity field `entity_id` (matching the Go `Event` struct and existing `smoke.gd`) in Tasks 8.5 + 10.2. ✓

**Known residual risk:** the skinned-glTF generator (Task 1) is the single highest-risk piece; the rig gate (1.5) fails fast before any animation work depends on it.
