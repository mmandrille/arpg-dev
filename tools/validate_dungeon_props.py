"""Cross-check: every server-placed dungeon prop id has a client presentation entry (v536)."""
from __future__ import annotations

import json
from pathlib import Path


def check(report, rules_dir: Path, assets_dir: Path) -> None:
    rules = json.loads((rules_dir / "dungeon_generation.v0.json").read_text(encoding="utf-8"))
    presentation = json.loads((assets_dir / "dungeon_kit_presentation.v0.json").read_text(encoding="utf-8"))
    server_ids = {entry["prop_id"] for entry in rules["obstacle_generation"]["props"]["catalog"]}
    client_ids = set(presentation["dressing"]["props"])
    missing = sorted(server_ids - client_ids)
    if missing:
        report.fail("dungeon props: server catalog ids have a client presentation", f"missing {missing}")
    else:
        report.ok("dungeon props: server catalog ids have a client presentation")
    unused = sorted(client_ids - server_ids)
    if unused:
        report.fail("dungeon props: client presentation ids exist in the server catalog", f"unused {unused}")
    else:
        report.ok("dungeon props: client presentation ids exist in the server catalog")
