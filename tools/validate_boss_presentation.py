"""Cross-checks for shared/assets/boss_presentation.v0.json (v512)."""
from __future__ import annotations

from pathlib import Path
from typing import Any, Callable

MIN_SCALE_RATIO = 1.15
BANNER_MIN_S = 1.0
BANNER_MAX_S = 6.0
LABEL = "boss_presentation"


def validate_boss_presentation(
    report: Any,
    load_json: Callable[[Path], dict],
    shared_dir: Path,
    manifest_path: Path,
    scenes_dir: Path | None = None,
) -> None:
    """Every boss template has presentation; assets, bones, scales, banner/aura ranges are sane."""
    path = shared_dir / "assets" / "boss_presentation.v0.json"
    if not path.exists():
        return
    cfg = load_json(path)
    bosses: dict = cfg.get("bosses", {})
    templates: dict = load_json(shared_dir / "rules" / "boss_templates.v0.json").get("bosses", {})
    kit: dict = load_json(shared_dir / "assets" / "kit_monster_presentation.v0.json").get("monsters", {})
    assets: dict = load_json(manifest_path).get("assets", {})
    failed = False

    def fail(detail: str) -> None:
        nonlocal failed
        failed = True
        report.fail(LABEL, detail)

    for template_id in sorted(templates):
        if template_id not in bosses:
            fail(f"boss template {template_id} has no presentation entry")
    for template_id in sorted(bosses):
        if template_id not in templates:
            fail(f"{template_id}: presentation entry has no boss template")

    for template_id, entry in sorted(bosses.items()):
        key = entry.get("visual_key", "")
        kit_entry = kit.get(key)
        if kit_entry is None:
            fail(f"{template_id}: visual_key {key} missing from kit_monster_presentation")
            continue
        if scenes_dir is not None and not (scenes_dir / f"{key}.tscn").is_file():
            fail(f"{template_id}: scene {key}.tscn missing")
        body = assets.get(kit_entry.get("asset_id", ""), {})
        bones = set(body.get("required_nodes", []))
        mounts = list(entry.get("attachments", []))
        headgear = entry.get("headgear", {})
        if headgear.get("kind") == "asset":
            mounts.append({"asset_id": headgear.get("asset_id", ""), "bone": headgear.get("bone", "")})
        elif headgear.get("kind") == "primitive":
            if headgear.get("shape") not in ("horns", "crown"):
                fail(f"{template_id}: primitive headgear needs shape horns|crown")
            mounts.append({"asset_id": None, "bone": headgear.get("bone", "")})
        for mount in mounts:
            asset_id = mount.get("asset_id")
            if asset_id is not None and assets.get(asset_id, {}).get("type") != "monster":
                fail(f"{template_id}: attachment asset {asset_id} is not a monster manifest asset")
            if mount.get("bone") not in bones:
                fail(f"{template_id}: bone {mount.get('bone')} not in required_nodes of {kit_entry.get('asset_id')}")
        banner = entry.get("banner", {})
        duration = banner.get("duration_s", 0)
        if not (BANNER_MIN_S <= duration <= BANNER_MAX_S):
            fail(f"{template_id}: banner.duration_s={duration} outside [{BANNER_MIN_S}, {BANNER_MAX_S}]")
        if banner.get("fade_in_s", 0) + banner.get("fade_out_s", 0) >= duration:
            fail(f"{template_id}: banner fades must be shorter than duration_s")
        if entry.get("scale", 0) <= 0:
            fail(f"{template_id}: scale must be positive")

    scales = sorted(entry.get("scale", 0) for entry in bosses.values() if entry.get("scale", 0) > 0)
    for low, high in zip(scales, scales[1:]):
        if high / low < MIN_SCALE_RATIO:
            fail(f"boss scales {low} and {high} differ by less than {int((MIN_SCALE_RATIO - 1) * 100)}%")

    if not failed:
        report.ok("boss_presentation covers every boss template with valid assets, bones, scales, and banner timing")
