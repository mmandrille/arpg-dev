#!/usr/bin/env python3
"""Download CC0 equipment GLBs from Poly Pizza and normalize scale for runtime.

Mirrors the hero Tier-3 pipeline: source GLB under assets/, runtime under client/.
"""
from __future__ import annotations

import hashlib
import shutil
import sys
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from tools.assets.rig_hero_glbs import (  # noqa: E402
    _bounds,
    parse_glb,
    read_position_accessor,
    write_glb,
    write_vec3_accessor,
)

POLY_REFERER = "https://poly.pizza/"
IMPORTS: dict[str, dict[str, object]] = {
    # v436 melee
    "assets/equipment/weapons/rusty_sword/rusty_sword.glb": {
        "uuid": "70e87f26-6cb2-4e69-8a76-9620d6cb66f7",
        "origin": "https://poly.pizza/m/cDsobPucFA (Sword by Quaternius)",
        "target_extent": 0.80,
    },
    "assets/equipment/weapons/long_sword/long_sword.glb": {
        "uuid": "65837148-8c3c-42d5-9ce7-c55f9295cc7e",
        "origin": "https://poly.pizza/m/9lLmH8Et4K (Sword by Quaternius)",
        "target_extent": 1.00,
    },
    "assets/equipment/weapons/rapier/rapier.glb": {
        "uuid": "49e2f3d1-9e8d-4322-83d0-9c0dd340ac84",
        "origin": "https://poly.pizza/m/Ds2bJiNI5w (Sword by Quaternius)",
        "target_extent": 0.88,
    },
    "assets/equipment/weapons/starter_axe/starter_axe.glb": {
        "uuid": "2b206f9f-30f7-43a4-8dd8-daa04953962b",
        "origin": "https://poly.pizza/m/W0UYZPYSXf (Axe by Quaternius)",
        "target_extent": 0.96,
    },
    "assets/equipment/armor/shield/kite_shield.glb": {
        "uuid": "21ae2d91-300c-43e0-b140-d119ef00ea9d",
        "origin": "https://poly.pizza/m/xoHSnOjsBG (Shield by Quaternius)",
        "target_extent": 0.48,
    },
    "assets/equipment/armor/shield/tower_shield.glb": {
        "uuid": "21ae2d91-300c-43e0-b140-d119ef00ea9d",
        "origin": "https://poly.pizza/m/xoHSnOjsBG (Shield by Quaternius; tower variant scaled in item_visuals)",
        "target_extent": 0.62,
    },
    # v437 ranged
    "assets/equipment/weapons/training_bow/training_bow.glb": {
        "uuid": "e1abc1c7-47db-48c3-9c65-0879273e8bc8",
        "origin": "https://poly.pizza/m/QnpqjLSKFU (Wooden Bow by Quaternius)",
        "target_extent": 1.05,
    },
    "assets/equipment/weapons/starter_staff/starter_staff.glb": {
        "uuid": "18031dd6-1a26-4fa1-bd09-53c45bd31880",
        "origin": "https://poly.pizza/m/PnGRvO4Lwd (Staff by Quaternius)",
        "target_extent": 1.28,
    },
    # v438 armor
    "assets/equipment/armor/helm/helm.glb": {
        "uuid": "34f2a899-7f26-4161-98f6-734f9f979d8f",
        "origin": "https://poly.pizza/m/apPuLbVJ4N5 (Viking Helmet by Michael Fuchs)",
        "target_extent": 0.21,
    },
    "assets/equipment/armor/chest/chest.glb": {
        "uuid": "60ccfcdb-6aa7-4caf-a688-9b16a2a5f300",
        "origin": "https://poly.pizza/m/TMUoxILh9w (Armor Metal by Quaternius)",
        "target_extent": 0.36,
    },
    "assets/equipment/armor/gloves/gloves.glb": {
        "uuid": "5bcc01d6-1f10-4e50-837f-1b53ba398c60",
        "origin": "https://poly.pizza/m/l1zv4LaA4I (Glove by Quaternius)",
        "target_extent": 0.18,
    },
    "assets/equipment/armor/boots/boots.glb": {
        "uuid": "888317ad-20f0-4b0d-ba01-0bdd017adfd8",
        "origin": "https://poly.pizza/m/7HbqG8RwRcA (Boots by Poly by Google)",
        "target_extent": 0.22,
    },
    "assets/equipment/armor/amulet/amulet.glb": {
        "uuid": "2afce515-0123-40df-9740-23e5b97725e4",
        "origin": "https://poly.pizza/m/Jvhs8DCNDZ (Necklace by Quaternius)",
        "target_extent": 0.14,
    },
    "assets/equipment/armor/ring/ring.glb": {
        "uuid": "c68bce85-929c-442b-b388-ab9d5fd6327c",
        "origin": "https://poly.pizza/m/f9AVG7oFxBi (Ring by Poly by Google)",
        "target_extent": 0.07,
    },
}


