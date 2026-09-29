#!/usr/bin/env python3
"""Bake a GLB's colour-only materials into one palette-textured material (ADR-0018 P4b).

Some CC0 models (Quaternius) colour each primitive with its own untextured material. The client
tint (``ModelTint``) and hit flash (``ModelReactionController``) work on one ``material_override``
per mesh, which is right for KayKit's single-atlas models but flattens a multi-material mesh to
surface 0's colour. This tool rewrites such a GLB into the KayKit shape:

- every material's ``baseColorFactor`` becomes a ``BLOCK``-pixel square in a small sRGB PNG
- every primitive gets a constant ``TEXCOORD_0`` at the centre of its colour's block, and uses
  material 0 (constant UVs mean zero UV derivatives, so mip level 0 is always sampled and
  colours never bleed)
- one matte material matching the KayKit factors (metallic 0, roughness 0.5)

Stdlib only and deterministic: the same input bytes always give the same output bytes.

Usage: python -m tools.assets.bake_material_palette <in.glb> <out.glb>
"""
from __future__ import annotations

import json
import struct
import sys
import zlib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from tools.assets import glb_reader  # noqa: E402
from tools.assets.gltf_to_glb import _pad, write_glb  # noqa: E402

BLOCK = 4
FLOAT = 5126
ARRAY_BUFFER = 34962
LINEAR = 9729
LINEAR_MIPMAP_LINEAR = 9987
# Matches the KayKit kit materials (e.g. Skeletons 1.0 `skeleton`).
KIT_METALLIC = 0.0
KIT_ROUGHNESS = 0.5


def srgb_byte(linear: float) -> int:
    """glTF baseColorFactor is linear; PNG texels are sRGB."""
    c = min(max(float(linear), 0.0), 1.0)
    encoded = c * 12.92 if c <= 0.0031308 else 1.055 * (c ** (1.0 / 2.4)) - 0.055
    return int(round(encoded * 255.0))


def png_rgb(width: int, height: int, rows: list[bytes]) -> bytes:
    """Minimal RGB8 PNG (filter 0 on every row)."""
    def chunk(tag: bytes, data: bytes) -> bytes:
        return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", zlib.crc32(tag + data) & 0xFFFFFFFF)

    raw = b"".join(b"\x00" + row for row in rows)
    header = struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0)
    return b"".join([
        b"\x89PNG\r\n\x1a\n",
        chunk(b"IHDR", header),
        chunk(b"IDAT", zlib.compress(raw, 9)),
        chunk(b"IEND", b""),
    ])


def palette_colors(materials: list[dict]) -> tuple[list[tuple[int, int, int]], list[int]]:
    """Unique sRGB colours in material order, plus each material's palette slot."""
    colors: list[tuple[int, int, int]] = []
    slot_of_material: list[int] = []
    for material in materials:
        factor = material.get("pbrMetallicRoughness", {}).get("baseColorFactor", [1.0, 1.0, 1.0, 1.0])
        rgb = (srgb_byte(factor[0]), srgb_byte(factor[1]), srgb_byte(factor[2]))
        if rgb not in colors:
            colors.append(rgb)
        slot_of_material.append(colors.index(rgb))
    return colors, slot_of_material


def palette_png(colors: list[tuple[int, int, int]]) -> bytes:
    row = b"".join(bytes(rgb) * BLOCK for rgb in colors)
    return png_rgb(BLOCK * len(colors), BLOCK, [row] * BLOCK)


def slot_uv(slot: int, slot_count: int) -> tuple[float, float]:
    return ((slot + 0.5) / slot_count, 0.5)


def _append_view(doc: dict, binary: bytearray, data: bytes, target: int | None = None) -> int:
    _pad(binary, 0)
    view = {"buffer": 0, "byteOffset": len(binary), "byteLength": len(data)}
    if target is not None:
        view["target"] = target
    binary.extend(data)
    doc.setdefault("bufferViews", []).append(view)
    return len(doc["bufferViews"]) - 1


def bake(glb_path: Path) -> bytes:
    gltf = glb_reader.load_gltf(glb_path)
    doc = json.loads(json.dumps(gltf.doc))  # deep copy; never mutate the reader's document
    if len(doc.get("buffers", [])) != 1 or gltf.buffers[0] is None:
        raise ValueError(f"{glb_path.name}: expected exactly one resolvable buffer")
    if doc.get("textures") or doc.get("images"):
        raise ValueError(f"{glb_path.name}: already textured; palette bake is for colour-only materials")
    materials = doc.get("materials", [])
    if not materials:
        raise ValueError(f"{glb_path.name}: no materials to bake")
    binary = bytearray(gltf.buffers[0])
    colors, slot_of_material = palette_colors(materials)
    accessors = doc.setdefault("accessors", [])
    for mesh in doc.get("meshes", []):
        for primitive in mesh.get("primitives", []):
            slot = slot_of_material[int(primitive.get("material", 0))]
            count = int(accessors[primitive["attributes"]["POSITION"]]["count"])
            u, v = slot_uv(slot, len(colors))
            view = _append_view(doc, binary, struct.pack("<ff", u, v) * count, ARRAY_BUFFER)
            accessors.append({"bufferView": view, "componentType": FLOAT, "count": count, "type": "VEC2",
                              "min": [u, v], "max": [u, v]})
            attributes = primitive["attributes"]
            attributes.pop("TEXCOORD_1", None)
            attributes["TEXCOORD_0"] = len(accessors) - 1
            primitive["material"] = 0
    image_view = _append_view(doc, binary, palette_png(colors))
    doc["images"] = [{"bufferView": image_view, "mimeType": "image/png", "name": "palette"}]
    doc["samplers"] = [{"magFilter": LINEAR, "minFilter": LINEAR_MIPMAP_LINEAR}]
    doc["textures"] = [{"sampler": 0, "source": 0}]
    doc["materials"] = [{
        "name": "palette",
        "pbrMetallicRoughness": {
            "baseColorTexture": {"index": 0},
            "metallicFactor": KIT_METALLIC,
            "roughnessFactor": KIT_ROUGHNESS,
        },
    }]
    return write_glb(doc, binary)


def main(argv: list[str] | None = None) -> int:
    args = sys.argv[1:] if argv is None else argv
    if len(args) != 2:
        print("usage: python -m tools.assets.bake_material_palette <in.glb> <out.glb>", file=sys.stderr)
        return 2
    source, target = Path(args[0]), Path(args[1])
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(bake(source))
    print(f"[bake-palette] {source} -> {target}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
