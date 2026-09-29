#!/usr/bin/env python3
"""Stdlib-only structural reader for glTF 2.0 files (``.glb`` and ``.gltf``).

Shared by ``validate_assets.py`` (manifest checks) and ``inspect_kit.py``
(ADR-0018 P0 kit reports). It reads structure — node tree, meshes, skins,
animations, image sizes, accessor bounds — and never decodes vertex data, so it
stays fast on multi-megabyte kit files and needs no third-party packages.
"""
from __future__ import annotations

import base64
import json
import math
import struct
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

GLB_MAGIC = b"glTF"
CHUNK_JSON = 0x4E4F534A
CHUNK_BIN = 0x004E4942
MODE_TRIANGLES = 4

Matrix = list[float]  # column-major 4x4, glTF convention


@dataclass
class Gltf:
    """Parsed glTF document plus resolved buffer bytes (None when unresolvable)."""

    path: Path
    doc: dict[str, Any]
    buffers: list[bytes | None] = field(default_factory=list)

    def list(self, key: str) -> list[dict[str, Any]]:
        value = self.doc.get(key, [])
        return value if isinstance(value, list) else []

    def node_name(self, index: int) -> str:
        nodes = self.list("nodes")
        if 0 <= index < len(nodes):
            return str(nodes[index].get("name", f"node_{index}"))
        return f"node_{index}"


def _parse_glb_bytes(data: bytes) -> tuple[dict[str, Any], bytes | None]:
    if len(data) < 20 or data[0:4] != GLB_MAGIC:
        raise ValueError("not a GLB file")
    doc: dict[str, Any] | None = None
    bin_chunk: bytes | None = None
    offset = 12
    while offset + 8 <= len(data):
        chunk_len, chunk_type = struct.unpack_from("<II", data, offset)
        offset += 8
        chunk = data[offset : offset + chunk_len]
        offset += chunk_len
        if chunk_type == CHUNK_JSON and doc is None:
            doc = json.loads(chunk.decode("utf-8"))
        elif chunk_type == CHUNK_BIN and bin_chunk is None:
            bin_chunk = bytes(chunk)
    if doc is None:
        raise ValueError("GLB has no JSON chunk")
    return doc, bin_chunk


def _resolve_uri(base_dir: Path, uri: str) -> bytes | None:
    if uri.startswith("data:"):
        _, _, payload = uri.partition(",")
        try:
            return base64.b64decode(payload)
        except ValueError:
            return None
    target = base_dir / uri
    return target.read_bytes() if target.is_file() else None


def load_gltf(path: Path) -> Gltf:
    """Load a ``.glb`` or ``.gltf`` file. Raises ValueError on malformed input."""
    path = Path(path)
    raw = path.read_bytes()
    bin_chunk: bytes | None = None
    if raw[0:4] == GLB_MAGIC:
        doc, bin_chunk = _parse_glb_bytes(raw)
    else:
        try:
            doc = json.loads(raw.decode("utf-8"))
        except (UnicodeDecodeError, json.JSONDecodeError) as exc:
            raise ValueError(f"{path.name} is neither GLB nor glTF JSON") from exc
    if not isinstance(doc, dict):
        raise ValueError(f"{path.name} glTF root is not an object")
    buffers: list[bytes | None] = []
    for index, buffer in enumerate(doc.get("buffers", []) or []):
        uri = buffer.get("uri") if isinstance(buffer, dict) else None
        if uri:
            buffers.append(_resolve_uri(path.parent, str(uri)))
        else:
            buffers.append(bin_chunk if index == 0 else None)
    return Gltf(path=path, doc=doc, buffers=buffers)


# --- accessors / meshes -------------------------------------------------------


def accessor_count(gltf: Gltf, index: int) -> int:
    accessors = gltf.list("accessors")
    if 0 <= index < len(accessors):
        return int(accessors[index].get("count", 0))
    return 0


def accessor_min_max(gltf: Gltf, index: int) -> tuple[list[float], list[float]] | None:
    accessors = gltf.list("accessors")
    if not 0 <= index < len(accessors):
        return None
    accessor = accessors[index]
    lo, hi = accessor.get("min"), accessor.get("max")
    if not isinstance(lo, list) or not isinstance(hi, list):
        return None
    return [float(v) for v in lo], [float(v) for v in hi]


def primitive_triangle_count(gltf: Gltf, primitive: dict[str, Any]) -> int | None:
    """Triangles for a mode-4 primitive; None for non-triangle-list modes."""
    if int(primitive.get("mode", MODE_TRIANGLES)) != MODE_TRIANGLES:
        return None
    if "indices" in primitive:
        return accessor_count(gltf, int(primitive["indices"])) // 3
    position = primitive.get("attributes", {}).get("POSITION")
    return accessor_count(gltf, int(position)) // 3 if position is not None else 0


