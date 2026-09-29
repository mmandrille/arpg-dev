#!/usr/bin/env python3
"""Read-only structural report for third-party model kits (ADR-0018 P0).

Usage:
  python -m tools.assets.inspect_kit <file-or-dir> [...] [--out .artifacts/kaykit/report]

Walks ``.glb`` / ``.gltf`` files and writes ``report.json`` + ``report.md`` with
node trees, per-mesh triangle counts, materials, image sizes, skins, animation
clips and world extents. Other model formats (``.fbx``, ``.obj``, ...) are listed
as unsupported so they surface as gaps instead of silently disappearing. The
Markdown report groups skinned files by identical joint-name sets, which answers
"do the characters and the animation pack share one rig?".
"""
from __future__ import annotations

import argparse
import hashlib
import json
import sys
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[2]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from tools.assets import glb_reader  # noqa: E402

SUPPORTED_SUFFIXES = {".glb", ".gltf"}
UNSUPPORTED_MODEL_SUFFIXES = {".fbx", ".obj", ".dae", ".blend", ".3ds", ".stl"}
DEFAULT_OUT = ROOT / ".artifacts" / "kaykit" / "report"


def discover(paths: list[Path]) -> tuple[list[Path], list[Path]]:
    supported: list[Path] = []
    unsupported: list[Path] = []
    for path in paths:
        candidates = sorted(p for p in path.rglob("*") if p.is_file()) if path.is_dir() else [path]
        for candidate in candidates:
            suffix = candidate.suffix.lower()
            if suffix in SUPPORTED_SUFFIXES:
                supported.append(candidate)
            elif suffix in UNSUPPORTED_MODEL_SUFFIXES:
                unsupported.append(candidate)
    return supported, unsupported


def _round_list(values: list[float]) -> list[float]:
    return [round(v, 4) for v in values]


def inspect_file(path: Path, base: Path | None = None) -> dict[str, Any]:
    rel = str(path.relative_to(base)) if base and path.is_relative_to(base) else str(path)
    entry: dict[str, Any] = {"path": rel, "sha256": hashlib.sha256(path.read_bytes()).hexdigest()}
    try:
        gltf = glb_reader.load_gltf(path)
    except (OSError, ValueError) as exc:
        entry["error"] = str(exc)
        return entry
    nodes = gltf.list("nodes")
    materials = gltf.list("materials")
    meshes: list[dict[str, Any]] = []
    for mesh_index, mesh in enumerate(gltf.list("meshes")):
        material_names = sorted({
            str(materials[int(p["material"])].get("name", f"material_{p['material']}"))
            for p in mesh.get("primitives", [])
            if "material" in p and 0 <= int(p["material"]) < len(materials)
        })
        non_triangle = [
            int(p.get("mode", glb_reader.MODE_TRIANGLES))
            for p in mesh.get("primitives", [])
            if glb_reader.primitive_triangle_count(gltf, p) is None
        ]
        meshes.append({
            "name": str(mesh.get("name", f"mesh_{mesh_index}")),
            "node_names": [gltf.node_name(i) for i, n in enumerate(nodes) if n.get("mesh") == mesh_index],
            "skinned": any("skin" in n for n in nodes if n.get("mesh") == mesh_index),
            "primitives": len(mesh.get("primitives", [])),
            "triangles": glb_reader.mesh_triangle_count(gltf, mesh_index),
            "materials": material_names,
            "non_triangle_modes": non_triangle,
        })
    extents = glb_reader.scene_extents(gltf)
    entry.update({
        "node_count": len(nodes),
        "node_tree": glb_reader.node_tree(gltf),
        "meshes": meshes,
        "total_triangles": sum(m["triangles"] for m in meshes),
        "materials": [str(m.get("name", f"material_{i}")) for i, m in enumerate(materials)],
        "images": glb_reader.image_summaries(gltf),
        "skins": [
            {"name": str(s.get("name", f"skin_{i}")), "joints": glb_reader.skin_joint_names(gltf, i)}
            for i, s in enumerate(gltf.list("skins"))
        ],
        "animations": glb_reader.animation_summaries(gltf),
        "extents": {k: _round_list(v) for k, v in extents.items()} if extents else None,
    })
    return entry


