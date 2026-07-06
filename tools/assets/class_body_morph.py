#!/usr/bin/env python3
"""Fork base_human static mesh into per-class body silhouettes (vertex morph only)."""
from __future__ import annotations

import argparse
import sys
from dataclasses import dataclass
from pathlib import Path

from tools.assets.rig_hero_glbs import (
    ROOT,
    parse_glb,
    read_position_accessor,
    write_glb,
    write_vec3_accessor,
)

SOURCE_REL = "assets/characters/base_human/base_human_mesh.glb"
PLAYER_CLASSES = ("barbarian", "paladin", "ranger", "rogue", "sorcerer")


@dataclass(frozen=True)
class BodyMorphParams:
    shoulder_scale: float = 1.0
    chest_scale: float = 1.0
    chest_depth: float = 1.0
    arm_scale: float = 1.0
    leg_scale: float = 1.0
    torso_length: float = 1.0
    neck_scale: float = 1.0
    forward_shift: float = 0.0


CLASS_MORPHS: dict[str, BodyMorphParams] = {
    "barbarian": BodyMorphParams(
        shoulder_scale=1.20,
        chest_scale=1.14,
        chest_depth=1.10,
        arm_scale=1.12,
        leg_scale=1.10,
        torso_length=1.02,
    ),
    "paladin": BodyMorphParams(
        shoulder_scale=1.12,
        chest_scale=1.10,
        chest_depth=1.08,
        arm_scale=1.06,
        leg_scale=1.05,
    ),
    "ranger": BodyMorphParams(
        shoulder_scale=0.92,
        chest_scale=0.94,
        arm_scale=0.96,
        leg_scale=0.95,
        torso_length=1.04,
    ),
    "sorcerer": BodyMorphParams(
        shoulder_scale=0.90,
        chest_scale=0.88,
        chest_depth=0.92,
        arm_scale=0.90,
        leg_scale=0.92,
        torso_length=1.06,
        neck_scale=1.06,
    ),
    "rogue": BodyMorphParams(
        shoulder_scale=0.88,
        chest_scale=0.90,
        arm_scale=0.90,
        leg_scale=0.88,
        forward_shift=0.028,
    ),
}


def _clamp(value: float, low: float, high: float) -> float:
    return max(low, min(high, value))


def _smoothstep(edge0: float, edge1: float, x: float) -> float:
    if edge0 == edge1:
        return 1.0 if x >= edge1 else 0.0

    t = _clamp((x - edge0) / (edge1 - edge0), 0.0, 1.0)

    return t * t * (3.0 - 2.0 * t)


def _lerp(a: float, b: float, t: float) -> float:
    return a + (b - a) * t


def morph_vertex(
    x: float,
    y: float,
    z: float,
    cx: float,
    mins: list[float],
    maxs: list[float],
    params: BodyMorphParams,
) -> tuple[float, float, float]:
    height = max(maxs[1] - mins[1], 1e-6)
    width = max(maxs[0] - mins[0], 1e-6)
    yn = (y - mins[1]) / height
    side = abs(x - cx) / (width * 0.5 + 1e-6)

    sx = 1.0
    sy = 1.0
    sz = 1.0
    dz = 0.0

    if yn >= 0.68:
        shoulder_t = _smoothstep(0.68, 0.86, yn)
        lateral = _clamp(side / 0.55, 0.0, 1.0)
        sx = _lerp(1.0, params.shoulder_scale, shoulder_t * (0.35 + 0.65 * lateral))
    elif yn >= 0.52:
        if side >= 0.28:
            sx = params.arm_scale
        else:
            sx = params.chest_scale
            sz = params.chest_depth
    elif yn < 0.52 and side >= 0.12:
        sx = params.leg_scale

    if 0.35 < yn < 0.90:
        sy = params.torso_length

    if yn >= 0.84 and side < 0.22:
        sx = max(sx, params.neck_scale)

    if params.forward_shift > 0.0 and yn >= 0.42:
        dz = params.forward_shift * _smoothstep(0.42, 0.82, yn)

    nx = cx + (x - cx) * sx
    ny = mins[1] + (y - mins[1]) * sy
    nz = z * sz + dz

    return nx, ny, nz


def morph_mesh_bytes(data: bytes, class_id: str) -> bytes:
    if class_id not in CLASS_MORPHS:
        raise KeyError(f"unknown class morph: {class_id}")

    params = CLASS_MORPHS[class_id]
    parsed = parse_glb(data)
    gltf = parsed.gltf
    if gltf.get("skins"):
        raise ValueError("source GLB is already skinned")

    bin_buf = bytearray(parsed.bin_blob)
    positions_by_accessor: dict[int, list[tuple[float, float, float]]] = {}
    mins = [float("inf"), float("inf"), float("inf")]
    maxs = [float("-inf"), float("-inf"), float("-inf")]

    for mesh in gltf.get("meshes", []):
        for primitive in mesh.get("primitives", []):
            attrs = primitive.get("attributes", {})
            if "POSITION" not in attrs:
                continue

            position_accessor = int(attrs["POSITION"])
            positions = read_position_accessor(gltf, bytes(bin_buf), position_accessor)
            positions_by_accessor[position_accessor] = positions
            for x, y, z in positions:
                mins[0] = min(mins[0], x)
                mins[1] = min(mins[1], y)
                mins[2] = min(mins[2], z)
                maxs[0] = max(maxs[0], x)
                maxs[1] = max(maxs[1], y)
                maxs[2] = max(maxs[2], z)

    cx = (mins[0] + maxs[0]) * 0.5
    for accessor_index, positions in positions_by_accessor.items():
        morphed = [morph_vertex(x, y, z, cx, mins, maxs, params) for x, y, z in positions]
        write_vec3_accessor(gltf, bin_buf, accessor_index, morphed)

    gltf["buffers"][0]["byteLength"] = len(bin_buf)
    gltf["asset"]["generator"] = "arpg-dev/tools/assets/class_body_morph.py"

    return write_glb(gltf, bytes(bin_buf))


def fork_class_mesh(class_id: str, *, root: Path = ROOT) -> Path:
    source = root / SOURCE_REL
    target = root / f"assets/characters/{class_id}/{class_id}_mesh.glb"
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(morph_mesh_bytes(source.read_bytes(), class_id))

    return target


def generate_all(*, root: Path = ROOT) -> list[Path]:
    return [fork_class_mesh(class_id, root=root) for class_id in PLAYER_CLASSES]


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Generate class body fork meshes from base_human")
    parser.add_argument(
        "command",
        nargs="?",
        choices=("generate",),
        default="generate",
        help="generate all class fork meshes",
    )
    parser.add_argument("--class-id", dest="class_id", help="generate one class only")
    args = parser.parse_args(argv)

    if args.class_id:
        path = fork_class_mesh(args.class_id)
        print(f"wrote {path.relative_to(ROOT)}")
        return 0

    for path in generate_all():
        print(f"wrote {path.relative_to(ROOT)}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
