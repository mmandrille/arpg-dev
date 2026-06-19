#!/usr/bin/env python3
"""Inject the shared biped rig into supplied static monster GLBs."""
from __future__ import annotations

import sys
from pathlib import Path

from tools.assets.rig_hero_glbs import REQUIRED_BONES, parse_glb, rig_glb_bytes
from tools.assets.validate_assets import parse_glb_skin_joint_names

ROOT = Path(__file__).resolve().parents[2]

BIPED_MONSTERS = {
    "dark_purple": (
        "assets/monsters/purple_fantasy/dark_purple_monster.glb",
        "client/assets/monsters/purple_fantasy/dark_purple_monster.glb",
    ),
    "crocodile_archer": (
        "assets/monsters/archer/crocodile_archer.glb",
        "client/assets/monsters/archer/crocodile_archer.glb",
    ),
}


def rig_monster_file(source: Path, target: Path) -> None:
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(rig_glb_bytes(source.read_bytes()))


def validate_target(target: Path) -> None:
    joints = parse_glb_skin_joint_names(target)
    required = set(REQUIRED_BONES)
    if joints is None:
        raise ValueError(f"{target}: missing skin")
    missing = sorted(required - joints)
    if missing:
        raise ValueError(f"{target}: missing joints {missing}")
    parsed = parse_glb(target.read_bytes())
    if "skins" not in parsed.gltf:
        raise ValueError(f"{target}: missing skins")


def main() -> int:
    for monster_id, (source_rel, target_rel) in BIPED_MONSTERS.items():
        source = ROOT / source_rel
        target = ROOT / target_rel
        rig_monster_file(source, target)
        validate_target(target)
        print(f"rigged {monster_id}: {source_rel} -> {target_rel}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
