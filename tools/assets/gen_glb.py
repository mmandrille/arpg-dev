#!/usr/bin/env python3
"""Deterministic glTF-binary (.glb) generator — the v3 asset source-of-truth.

ADR-0006 decision #5 fallback (chosen for v2, extended in v3): instead of
fetching a CC0 model, emit low-poly primitive characters/monsters/weapons from
stdlib only (``struct`` + ``json``, no extra deps). The proof of this slice is
the manifest -> import -> mount/animate contract, which is identical regardless
of how the bytes were authored; a generator gives **byte-deterministic** output,
hence a stable ``sha256`` for manifest provenance and reproducible CI.

Geometry is built from unit cubes (24 verts with per-face normals). v3 emits
**skinned** humanoid + dummy rigs: each cube part is baked into mesh space and
weighted 100% (rigid skinning) to a single translation-only joint, producing a
valid glTF ``skin`` + skeleton that Godot imports as a ``Skeleton3D`` scene so
attack/walk/hit/death clips can drive the joints. The sword stays a static mesh
(``_build_glb`` + node TRS). Materials are embedded PBR ``baseColorFactor`` (no
textures -> no network fetch at import/runtime).

Run via ``make gen-assets`` (or directly) to regenerate the committed runtime
``.glb`` files under ``client/assets/...``. Output is stable across runs/machines.
"""
from __future__ import annotations

import json
import struct
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]

# --- unit cube (centered, edge length 1), 24 verts so each face has flat normals.
_FACES = [
    # (normal, [four corner offsets ccw])
    ((0, 0, 1), [(-0.5, -0.5, 0.5), (0.5, -0.5, 0.5), (0.5, 0.5, 0.5), (-0.5, 0.5, 0.5)]),
    ((0, 0, -1), [(0.5, -0.5, -0.5), (-0.5, -0.5, -0.5), (-0.5, 0.5, -0.5), (0.5, 0.5, -0.5)]),
    ((1, 0, 0), [(0.5, -0.5, 0.5), (0.5, -0.5, -0.5), (0.5, 0.5, -0.5), (0.5, 0.5, 0.5)]),
    ((-1, 0, 0), [(-0.5, -0.5, -0.5), (-0.5, -0.5, 0.5), (-0.5, 0.5, 0.5), (-0.5, 0.5, -0.5)]),
    ((0, 1, 0), [(-0.5, 0.5, 0.5), (0.5, 0.5, 0.5), (0.5, 0.5, -0.5), (-0.5, 0.5, -0.5)]),
    ((0, -1, 0), [(-0.5, -0.5, -0.5), (0.5, -0.5, -0.5), (0.5, -0.5, 0.5), (-0.5, -0.5, 0.5)]),
]


def _cube_geometry() -> tuple[list[tuple[float, float, float]], list[tuple[float, float, float]], list[int]]:
    positions: list[tuple[float, float, float]] = []
    normals: list[tuple[float, float, float]] = []
    indices: list[int] = []
    for normal, corners in _FACES:
        base = len(positions)
        for c in corners:
            positions.append(c)
            normals.append(normal)
        indices += [base, base + 1, base + 2, base, base + 2, base + 3]
    return positions, normals, indices


def _pad(buf: bytearray, alignment: int = 4, fill: int = 0) -> None:
    while len(buf) % alignment != 0:
        buf.append(fill)