def mesh_triangle_count(gltf: Gltf, mesh_index: int) -> int:
    meshes = gltf.list("meshes")
    if not 0 <= mesh_index < len(meshes):
        return 0
    total = 0
    for primitive in meshes[mesh_index].get("primitives", []):
        count = primitive_triangle_count(gltf, primitive)
        total += count or 0
    return total


# --- node tree / transforms ---------------------------------------------------


def root_nodes(gltf: Gltf) -> list[int]:
    scenes = gltf.list("scenes")
    scene_index = int(gltf.doc.get("scene", 0))
    if 0 <= scene_index < len(scenes):
        return [int(i) for i in scenes[scene_index].get("nodes", [])]
    children = {c for node in gltf.list("nodes") for c in node.get("children", [])}
    return [i for i in range(len(gltf.list("nodes"))) if i not in children]


def node_tree(gltf: Gltf) -> list[dict[str, Any]]:
    """Nested ``{index, name, mesh?, skin?, children}`` dicts from the scene roots."""
    nodes = gltf.list("nodes")

    def build(index: int, depth: int) -> dict[str, Any]:
        node = nodes[index]
        entry: dict[str, Any] = {"index": index, "name": gltf.node_name(index)}
        if "mesh" in node:
            entry["mesh"] = int(node["mesh"])
        if "skin" in node:
            entry["skin"] = int(node["skin"])
        kids = [int(c) for c in node.get("children", []) if 0 <= int(c) < len(nodes)]
        entry["children"] = [build(c, depth + 1) for c in kids] if depth < 64 else []
        return entry

    return [build(i, 0) for i in root_nodes(gltf) if 0 <= i < len(nodes)]


def _mat_mul(a: Matrix, b: Matrix) -> Matrix:
    out = [0.0] * 16
    for col in range(4):
        for row in range(4):
            out[col * 4 + row] = sum(a[k * 4 + row] * b[col * 4 + k] for k in range(4))
    return out


def local_matrix(node: dict[str, Any]) -> Matrix:
    if isinstance(node.get("matrix"), list) and len(node["matrix"]) == 16:
        return [float(v) for v in node["matrix"]]
    tx, ty, tz = (float(v) for v in node.get("translation", [0.0, 0.0, 0.0]))
    qx, qy, qz, qw = (float(v) for v in node.get("rotation", [0.0, 0.0, 0.0, 1.0]))
    sx, sy, sz = (float(v) for v in node.get("scale", [1.0, 1.0, 1.0]))
    r = [
        1 - 2 * (qy * qy + qz * qz), 2 * (qx * qy + qz * qw), 2 * (qx * qz - qy * qw),
        2 * (qx * qy - qz * qw), 1 - 2 * (qx * qx + qz * qz), 2 * (qy * qz + qx * qw),
        2 * (qx * qz + qy * qw), 2 * (qy * qz - qx * qw), 1 - 2 * (qx * qx + qy * qy),
    ]
    return [
        r[0] * sx, r[1] * sx, r[2] * sx, 0.0,
        r[3] * sy, r[4] * sy, r[5] * sy, 0.0,
        r[6] * sz, r[7] * sz, r[8] * sz, 0.0,
        tx, ty, tz, 1.0,
    ]


def world_matrices(gltf: Gltf) -> dict[int, Matrix]:
    nodes = gltf.list("nodes")
    out: dict[int, Matrix] = {}
    identity = local_matrix({})
    stack = [(i, identity) for i in root_nodes(gltf) if 0 <= i < len(nodes)]
    while stack:
        index, parent = stack.pop()
        if index in out:
            continue
        world = _mat_mul(parent, local_matrix(nodes[index]))
        out[index] = world
        stack.extend((int(c), world) for c in nodes[index].get("children", []) if 0 <= int(c) < len(nodes))
    return out


def _transform_point(m: Matrix, p: tuple[float, float, float]) -> tuple[float, float, float]:
    x, y, z = p
    return (
        m[0] * x + m[4] * y + m[8] * z + m[12],
        m[1] * x + m[5] * y + m[9] * z + m[13],
        m[2] * x + m[6] * y + m[10] * z + m[14],
    )


def scene_extents(gltf: Gltf) -> dict[str, list[float]] | None:
    """World-space AABB of all POSITION accessor bounds.

    Rigid meshes use their node's world transform. Skinned meshes use raw bind-pose
    bounds, because glTF ignores the transform of a skinned mesh node.
    """
    nodes = gltf.list("nodes")
    meshes = gltf.list("meshes")
    worlds = world_matrices(gltf)
    identity = local_matrix({})
    lo = [math.inf] * 3
    hi = [-math.inf] * 3
    for index, node in enumerate(nodes):
        mesh_index = node.get("mesh")
        if mesh_index is None or not 0 <= int(mesh_index) < len(meshes) or index not in worlds:
            continue
        matrix = identity if "skin" in node else worlds[index]
        for primitive in meshes[int(mesh_index)].get("primitives", []):
            position = primitive.get("attributes", {}).get("POSITION")
            bounds = accessor_min_max(gltf, int(position)) if position is not None else None
            if bounds is None:
                continue
            (x0, y0, z0), (x1, y1, z1) = bounds[0][:3], bounds[1][:3]
            for corner in ((x, y, z) for x in (x0, x1) for y in (y0, y1) for z in (z0, z1)):
                point = _transform_point(matrix, corner)
                for axis in range(3):
                    lo[axis] = min(lo[axis], point[axis])
                    hi[axis] = max(hi[axis], point[axis])
    if lo[0] == math.inf:
        return None
    return {"min": lo, "max": hi, "size": [hi[a] - lo[a] for a in range(3)]}


