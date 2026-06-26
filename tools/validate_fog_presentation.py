from __future__ import annotations

from pathlib import Path
from typing import Any, Callable


def validate_fog_presentation_ranges(report: Any, load_json: Callable[[Path], dict], assets_dir: Path) -> None:
    """Semantic range guard for fog_presentation tuning values."""
    fog_path = assets_dir / "fog_presentation.v0.json"
    if not fog_path.exists():
        return
    cfg = load_json(fog_path)

    fp = cfg.get("falloff_power", 2.0)
    if not (0.5 <= fp <= 10.0):
        report.fail("fog_presentation range", f"falloff_power={fp} is outside [0.5, 10.0]; values near 0 suppress all visibility")
    else:
        report.ok("fog_presentation falloff_power in range [0.5, 10.0]")

    da = cfg.get("darkness_alpha", 1.0)
    if not (0.0 <= da <= 1.0):
        report.fail("fog_presentation range", f"darkness_alpha={da} is outside [0, 1]")
    else:
        report.ok("fog_presentation darkness_alpha in [0, 1]")

    pl = cfg.get("point_light", {})
    energy = pl.get("energy", 0.0)
    if energy <= 0.0 or energy > 20.0:
        report.fail("fog_presentation range", f"point_light.energy={energy} is outside (0, 20]; zero makes the hero invisible in first-person")
    else:
        report.ok("fog_presentation point_light.energy in (0, 20]")

    rm = pl.get("range_multiplier", 1.0)
    if rm <= 0.0 or rm > 10.0:
        report.fail("fog_presentation range", f"point_light.range_multiplier={rm} is outside (0, 10]; near-zero collapses the light radius")
    else:
        report.ok("fog_presentation point_light.range_multiplier in (0, 10]")

    sc = cfg.get("shadow_cache", {})
    move_eps = sc.get("move_epsilon", 0.006)
    if not (0.0 <= move_eps <= 1.0):
        report.fail("fog_presentation range", f"shadow_cache.move_epsilon={move_eps} is outside [0, 1]")
    else:
        report.ok("fog_presentation shadow_cache.move_epsilon in [0, 1]")

    viewport_eps = sc.get("viewport_size_epsilon_px", 1.0)
    if not (0.0 <= viewport_eps <= 64.0):
        report.fail("fog_presentation range", f"shadow_cache.viewport_size_epsilon_px={viewport_eps} is outside [0, 64]")
    else:
        report.ok("fog_presentation shadow_cache.viewport_size_epsilon_px in [0, 64]")

    throttle = sc.get("performance_min_rebuild_interval_frames", 3)
    if not isinstance(throttle, int) or throttle < 0 or throttle > 60:
        report.fail(
            "fog_presentation range",
            f"shadow_cache.performance_min_rebuild_interval_frames={throttle} is outside [0, 60]",
        )
    else:
        report.ok("fog_presentation shadow_cache.performance_min_rebuild_interval_frames in [0, 60]")


def validate_camera_fog_mode_alignment(report: Any, load_json: Callable[[Path], dict], assets_dir: Path) -> None:
    """Ensure fog organic-edge toggles align with declared camera presentation modes."""
    camera_path = assets_dir / "camera_presentations.v0.json"
    fog_path = assets_dir / "fog_presentation.v0.json"
    if not camera_path.exists() or not fog_path.exists():
        return

    camera_modes = load_json(camera_path).get("modes", {})
    organic = load_json(fog_path).get("organic_edge", {})
    perspective_modes = [
        mode_id
        for mode_id, cfg in camera_modes.items()
        if str(cfg.get("projection", "")).lower() == "perspective"
    ]

    if organic.get("enabled_isometric", True) and "isometric" not in camera_modes:
        report.fail(
            "camera/fog alignment",
            "fog_presentation.organic_edge.enabled_isometric is true but camera_presentations has no isometric mode",
        )
    else:
        report.ok("fog organic_edge isometric toggle matches camera_presentations.isometric")

    if organic.get("enabled_perspective", False) and not perspective_modes:
        report.fail(
            "camera/fog alignment",
            "fog_presentation.organic_edge.enabled_perspective is true but camera_presentations has no perspective mode",
        )
    else:
        report.ok("fog organic_edge perspective toggle matches camera_presentations perspective modes")
