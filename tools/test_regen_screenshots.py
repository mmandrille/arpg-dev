"""Tests for data-driven showme screenshot regeneration."""
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

from tools.showme.screenshot_catalog import (
    DEFAULT_SUITES,
    DUNGEON_ROOM_FOCUS,
    SCENE_FOCUSES,
    TOWN_PLAY_VIEWS,
    biome_palette_levels,
    class_ids,
    discover_jobs,
    item_asset_ids,
    item_def_ids_with_visuals,
    item_family_ids,
    skill_ids,
    suite_summary,
)

ROOT = Path(__file__).resolve().parents[1]


def test_default_suites_cover_requested_focus_areas() -> None:
    assert "skeleton" in DEFAULT_SUITES
    assert "gear" in DEFAULT_SUITES
    assert "skill-icon" in DEFAULT_SUITES
    assert "item-icon" in DEFAULT_SUITES
    assert "item-icons" in DEFAULT_SUITES
    assert "floor-item" in DEFAULT_SUITES
    assert "item-asset" in DEFAULT_SUITES
    assert "scenes" in DEFAULT_SUITES
    assert "character-screen" in DEFAULT_SUITES
    assert "skill-tree" in DEFAULT_SUITES


def test_discover_jobs_are_data_driven() -> None:
    jobs = discover_jobs(["skeleton", "gear", "skill-icon"])
    class_set = set(class_ids())
    skill_set = set(skill_ids())
    assert class_set
    assert skill_set
    skeleton_classes = {job.slug for job in jobs if job.suite == "skeleton"}
    gear_classes = {job.slug for job in jobs if job.suite == "gear"}
    skill_slugs = {job.slug for job in jobs if job.suite == "skill-icon"}
    assert skeleton_classes == class_set
    assert gear_classes == class_set
    assert skill_slugs == skill_set


def test_skill_tree_suite_covers_points_and_hover_states() -> None:
    jobs = discover_jobs(["skill-tree"])
    assert {job.slug for job in jobs} == {"points", "nopoints", "hover"}
    assert all(job.focus == "skills" for job in jobs)
    assert all("--variant" in job.extra_args for job in jobs)


def test_item_jobs_use_shared_catalog_ids() -> None:
    jobs = discover_jobs(["item-icon", "floor-item", "item-asset"])
    families = set(item_family_ids())
    item_defs = set(item_def_ids_with_visuals())
    assets = set(item_asset_ids())
    assert {job.slug for job in jobs if job.suite == "item-icon"} == families
    assert {job.slug for job in jobs if job.suite == "floor-item"} == item_defs
    assert {job.slug for job in jobs if job.suite == "item-asset"}.issubset(assets)


def test_suite_summary_matches_discovered_counts() -> None:
    for row in suite_summary():
        expected = len(discover_jobs([row["suite"]]))
        assert row["count"] == expected


def _dry_run(*extra: str) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, "-m", "tools.showme.regen_screenshots", *extra, "--dry-run"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )


def test_regen_screenshots_dry_run(tmp_path: Path) -> None:
    out_dir = tmp_path / "shots"
    result = _dry_run("--suite", "skeleton", "--suite", "gear", "--out-dir", str(out_dir))
    assert result.returncode == 0, result.stderr
    assert "--focus skeleton" in result.stdout
    assert "--focus gear" in result.stdout
    assert not out_dir.exists(), "dry-run must not create output directories"


def test_regen_screenshots_dry_run_default_out_dir_writes_nothing() -> None:
    screenshots_root = ROOT / ".artifacts" / "screenshots"
    before = sorted(p.name for p in screenshots_root.iterdir()) if screenshots_root.exists() else []
    result = _dry_run("--suite", "skeleton")
    assert result.returncode == 0, result.stderr
    after = sorted(p.name for p in screenshots_root.iterdir()) if screenshots_root.exists() else []
    assert after == before


def test_makefile_exposes_regen_screenshots_target() -> None:
    makefile = (ROOT / "make" / "client.mk").read_text(encoding="utf-8")
    assert "regen-screenshots:" in makefile
    assert "tools.showme.regen_screenshots" in makefile


def test_scenes_suite_covers_scene_focuses_and_every_biome_palette() -> None:
    jobs = discover_jobs(["scenes"])
    slugs = [job.slug for job in jobs]
    assert len(slugs) == len(set(slugs))
    assert {job.focus for job in jobs if job.focus != DUNGEON_ROOM_FOCUS} == set(SCENE_FOCUSES) | {"town-play"}
    for view, zoom in TOWN_PLAY_VIEWS:
        assert any(job.slug == f"town-play-{view}-{zoom}" and job.extra_args == ("--town-view", view, "--town-zoom", zoom) for job in jobs)
    palettes = biome_palette_levels()
    assert palettes, "dungeon_generation biome_palettes must yield at least one capture level"
    room_jobs = [job for job in jobs if job.focus == DUNGEON_ROOM_FOCUS]
    assert len(room_jobs) == len(palettes)
    for job, (palette_id, level) in zip(room_jobs, palettes):
        assert job.slug == f"{DUNGEON_ROOM_FOCUS}-{palette_id}"
        assert job.extra_args == ("--level", str(level))
        assert level < 0


def test_scenes_dry_run_forwards_dungeon_room_level(tmp_path: Path) -> None:
    result = _dry_run("--suite", "scenes", "--out-dir", str(tmp_path / "shots"))
    assert result.returncode == 0, result.stderr
    assert f"--focus {DUNGEON_ROOM_FOCUS}" in result.stdout
    assert "--level -" in result.stdout


def test_character_screen_suite_covers_every_class_and_variant() -> None:
    jobs = discover_jobs(["character-screen"])
    assert {job.focus for job in jobs} == {"character-screen"}
    by_class: dict[str, set[str]] = {}
    for job in jobs:
        args = list(job.extra_args)
        class_id = args[args.index("--class-id") + 1]
        by_class.setdefault(class_id, set()).add(args[args.index("--variant") + 1])
    assert set(by_class) == set(class_ids())
    for variants in by_class.values():
        assert {"points", "nopoints", "paper-doll"} <= variants
    assert sum("nopoints-dual" in variants for variants in by_class.values()) == 2


def test_item_tooltip_suite_covers_every_rarity() -> None:
    jobs = discover_jobs(["item-tooltip"])
    assert {job.focus for job in jobs} == {"item-tooltip"}
    assert [job.extra_args for job in jobs] == [("--variant", rarity) for rarity in ("common", "magic", "rare", "unique", "set")]
    assert len({job.output_rel for job in jobs}) == len(jobs)
