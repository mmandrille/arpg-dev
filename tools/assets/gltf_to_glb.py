#!/usr/bin/env python3
"""Pack a ``.gltf`` (single external ``.bin`` buffer, external/data-URI images) into one ``.glb``.

ADR-0006 D1 requires single-file runtime GLBs; some CC0 kits ship accessories as
``.gltf`` + ``.bin`` + a shared texture PNG. This packer is deterministic (same
input bytes -> same output bytes), stdlib-only, and keeps every glTF field except
buffer/image URIs, which become BIN-chunk bufferViews.

Usage: python -m tools.assets.gltf_to_glb <in.gltf> <out.glb>
"""
from __future__ import annotations

import json
import struct
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from tools.assets import glb_reader  # noqa: E402


def _pad(data: bytearray, fill: int) -> None:
    while len(data) % 4:
        data.append(fill)


def pack(gltf_path: Path) -> bytes:
    gltf = glb_reader.load_gltf(gltf_path)
    doc = json.loads(json.dumps(gltf.doc))  # deep copy; never mutate the reader's document
    buffers = doc.get("buffers", [])
    if len(buffers) != 1:
        raise ValueError(f"{gltf_path.name}: expected exactly one buffer, found {len(buffers)}")
    if gltf.buffers[0] is None:
        raise ValueError(f"{gltf_path.name}: buffer 0 could not be resolved")
    binary = bytearray(gltf.buffers[0])
    views = doc.setdefault("bufferViews", [])
    for index, image in enumerate(doc.get("images", [])):
        uri = image.pop("uri", None)
        if uri is None:
            continue
        data = glb_reader._resolve_uri(gltf_path.parent, str(uri))
        if data is None:
            raise ValueError(f"{gltf_path.name}: image {index} uri {uri} not found")
        _pad(binary, 0)
        views.append({"buffer": 0, "byteOffset": len(binary), "byteLength": len(data)})
        binary.extend(data)
        image["bufferView"] = len(views) - 1
        if "mimeType" not in image:
            image["mimeType"] = "image/png" if str(uri).lower().endswith(".png") else "image/jpeg"
    _pad(binary, 0)
    doc["buffers"] = [{"byteLength": len(binary)}]
    json_bytes = bytearray(json.dumps(doc, separators=(",", ":"), sort_keys=True).encode("utf-8"))
    _pad(json_bytes, 0x20)
    total = 12 + 8 + len(json_bytes) + 8 + len(binary)
    return b"".join([
        glb_reader.GLB_MAGIC,
        struct.pack("<II", 2, total),
        struct.pack("<II", len(json_bytes), glb_reader.CHUNK_JSON),
        bytes(json_bytes),
        struct.pack("<II", len(binary), glb_reader.CHUNK_BIN),
        bytes(binary),
    ])


def main(argv: list[str] | None = None) -> int:
    args = sys.argv[1:] if argv is None else argv
    if len(args) != 2:
        print("usage: python -m tools.assets.gltf_to_glb <in.gltf> <out.glb>", file=sys.stderr)
        return 2
    source, target = Path(args[0]), Path(args[1])
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(pack(source))
    print(f"[gltf-to-glb] {source} -> {target}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