def _build_glb(color: tuple[float, float, float, float], parts: list[dict], empties: list[dict]) -> bytes:
    """Build a .glb whose mesh is one shared unit cube, instanced by `parts`.

    parts:   [{"name", "translation":[x,y,z], "scale":[x,y,z]}]  -> cube nodes
    empties: [{"name", "translation":[x,y,z]}]                    -> meshless Node3D
    """
    positions, normals, indices = _cube_geometry()

    bin_buf = bytearray()
    pos_off = len(bin_buf)
    for p in positions:
        bin_buf += struct.pack("<fff", *p)
    nrm_off = len(bin_buf)
    for n in normals:
        bin_buf += struct.pack("<fff", *n)
    idx_off = len(bin_buf)
    for i in indices:
        bin_buf += struct.pack("<H", i)
    _pad(bin_buf)

    pmin = [min(c[i] for c in positions) for i in range(3)]
    pmax = [max(c[i] for c in positions) for i in range(3)]

    nodes: list[dict] = []
    child_indices: list[int] = []
    for part in parts:
        nodes.append({
            "name": part["name"],
            "mesh": 0,
            "translation": part["translation"],
            "scale": part["scale"],
        })
        child_indices.append(len(nodes) - 1)
    for empty in empties:
        nodes.append({"name": empty["name"], "translation": empty["translation"]})
        child_indices.append(len(nodes) - 1)

    gltf = {
        "asset": {"version": "2.0", "generator": "arpg-dev/tools/assets/gen_glb.py"},
        "scene": 0,
        "scenes": [{"nodes": child_indices}],
        "nodes": nodes,
        "meshes": [{
            "primitives": [{
                "attributes": {"POSITION": 0, "NORMAL": 1},
                "indices": 2,
                "material": 0,
                "mode": 4,
            }],
        }],
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
        ],
        "bufferViews": [
            {"buffer": 0, "byteOffset": pos_off, "byteLength": nrm_off - pos_off, "target": 34962},
            {"buffer": 0, "byteOffset": nrm_off, "byteLength": idx_off - nrm_off, "target": 34962},
            {"buffer": 0, "byteOffset": idx_off, "byteLength": len(indices) * 2, "target": 34963},
        ],
        "buffers": [{"byteLength": len(bin_buf)}],
    }

    # Deterministic JSON: sorted keys, compact separators, padded with spaces.
    json_bytes = bytearray(json.dumps(gltf, sort_keys=True, separators=(",", ":")).encode("utf-8"))
    while len(json_bytes) % 4 != 0:
        json_bytes.append(0x20)  # space

    json_chunk = struct.pack("<II", len(json_bytes), 0x4E4F534A) + bytes(json_bytes)  # 'JSON'
    bin_chunk = struct.pack("<II", len(bin_buf), 0x004E4942) + bytes(bin_buf)         # 'BIN\0'
    total = 12 + len(json_chunk) + len(bin_chunk)
    header = b"glTF" + struct.pack("<II", 2, total)
    return header + json_chunk + bin_chunk


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
    for idx, (name, _parent, local) in enumerate(joints):
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


def rusty_sword_glb() -> bytes:
    """Low-poly rusty one-handed sword, grip at origin, blade pointing +Y."""
    color = (0.45, 0.3, 0.18, 1.0)  # rusty brown
    parts = [
        {"name": "grip", "translation": [0.0, -0.08, 0.0], "scale": [0.05, 0.2, 0.05]},
        {"name": "guard", "translation": [0.0, 0.04, 0.0], "scale": [0.26, 0.05, 0.07]},
        {"name": "blade", "translation": [0.0, 0.5, 0.0], "scale": [0.07, 0.9, 0.02]},
    ]
    return _build_glb(color, parts, [])


def training_bow_glb() -> bytes:
    """Low-poly bow, grip at origin, bow/string standing along local Y."""
    color = (0.38, 0.24, 0.12, 1.0)
    parts = [
        {"name": "grip", "translation": [0.0, 0.0, 0.0], "scale": [0.08, 0.24, 0.08]},
        {"name": "upper_limb_inner", "translation": [-0.08, 0.38, 0.0], "scale": [0.05, 0.55, 0.05]},
        {"name": "lower_limb_inner", "translation": [-0.08, -0.38, 0.0], "scale": [0.05, 0.55, 0.05]},
        {"name": "upper_limb_tip", "translation": [-0.19, 0.78, 0.0], "scale": [0.05, 0.36, 0.05]},
        {"name": "lower_limb_tip", "translation": [-0.19, -0.78, 0.0], "scale": [0.05, 0.36, 0.05]},
        {"name": "string", "translation": [0.16, 0.0, 0.0], "scale": [0.018, 1.45, 0.018]},
    ]
    return _build_glb(color, parts, [])


TARGETS = {
    "client/assets/characters/base_humanoid/base_humanoid.glb": base_humanoid_glb,
    "client/assets/equipment/weapons/rusty_sword/rusty_sword.glb": rusty_sword_glb,
    "client/assets/equipment/weapons/training_bow/training_bow.glb": training_bow_glb,
    "client/assets/monsters/dummy/monster_dummy.glb": monster_dummy_glb,
}


def main() -> int:
    for rel, fn in TARGETS.items():
        out = ROOT / rel
        out.parent.mkdir(parents=True, exist_ok=True)
        data = fn()
        out.write_bytes(data)
        print(f"wrote {rel} ({len(data)} bytes)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
