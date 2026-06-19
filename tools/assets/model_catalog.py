#!/usr/bin/env python3
"""Discover previewable character and monster models from repo data."""
from __future__ import annotations

import argparse
import json
import sys
from dataclasses import dataclass
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
MANIFEST_REL = "assets/manifests/assets.v0.json"
CLASS_PRESENTATIONS_REL = "shared/assets/class_presentations.v0.json"
MONSTER_VISUALS_REL = "shared/assets/monster_visuals.v0.json"


@dataclass(frozen=True)
class ModelRow:
    asset_id: str
    asset_type: str
    runtime_path: str
    used_by: tuple[str, ...]
    scene: str = ""
    scale: float = 1.0
    height_offset: float = 0.0


def load_json(path: Path) -> dict:
    with path.open(encoding="utf-8") as fh:
        return json.load(fh)


def load_catalog(root: Path = ROOT) -> list[ModelRow]:
    manifest = load_json(root / MANIFEST_REL)
    class_presentations = load_json(root / CLASS_PRESENTATIONS_REL)
    monster_visuals = load_json(root / MONSTER_VISUALS_REL)
    assets: dict = manifest.get("assets", {})

    usage: dict[str, set[str]] = {}
    metadata: dict[str, dict] = {}

    for class_id, entry in sorted(class_presentations.get("classes", {}).items()):
        model = entry.get("model", {})
        asset_id = str(model.get("asset_id", ""))
        if not asset_id:
            continue
        usage.setdefault(asset_id, set()).add(str(class_id))
        metadata.setdefault(asset_id, {}).update({
            "scale": _positive_float(model.get("scale", 1.0), 1.0),
            "height_offset": float(model.get("height_offset", 0.0)),
        })

    for monster_def_id, entry in sorted(monster_visuals.get("monster_visuals", {}).items()):
        asset_id = str(entry.get("asset_id", ""))
        if not asset_id:
            continue
        usage.setdefault(asset_id, set()).add(str(monster_def_id))
        metadata.setdefault(asset_id, {}).update({
            "scene": str(entry.get("scene", "")),
            "scale": _positive_float(entry.get("scale", 1.0), 1.0),
            "height_offset": float(entry.get("height_offset", 0.0)),
        })

    rows: list[ModelRow] = []
    for asset_id, used_by in usage.items():
        entry = assets.get(asset_id)
        if not isinstance(entry, dict):
            continue
        asset_type = str(entry.get("type", ""))
        if asset_type not in {"character", "monster"}:
            continue
        meta = metadata.get(asset_id, {})
        rows.append(ModelRow(
            asset_id=asset_id,
            asset_type=asset_type,
            runtime_path=str(entry.get("runtime_path", "")),
            used_by=tuple(sorted(used_by)),
            scene=str(meta.get("scene", "")),
            scale=_positive_float(meta.get("scale", 1.0), 1.0),
            height_offset=float(meta.get("height_offset", 0.0)),
        ))
    return sorted(rows, key=lambda row: (row.asset_type, row.asset_id))


def resolve(asset_id: str, root: Path = ROOT) -> ModelRow:
    for row in load_catalog(root):
        if row.asset_id == asset_id:
            return row
    raise KeyError(asset_id)


def format_row(row: ModelRow) -> str:
    return f"{row.asset_id:<32} {row.runtime_path:<62} used_by={','.join(row.used_by)}"


def _positive_float(value, fallback: float) -> float:
    try:
        parsed = float(value)
    except (TypeError, ValueError):
        return fallback
    if parsed <= 0.0:
        return fallback
    return parsed


def _cmd_list(root: Path) -> int:
    for row in load_catalog(root):
        print(format_row(row))
    return 0


def _cmd_resolve(asset_id: str, root: Path, as_json: bool) -> int:
    try:
        row = resolve(asset_id, root)
    except KeyError:
        print(f"unknown model asset_id: {asset_id}; run `make model-list`", file=sys.stderr)
        return 2
    if as_json:
        print(json.dumps({
            "asset_id": row.asset_id,
            "type": row.asset_type,
            "runtime_path": row.runtime_path,
            "used_by": list(row.used_by),
            "scene": row.scene,
            "scale": row.scale,
            "height_offset": row.height_offset,
        }, sort_keys=True))
    else:
        print(format_row(row))
    return 0


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="List and resolve previewable model assets.")
    parser.add_argument("--root", type=Path, default=ROOT)
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("list")
    resolve_parser = sub.add_parser("resolve")
    resolve_parser.add_argument("asset_id")
    resolve_parser.add_argument("--json", action="store_true")
    args = parser.parse_args(argv)

    if args.command == "list":
        return _cmd_list(args.root)
    if args.command == "resolve":
        return _cmd_resolve(args.asset_id, args.root, args.json)
    return 2


if __name__ == "__main__":
    sys.exit(main())