def rig_groups(entries: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """Group skinned files by identical joint-name sets (order-insensitive)."""
    groups: dict[tuple[str, ...], list[str]] = {}
    for entry in entries:
        for skin in entry.get("skins", []):
            key = tuple(sorted(skin["joints"]))
            if key:
                groups.setdefault(key, [])
                if entry["path"] not in groups[key]:
                    groups[key].append(entry["path"])
    ordered = sorted(groups.items(), key=lambda item: (-len(item[1]), item[0]))
    return [{"joint_count": len(k), "joints": list(k), "files": sorted(v)} for k, v in ordered]


def build_report(paths: list[Path]) -> dict[str, Any]:
    supported, unsupported = discover(paths)
    base = paths[0] if len(paths) == 1 and paths[0].is_dir() else None
    entries = [inspect_file(p, base) for p in supported]
    return {
        "inputs": [str(p) for p in paths],
        "files": entries,
        "unsupported": [str(p.relative_to(base)) if base else str(p) for p in unsupported],
        "rig_groups": rig_groups(entries),
    }


def render_markdown(report: dict[str, Any]) -> str:
    lines = ["# Kit inspection report", "", f"Inputs: {', '.join(report['inputs'])}", ""]
    lines += ["| File | Tris | Meshes | Skins (joints) | Clips | Images | Size (x,y,z) |", "|---|---|---|---|---|---|---|"]
    for entry in report["files"]:
        if "error" in entry:
            lines.append(f"| `{entry['path']}` | error: {entry['error']} | | | | | |")
            continue
        skins = ", ".join(str(len(s["joints"])) for s in entry["skins"]) or "-"
        images = ", ".join(f"{i['width']}x{i['height']}" for i in entry["images"]) or "-"
        size = ", ".join(f"{v:g}" for v in entry["extents"]["size"]) if entry["extents"] else "-"
        lines.append(
            f"| `{entry['path']}` | {entry['total_triangles']} | {len(entry['meshes'])} | {skins} | "
            f"{len(entry['animations'])} | {images} | {size} |"
        )
    lines += ["", "## Rig comparison", ""]
    groups = report["rig_groups"]
    if not groups:
        lines.append("No skinned files.")
    reference = set(groups[0]["joints"]) if groups else set()
    for index, group in enumerate(groups):
        lines.append(f"### Rig group {index + 1}: {group['joint_count']} joints, {len(group['files'])} file(s)")
        lines.append("")
        lines.append("Files: " + ", ".join(f"`{f}`" for f in group["files"]))
        lines.append("")
        if index == 0:
            lines.append("Joints: " + ", ".join(f"`{j}`" for j in group["joints"]))
        else:
            joints = set(group["joints"])
            lines.append("Missing vs group 1: " + (", ".join(sorted(reference - joints)) or "none"))
            lines.append("Extra vs group 1: " + (", ".join(sorted(joints - reference)) or "none"))
        lines.append("")
    clips = sorted({(e["path"], a["name"], a["duration_s"]) for e in report["files"] for a in e.get("animations", [])})
    lines += ["## Animation clips", ""]
    lines += [f"- `{path}`: **{name}** ({duration:g}s)" for path, name, duration in clips] or ["None."]
    if report["unsupported"]:
        lines += ["", "## Unsupported formats (gaps)", ""]
        lines += [f"- `{p}`" for p in report["unsupported"]]
    return "\n".join(lines) + "\n"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Read-only structural report for third-party model kits.")
    parser.add_argument("paths", nargs="+", type=Path, help="Model files or directories")
    parser.add_argument("--out", type=Path, default=DEFAULT_OUT, help="Report output directory")
    args = parser.parse_args(argv)
    missing = [p for p in args.paths if not p.exists()]
    if missing:
        print(f"[inspect-kit] missing: {', '.join(map(str, missing))}", file=sys.stderr)
        return 2
    report = build_report([p.resolve() for p in args.paths])
    args.out.mkdir(parents=True, exist_ok=True)
    (args.out / "report.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    (args.out / "report.md").write_text(render_markdown(report), encoding="utf-8")
    print(f"[inspect-kit] {len(report['files'])} file(s), {len(report['rig_groups'])} rig group(s), "
          f"{len(report['unsupported'])} unsupported -> {args.out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