def _download_poly(uuid: str) -> bytes:
    url = f"https://static.poly.pizza/{uuid}.glb"
    req = urllib.request.Request(url, headers={"Referer": POLY_REFERER, "User-Agent": "arpg-asset-import/1.0"})
    with urllib.request.urlopen(req, timeout=120) as resp:
        return resp.read()


def _positions_by_accessor(gltf: dict, bin_blob: bytes) -> dict[int, list[tuple[float, float, float]]]:
    out: dict[int, list[tuple[float, float, float]]] = {}
    for mesh in gltf.get("meshes", []):
        for primitive in mesh.get("primitives", []):
            pos_idx = primitive.get("attributes", {}).get("POSITION")
            if pos_idx is None:
                continue
            out[int(pos_idx)] = read_position_accessor(gltf, bin_blob, int(pos_idx))
    return out


def _normalize_extent(
    gltf: dict,
    bin_buf: bytearray,
    positions_by_accessor: dict[int, list[tuple[float, float, float]]],
    target_extent: float,
) -> None:
    mins, maxs = _bounds(positions_by_accessor)
    extent = max(maxs[i] - mins[i] for i in range(3))
    if extent <= 0.001:
        return
    scale = target_extent / extent
    origin = tuple((mins[i] + maxs[i]) * 0.5 for i in range(3))
    for accessor_index, positions in positions_by_accessor.items():
        scaled = []
        for x, y, z in positions:
            scaled.append((
                origin[0] + (x - origin[0]) * scale,
                origin[1] + (y - origin[1]) * scale,
                origin[2] + (z - origin[2]) * scale,
            ))
        write_vec3_accessor(gltf, bin_buf, accessor_index, scaled)
        positions_by_accessor[accessor_index] = scaled


def _flatten_node_scales(gltf: dict) -> None:
    """Poly Pizza exports often keep cm mesh data under node scale=100; vertices are
    normalized to meters in accessors, so non-identity node scales double-apply size."""
    for node in gltf.get("nodes", []):
        if "scale" in node:
            del node["scale"]


def import_one(source_rel: str, spec: dict[str, object]) -> tuple[Path, str]:
    source_path = ROOT / str(source_rel)
    runtime_rel = str(source_rel).replace("assets/", "client/assets/", 1)
    runtime_path = ROOT / runtime_rel
    legacy_path = source_path.with_name(f"{source_path.stem}_legacy.glb")

    raw = _download_poly(str(spec["uuid"]))
    chunked = parse_glb(raw)
    bin_buf = bytearray(chunked.bin_blob)
    positions = _positions_by_accessor(chunked.gltf, bytes(bin_buf))
    _normalize_extent(chunked.gltf, bin_buf, positions, float(spec["target_extent"]))
    _flatten_node_scales(chunked.gltf)
    normalized = write_glb(chunked.gltf, bytes(bin_buf))

    source_path.parent.mkdir(parents=True, exist_ok=True)
    runtime_path.parent.mkdir(parents=True, exist_ok=True)
    if source_path.exists() and not legacy_path.exists():
        shutil.copy2(source_path, legacy_path)

    source_path.write_bytes(normalized)
    runtime_path.write_bytes(normalized)
    digest = hashlib.sha256(normalized).hexdigest()
    print(f"imported {source_rel} -> {runtime_rel} sha256={digest}")
    return runtime_path, digest


def main(argv: list[str] | None = None) -> int:
    keys = list(IMPORTS.keys())
    if argv and len(argv) > 1:
        keys = [k for k in argv[1:] if k in IMPORTS]
        if not keys:
            print("unknown keys; known:", ", ".join(IMPORTS))
            return 1
    for key in keys:
        import_one(key, IMPORTS[key])
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
