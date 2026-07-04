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
    """Column-major 4x4 inverse-translation matrix for a translation-only transform."""
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