def non_unit_node_scales(gltf: Gltf, tolerance: float = 0.01) -> list[tuple[str, list[float]]]:
    issues: list[tuple[str, list[float]]] = []
    for index, node in enumerate(gltf.list("nodes")):
        scale = node.get("scale")
        if not isinstance(scale, list) or len(scale) != 3:
            continue
        if any(abs(float(value) - 1.0) > tolerance for value in scale):
            issues.append((gltf.node_name(index), [float(value) for value in scale]))
    return issues


# --- skins / animations -------------------------------------------------------


def skin_joint_names(gltf: Gltf, skin_index: int) -> list[str]:
    skins = gltf.list("skins")
    if not 0 <= skin_index < len(skins):
        return []
    nodes = gltf.list("nodes")
    return [gltf.node_name(int(i)) for i in skins[skin_index].get("joints", []) if 0 <= int(i) < len(nodes)]


def skin_joint_name_set(gltf: Gltf) -> set[str]:
    names: set[str] = set()
    for skin_index in range(len(gltf.list("skins"))):
        names.update(skin_joint_names(gltf, skin_index))
    return names


def animation_summaries(gltf: Gltf) -> list[dict[str, Any]]:
    out: list[dict[str, Any]] = []
    for index, anim in enumerate(gltf.list("animations")):
        samplers = anim.get("samplers", [])
        duration = 0.0
        for sampler in samplers:
            bounds = accessor_min_max(gltf, int(sampler.get("input", -1)))
            if bounds is not None and bounds[1]:
                duration = max(duration, bounds[1][0])
        targets: set[tuple[str, str]] = set()
        for channel in anim.get("channels", []):
            target = channel.get("target", {})
            if "node" in target:
                targets.add((gltf.node_name(int(target["node"])), str(target.get("path", ""))))
        out.append({
            "name": str(anim.get("name", f"animation_{index}")),
            "duration_s": round(duration, 4),
            "channel_count": len(anim.get("channels", [])),
            "target_nodes": sorted({node for node, _ in targets}),
            "target_paths": sorted({path for _, path in targets}),
        })
    return out


# --- images -------------------------------------------------------------------


def _png_size(data: bytes) -> tuple[int, int] | None:
    if data[:8] != b"\x89PNG\r\n\x1a\n" or len(data) < 24:
        return None
    return struct.unpack(">II", data[16:24])


def _jpeg_size(data: bytes) -> tuple[int, int] | None:
    if data[:2] != b"\xff\xd8":
        return None
    offset = 2
    while offset + 9 < len(data):
        if data[offset] != 0xFF:
            offset += 1
            continue
        marker = data[offset + 1]
        if marker in (0xD8, 0x01) or 0xD0 <= marker <= 0xD7:
            offset += 2
            continue
        seg_len = struct.unpack(">H", data[offset + 2 : offset + 4])[0]
        if 0xC0 <= marker <= 0xCF and marker not in (0xC4, 0xC8, 0xCC):
            height, width = struct.unpack(">HH", data[offset + 5 : offset + 9])
            return width, height
        offset += 2 + seg_len
    return None


def _image_bytes(gltf: Gltf, image: dict[str, Any]) -> bytes | None:
    if "bufferView" in image:
        views = gltf.list("bufferViews")
        view_index = int(image["bufferView"])
        if not 0 <= view_index < len(views):
            return None
        view = views[view_index]
        buffer_index = int(view.get("buffer", 0))
        buffer = gltf.buffers[buffer_index] if 0 <= buffer_index < len(gltf.buffers) else None
        if buffer is None:
            return None
        start = int(view.get("byteOffset", 0))
        return buffer[start : start + int(view.get("byteLength", 0))]
    if image.get("uri"):
        return _resolve_uri(gltf.path.parent, str(image["uri"]))
    return None


def image_summaries(gltf: Gltf) -> list[dict[str, Any]]:
    out: list[dict[str, Any]] = []
    for index, image in enumerate(gltf.list("images")):
        data = _image_bytes(gltf, image)
        size = (_png_size(data) or _jpeg_size(data)) if data else None
        out.append({
            "name": str(image.get("name") or image.get("uri") or f"image_{index}"),
            "mime": str(image.get("mimeType", "")),
            "width": size[0] if size else None,
            "height": size[1] if size else None,
        })
    return out
