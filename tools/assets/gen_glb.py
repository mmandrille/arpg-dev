#!/usr/bin/env python3
"""Deterministic glTF-binary (.glb) generator — the v2 asset source-of-truth.

ADR-0006 decision #5 fallback (chosen for v2): instead of fetching a CC0 model,
emit low-poly primitive characters/weapons from stdlib only (``struct`` + ``json``,
no extra deps). The proof of this slice is the manifest -> import -> mount
contract, which is identical regardless of how the bytes were authored; a
generator gives **byte-deterministic** output, hence a stable ``sha256`` for
manifest provenance and reproducible CI.

Geometry is built from unit cubes (24 verts with per-face normals) placed via
node TRS, plus empty ``Node3D`` nodes for mount sockets. Materials are embedded
PBR ``baseColorFactor`` (no textures -> no network fetch at import/runtime).

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


def base_humanoid_glb() -> bytes:
    """Low-poly blue-grey humanoid (~1.9 m tall) with a right_hand_socket empty."""
    color = (0.55, 0.62, 0.72, 1.0)
    parts = [
        {"name": "torso", "translation": [0.0, 1.15, 0.0], "scale": [0.5, 0.8, 0.3]},
        {"name": "head", "translation": [0.0, 1.78, 0.0], "scale": [0.34, 0.34, 0.34]},
        {"name": "arm_left", "translation": [-0.42, 1.15, 0.0], "scale": [0.16, 0.72, 0.16]},
        {"name": "arm_right", "translation": [0.42, 1.15, 0.0], "scale": [0.16, 0.72, 0.16]},
        {"name": "leg_left", "translation": [-0.16, 0.45, 0.0], "scale": [0.2, 0.9, 0.2]},
        {"name": "leg_right", "translation": [0.16, 0.45, 0.0], "scale": [0.2, 0.9, 0.2]},
    ]
    # Socket at the bottom of the right arm (the hand), slightly forward.
    empties = [{"name": "right_hand_socket", "translation": [0.42, 0.82, 0.12]}]
    return _build_glb(color, parts, empties)


def rusty_sword_glb() -> bytes:
    """Low-poly rusty one-handed sword, grip at origin, blade pointing +Y."""
    color = (0.45, 0.3, 0.18, 1.0)  # rusty brown
    parts = [
        {"name": "grip", "translation": [0.0, -0.08, 0.0], "scale": [0.05, 0.2, 0.05]},
        {"name": "guard", "translation": [0.0, 0.04, 0.0], "scale": [0.26, 0.05, 0.07]},
        {"name": "blade", "translation": [0.0, 0.5, 0.0], "scale": [0.07, 0.9, 0.02]},
    ]
    return _build_glb(color, parts, [])


TARGETS = {
    "client/assets/characters/base_humanoid/base_humanoid.glb": base_humanoid_glb,
    "client/assets/equipment/weapons/rusty_sword/rusty_sword.glb": rusty_sword_glb,
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
