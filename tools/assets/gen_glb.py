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
import math
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


def _prism_geom(n: int, r_bot: float, r_top: float, h: float):
    """N-sided frustum centered on Y axis (y: -h/2 to +h/2), flat-shaded faces."""
    pos, nrm, idx = [], [], []
    hh = h / 2.0
    # Side faces: 4n vertices (4 per quad face)
    for i in range(n):
        a0 = 2 * math.pi * i / n
        a1 = 2 * math.pi * (i + 1) / n
        am = (a0 + a1) / 2.0
        bl = (r_bot * math.cos(a0), -hh, r_bot * math.sin(a0))
        br = (r_bot * math.cos(a1), -hh, r_bot * math.sin(a1))
        tr = (r_top * math.cos(a1),  hh, r_top * math.sin(a1))
        tl = (r_top * math.cos(a0),  hh, r_top * math.sin(a0))
        dr = (r_top - r_bot) / h if h else 0.0
        raw = (math.cos(am), -dr, math.sin(am))
        nl = math.sqrt(sum(v * v for v in raw))
        fn = tuple(v / nl for v in raw)
        b = len(pos)
        for v in (bl, tl, tr, br):  # CCW from outside → outward normal
            pos.append(v)
            nrm.append(fn)
        idx += [b, b + 1, b + 2, b, b + 2, b + 3]
    # Bottom cap: 1 center + n edge vertices
    cy = -hh
    cn = (0.0, -1.0, 0.0)
    c_bot = len(pos)
    pos.append((0.0, cy, 0.0))
    nrm.append(cn)
    bot_edges = []
    for i in range(n):
        a = 2 * math.pi * i / n
        bot_edges.append(len(pos))
        pos.append((r_bot * math.cos(a), cy, r_bot * math.sin(a)))
        nrm.append(cn)
    for i in range(n):
        p0 = bot_edges[i]
        p1 = bot_edges[(i + 1) % n]
        idx += [c_bot, p1, p0]  # CW from below = CCW viewed from -y
    # Top cap: 1 center + n edge vertices
    cy = hh
    cn = (0.0, 1.0, 0.0)
    c_top = len(pos)
    pos.append((0.0, cy, 0.0))
    nrm.append(cn)
    top_edges = []
    for i in range(n):
        a = 2 * math.pi * i / n
        top_edges.append(len(pos))
        pos.append((r_top * math.cos(a), cy, r_top * math.sin(a)))
        nrm.append(cn)
    for i in range(n):
        p0 = top_edges[i]
        p1 = top_edges[(i + 1) % n]
        idx += [c_top, p0, p1]  # CCW viewed from +y
    return pos, nrm, idx


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

    positions, normals, colors0, indices, joints0, weights0 = [], [], [], [], [], []
    for part in parts:
        if len(part) == 5:
            joint_idx, (tx, ty, tz), _, part_color, (geom_pos, geom_nrm, geom_idx) = part
            if part_color is None:
                part_color = color
        elif len(part) == 4:
            joint_idx, (tx, ty, tz), (sx, sy, sz), part_color = part
            geom_pos = [(px * sx, py * sy, pz * sz) for px, py, pz in cube_pos]
            geom_nrm, geom_idx = cube_nrm, cube_idx
        else:
            joint_idx, (tx, ty, tz), (sx, sy, sz) = part
            part_color = color
            geom_pos = [(px * sx, py * sy, pz * sz) for px, py, pz in cube_pos]
            geom_nrm, geom_idx = cube_nrm, cube_idx
        base = len(positions)
        for (px, py, pz), n in zip(geom_pos, geom_nrm):
            positions.append((px + tx, py + ty, pz + tz))
            normals.append(n)
            colors0.append(part_color)
            joints0.append((joint_idx, 0, 0, 0))
            weights0.append((1.0, 0.0, 0.0, 0.0))
        for i in geom_idx:
            indices.append(base + i)

    bin_buf = bytearray()
    pos_off = len(bin_buf)
    for p in positions:
        bin_buf += struct.pack("<fff", *p)
    nrm_off = len(bin_buf)
    for n in normals:
        bin_buf += struct.pack("<fff", *n)
    color_off = len(bin_buf)
    for c in colors0:
        bin_buf += struct.pack("<ffff", *c)
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
                "attributes": {"POSITION": 0, "NORMAL": 1, "JOINTS_0": 3, "WEIGHTS_0": 4, "COLOR_0": 6},
                "indices": 2,
                "material": 0,
                "mode": 4,
            }],
        }],
        "skins": [{"joints": list(range(len(joints))), "inverseBindMatrices": 5}],
        "materials": [{
            "pbrMetallicRoughness": {
                "baseColorFactor": [1.0, 1.0, 1.0, 1.0],
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
            {"bufferView": 6, "componentType": 5126, "count": len(colors0), "type": "VEC4"},
        ],
        "bufferViews": [
            {"buffer": 0, "byteOffset": pos_off, "byteLength": nrm_off - pos_off, "target": 34962},
            {"buffer": 0, "byteOffset": nrm_off, "byteLength": color_off - nrm_off, "target": 34962},
            {"buffer": 0, "byteOffset": idx_off, "byteLength": ibm_off - idx_off, "target": 34963},
            {"buffer": 0, "byteOffset": j_off, "byteLength": w_off - j_off, "target": 34962},
            {"buffer": 0, "byteOffset": w_off, "byteLength": idx_off - w_off, "target": 34962},
            {"buffer": 0, "byteOffset": ibm_off, "byteLength": len(joints) * 64},
            {"buffer": 0, "byteOffset": color_off, "byteLength": j_off - color_off, "target": 34962},
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


def _humanoid_glb(color, parts) -> bytes:
    """Low-poly humanoid as a SKINNED rig.

    Joints (translation-only bind): arm_r/arm_l pivot at the shoulders so attack
    clips swing the correct arm; hand_r/hand_l are children of those arm joints
    so BoneAttachment3D sockets ride the swing. leg_l/leg_r drive the walk clip.
    """
    joints = [
        ("root", -1, (0.0, 0.0, 0.0)),     # 0
        ("spine", 0, (0.0, 1.15, 0.0)),    # 1  global (0,1.15,0)
        ("arm_l", 1, (-0.42, 0.35, 0.0)),  # 2  global (-0.42,1.5,0) = shoulder
        ("hand_l", 2, (0.0, -0.68, 0.12)), # 3  global (-0.42,0.82,0.12) = hand
        ("arm_r", 1, (0.42, 0.35, 0.0)),   # 4  global (0.42,1.5,0) = shoulder
        ("hand_r", 4, (0.0, -0.68, 0.12)), # 5  global (0.42,0.82,0.12) = hand
        ("leg_l", 0, (-0.16, 0.9, 0.0)),   # 6
        ("leg_r", 0, (0.16, 0.9, 0.0)),    # 7
    ]
    return _build_skinned_glb(color, joints, parts)


def base_humanoid_glb() -> bytes:
    """Low-poly blue-grey humanoid (~1.9 m) as the generic fallback."""
    return _humanoid_glb((0.55, 0.62, 0.72, 1.0), [
        (1, (0.0, 1.15, 0.0), (0.5, 0.8, 0.3)),     # torso -> spine
        (1, (0.0, 1.78, 0.0), (0.34, 0.34, 0.34)),  # head  -> spine
        (2, (-0.42, 1.15, 0.0), (0.16, 0.72, 0.16)),# left arm -> arm_l
        (4, (0.42, 1.15, 0.0), (0.16, 0.72, 0.16)), # right arm -> arm_r
        (6, (-0.16, 0.45, 0.0), (0.2, 0.9, 0.2)),   # left leg -> leg_l
        (7, (0.16, 0.45, 0.0), (0.2, 0.9, 0.2)),    # right leg -> leg_r
    ])


def _full_humanoid_glb(color, parts=None, extra_parts=None) -> bytes:
    """17-bone humanoid — rest pose matches bind pose so bones sit inside mesh.

    Bone local translations match the re-rigged barbarian.glb (arm chain goes
    from shoulder downward-inward; foot bones extend past the ankle).  Each
    mesh part is centered at its controlling bone's world position so the
    editor shows bones inside the mesh without any Blender work.

    Bone index map (used by extra_parts):
      0=root 1=spine 2=chest 3=neck 4=head
      5=arm_l  6=elbow_l  7=hand_l
      8=arm_r  9=elbow_r 10=hand_r
     11=leg_l 12=knee_l 13=foot_l
     14=leg_r 15=knee_r 16=foot_r
    """
    joints = [
        # name        parent   local translation (matches re-rigged GLB)
        ("root",       -1, ( 0.0,     0.0,     0.1588)),   # 0  world (0, 0, 0.159)
        ("spine",        0, ( 0.0,     1.1327,  0.0   )),   # 1  world (0, 1.133, 0.159)
        ("chest",        1, ( 0.0,     0.2068,  0.0   )),   # 2  world (0, 1.340, 0.159)
        ("neck",         2, ( 0.0,     0.3349,  0.0   )),   # 3  world (0, 1.674, 0.159)
        ("head",         3, ( 0.0,     0.1379,  0.0   )),   # 4  world (0, 1.812, 0.159)
        ("arm_l",        2, (-0.3506,  0.1009, -0.129 )),   # 5  world (-0.351, 1.440, 0.030)
        ("elbow_l",      5, ( 0.07,   -0.3099,  0.0099)),   # 6  world (-0.281, 1.131, 0.040)
        ("hand_l",       6, ( 0.058,  -0.2539,  0.0079)),   # 7  world (-0.223, 0.877, 0.048)
        ("arm_r",        2, ( 0.3506,  0.1009, -0.129 )),   # 8  world ( 0.351, 1.440, 0.030)
        ("elbow_r",      8, (-0.07,   -0.3099,  0.0099)),   # 9  world ( 0.281, 1.131, 0.040)
        ("hand_r",       9, (-0.058,  -0.2539,  0.0079)),   # 10 world ( 0.223, 0.877, 0.048)
        ("leg_l",        0, (-0.1551,  0.8865,  0.0   )),   # 11 world (-0.155, 0.887, 0.159)
        ("knee_l",      11, ( 0.0,    -0.394,   0.0   )),   # 12 world (-0.155, 0.493, 0.159)
        ("foot_l",      12, (-0.077,  -0.454,  -0.054 )),   # 13 world (-0.232, 0.039, 0.105)
        ("leg_r",        0, ( 0.1551,  0.8865,  0.0   )),   # 14 world ( 0.155, 0.887, 0.159)
        ("knee_r",      14, ( 0.0,    -0.394,   0.0   )),   # 15 world ( 0.155, 0.493, 0.159)
        ("foot_r",      15, ( 0.077,  -0.454,  -0.054 )),   # 16 world ( 0.232, 0.039, 0.105)
    ]
    _default_parts = [
        # head
        ( 4, ( 0.000,  1.762,  0.159), (0.26, 0.28, 0.26)),
        # torso: upper (chest) + lower (spine)
        ( 2, ( 0.000,  1.507,  0.159), (0.46, 0.33, 0.22)),
        ( 1, ( 0.000,  1.236,  0.159), (0.38, 0.32, 0.18)),
        # hips
        ( 1, ( 0.000,  0.980,  0.159), (0.34, 0.18, 0.17)),
        # right arm: upper (arm_r), forearm (elbow_r), hand (hand_r)
        ( 8, ( 0.316,  1.285,  0.035), (0.09, 0.32, 0.09)),
        ( 9, ( 0.252,  1.004,  0.044), (0.08, 0.26, 0.08)),
        (10, ( 0.223,  0.877,  0.048), (0.09, 0.10, 0.07)),
        # left arm (mirrored X)
        ( 5, (-0.316,  1.285,  0.035), (0.09, 0.32, 0.09)),
        ( 6, (-0.252,  1.004,  0.044), (0.08, 0.26, 0.08)),
        ( 7, (-0.223,  0.877,  0.048), (0.09, 0.10, 0.07)),
        # right leg: upper (leg_r), lower (knee_r), foot (foot_r)
        (14, ( 0.155,  0.690,  0.159), (0.13, 0.40, 0.13)),
        (15, ( 0.194,  0.266,  0.132), (0.10, 0.46, 0.10)),
        (16, ( 0.232,  0.039,  0.105), (0.13, 0.07, 0.20)),
        # left leg (mirrored X)
        (11, (-0.155,  0.690,  0.159), (0.13, 0.40, 0.13)),
        (12, (-0.194,  0.266,  0.132), (0.10, 0.46, 0.10)),
        (13, (-0.232,  0.039,  0.105), (0.13, 0.07, 0.20)),
    ]
    if parts is None:
        parts = _default_parts
    if extra_parts:
        parts = list(parts) + list(extra_parts)
    return _build_skinned_glb(color, joints, parts)


def barbarian_glb() -> bytes:
    """17-bone low-poly humanoid — frustum segments weighted to each bone."""
    skin = (0.66, 0.36, 0.25, 1.0)
    parts = [
        # head (bone 4)
        ( 4, ( 0.000,  1.762,  0.159), None, skin, _prism_geom(8, 0.13, 0.13, 0.28)),
        # upper torso (bone 2): wide at shoulders, narrow at waist
        ( 2, ( 0.000,  1.507,  0.159), None, skin, _prism_geom(8, 0.19, 0.23, 0.33)),
        # lower torso (bone 1)
        ( 1, ( 0.000,  1.236,  0.159), None, skin, _prism_geom(8, 0.17, 0.19, 0.32)),
        # hips (bone 1)
        ( 1, ( 0.000,  0.980,  0.159), None, skin, _prism_geom(8, 0.15, 0.17, 0.18)),
        # right arm: upper (bone 8), forearm (bone 9), hand box (bone 10)
        ( 8, ( 0.316,  1.285,  0.035), None, skin, _prism_geom(6, 0.045, 0.055, 0.32)),
        ( 9, ( 0.252,  1.004,  0.044), None, skin, _prism_geom(6, 0.036, 0.045, 0.26)),
        (10, ( 0.223,  0.877,  0.048), (0.09, 0.10, 0.07)),
        # left arm: upper (bone 5), forearm (bone 6), hand box (bone 7)
        ( 5, (-0.316,  1.285,  0.035), None, skin, _prism_geom(6, 0.045, 0.055, 0.32)),
        ( 6, (-0.252,  1.004,  0.044), None, skin, _prism_geom(6, 0.036, 0.045, 0.26)),
        ( 7, (-0.223,  0.877,  0.048), (0.09, 0.10, 0.07)),
        # right leg: thigh (bone 14), shin (bone 15), foot box (bone 16)
        (14, ( 0.155,  0.690,  0.159), None, skin, _prism_geom(6, 0.062, 0.075, 0.40)),
        (15, ( 0.194,  0.266,  0.132), None, skin, _prism_geom(6, 0.048, 0.062, 0.46)),
        (16, ( 0.232,  0.039,  0.105), (0.13, 0.07, 0.20)),
        # left leg: thigh (bone 11), shin (bone 12), foot box (bone 13)
        (11, (-0.155,  0.690,  0.159), None, skin, _prism_geom(6, 0.062, 0.075, 0.40)),
        (12, (-0.194,  0.266,  0.132), None, skin, _prism_geom(6, 0.048, 0.062, 0.46)),
        (13, (-0.232,  0.039,  0.105), (0.13, 0.07, 0.20)),
    ]
    return _full_humanoid_glb(skin, parts=parts)


def sorcerer_glb() -> bytes:
    """Robe silhouette with a pointed hat."""
    return _humanoid_glb((0.26, 0.30, 0.70, 1.0), [
        (1, (0.0, 1.06, 0.0), (0.54, 0.96, 0.36)),   # robe body
        (1, (0.0, 0.55, 0.0), (0.68, 0.42, 0.40)),   # robe hem
        (1, (0.0, 1.78, 0.0), (0.30, 0.30, 0.30)),   # head
        (1, (0.0, 2.08, 0.0), (0.34, 0.34, 0.34)),   # hat brim
        (1, (0.0, 2.32, 0.0), (0.18, 0.50, 0.18)),   # pointed hat
        (2, (-0.38, 1.14, 0.0), (0.13, 0.70, 0.14)),
        (4, (0.38, 1.14, 0.0), (0.13, 0.70, 0.14)),
        (6, (-0.14, 0.38, 0.0), (0.15, 0.62, 0.15)),
        (7, (0.14, 0.38, 0.0), (0.15, 0.62, 0.15)),
    ])


def paladin_glb() -> bytes:
    """Armored but thinner than the barbarian, with a chest cross."""
    return _humanoid_glb((0.62, 0.58, 0.48, 1.0), [
        (1, (0.0, 1.15, 0.0), (0.52, 0.82, 0.32)),
        (1, (0.0, 1.78, 0.0), (0.32, 0.32, 0.32)),
        (1, (0.0, 1.20, 0.21), (0.08, 0.62, 0.06), (0.96, 0.91, 0.50, 1.0)), # front cross vertical
        (1, (0.0, 1.34, 0.24), (0.38, 0.08, 0.06), (0.96, 0.91, 0.50, 1.0)), # front cross horizontal
        (1, (0.0, 1.20, -0.21), (0.08, 0.62, 0.06), (0.96, 0.91, 0.50, 1.0)), # back cross vertical
        (1, (0.0, 1.34, -0.24), (0.38, 0.08, 0.06), (0.96, 0.91, 0.50, 1.0)), # back cross horizontal
        (1, (0.31, 1.20, 0.0), (0.06, 0.62, 0.08), (0.96, 0.91, 0.50, 1.0)), # side cross vertical
        (1, (0.34, 1.34, 0.0), (0.06, 0.08, 0.38), (0.96, 0.91, 0.50, 1.0)), # side cross horizontal
        (1, (-0.31, 1.20, 0.0), (0.06, 0.62, 0.08), (0.96, 0.91, 0.50, 1.0)), # opposite side cross vertical
        (1, (-0.34, 1.34, 0.0), (0.06, 0.08, 0.38), (0.96, 0.91, 0.50, 1.0)), # opposite side cross horizontal
        (2, (-0.43, 1.15, 0.0), (0.17, 0.74, 0.17)),
        (4, (0.43, 1.15, 0.0), (0.17, 0.74, 0.17)),
        (6, (-0.17, 0.45, 0.0), (0.19, 0.88, 0.19)),
        (7, (0.17, 0.45, 0.0), (0.19, 0.88, 0.19)),
    ])


def rogue_glb() -> bytes:
    """Smaller, thinner dual-wield class silhouette."""
    return _humanoid_glb((0.24, 0.48, 0.36, 1.0), [
        (1, (0.0, 1.05, 0.0), (0.42, 0.76, 0.26)),
        (1, (0.0, 1.63, 0.0), (0.28, 0.28, 0.28)),
        (1, (0.0, 1.05, 0.18), (0.34, 0.12, 0.05), (0.08, 0.12, 0.10, 1.0)),
        (2, (-0.34, 1.06, 0.0), (0.12, 0.64, 0.12)),
        (4, (0.34, 1.06, 0.0), (0.12, 0.64, 0.12)),
        (6, (-0.13, 0.40, 0.0), (0.15, 0.80, 0.15)),
        (7, (0.13, 0.40, 0.0), (0.15, 0.80, 0.15)),
    ])


def ranger_glb() -> bytes:
    """Tall, thin hooded bow class silhouette."""
    return _humanoid_glb((0.20, 0.50, 0.27, 1.0), [
        (1, (0.0, 1.16, 0.0), (0.40, 0.92, 0.25)),
        (1, (0.0, 1.86, 0.0), (0.27, 0.30, 0.27)),
        (1, (0.0, 1.94, 0.0), (0.42, 0.34, 0.36), (0.07, 0.15, 0.10, 1.0)),
        (1, (0.0, 1.20, -0.18), (0.36, 0.78, 0.06), (0.09, 0.22, 0.13, 1.0)),
        (1, (0.0, 1.28, 0.18), (0.30, 0.08, 0.05), (0.74, 0.66, 0.34, 1.0)),
        (2, (-0.36, 1.18, 0.0), (0.12, 0.78, 0.12)),
        (4, (0.36, 1.18, 0.0), (0.12, 0.78, 0.12)),
        (6, (-0.13, 0.46, 0.0), (0.14, 0.92, 0.14)),
        (7, (0.13, 0.46, 0.0), (0.14, 0.92, 0.14)),
    ])


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


def monster_tiny_flyer_glb() -> bytes:
    """Tiny bat-like flyer with wing joints for client-side flap clips."""
    color = (0.20, 0.18, 0.28, 1.0)
    joints = [
        ("root", -1, (0.0, 0.0, 0.0)),
        ("body", 0, (0.0, 0.36, 0.0)),
        ("head", 1, (0.0, 0.10, -0.24)),
        ("wing_l", 1, (-0.28, 0.03, 0.0)),
        ("wing_r", 1, (0.28, 0.03, 0.0)),
    ]
    parts = [
        (1, (0.0, 0.38, 0.0), (0.30, 0.26, 0.24)),
        (2, (0.0, 0.46, -0.24), (0.22, 0.18, 0.18)),
        (2, (-0.08, 0.58, -0.30), (0.06, 0.10, 0.05)),
        (2, (0.08, 0.58, -0.30), (0.06, 0.10, 0.05)),
        (3, (-0.48, 0.40, 0.02), (0.55, 0.05, 0.34)),
        (4, (0.48, 0.40, 0.02), (0.55, 0.05, 0.34)),
    ]
    return _build_skinned_glb(color, joints, parts)


def monster_skeleton_glb() -> bytes:
    """Low-poly skeleton silhouette with biped joints for existing hit/death clips."""
    color = (0.82, 0.78, 0.66, 1.0)
    joints = [
        ("root", -1, (0.0, 0.0, 0.0)),
        ("spine", 0, (0.0, 0.92, 0.0)),
        ("head", 1, (0.0, 0.62, 0.0)),
        ("arm_l", 1, (-0.34, 0.38, 0.0)),
        ("arm_r", 1, (0.34, 0.38, 0.0)),
        ("leg_l", 0, (-0.13, 0.42, 0.0)),
        ("leg_r", 0, (0.13, 0.42, 0.0)),
    ]
    parts = [
        (1, (0.0, 0.92, 0.0), (0.22, 0.74, 0.16)),
        (1, (-0.18, 1.05, 0.0), (0.05, 0.34, 0.06)),
        (1, (0.18, 1.05, 0.0), (0.05, 0.34, 0.06)),
        (2, (0.0, 1.55, -0.02), (0.30, 0.30, 0.24)),
        (2, (0.0, 1.55, -0.18), (0.12, 0.10, 0.08)),
        (3, (-0.44, 0.95, 0.0), (0.10, 0.58, 0.10)),
        (3, (-0.46, 0.47, 0.0), (0.09, 0.50, 0.09)),
        (4, (0.44, 0.95, 0.0), (0.10, 0.58, 0.10)),
        (4, (0.46, 0.47, 0.0), (0.09, 0.50, 0.09)),
        (5, (-0.13, 0.45, 0.0), (0.11, 0.78, 0.11)),
        (6, (0.13, 0.45, 0.0), (0.11, 0.78, 0.11)),
    ]
    return _build_skinned_glb(color, joints, parts)


def _all_targets() -> dict:
    from tools.assets import gen_glb_equipment

    return {
        # base_humanoid.glb removed — character_base_humanoid_v0 now points to barbarian.glb.
        "client/assets/characters/barbarian/barbarian.glb": barbarian_glb,
        "client/assets/characters/sorcerer/sorcerer.glb": sorcerer_glb,
        "client/assets/characters/paladin/paladin.glb": paladin_glb,
        "client/assets/characters/rogue/rogue.glb": rogue_glb,
        "client/assets/characters/ranger/ranger.glb": ranger_glb,
        **gen_glb_equipment.EQUIPMENT_TARGETS,
        "client/assets/monsters/dummy/monster_dummy.glb": monster_dummy_glb,
        "client/assets/monsters/skeleton/monster_skeleton.glb": monster_skeleton_glb,
    }


TARGETS = _all_targets()


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
